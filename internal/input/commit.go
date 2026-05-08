package input

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/ports"
	"ghgo/internal/vocab"
)

func CommitParsedInput(ctx context.Context, st ports.Store, c CommitContext, parsed ParseResult) (CommitResult, error) {
	result := CommitResult{
		RowsTotal: parsed.RowsTotal,
		RowsValid: parsed.RowsValid,
		RowsError: parsed.RowsError,
	}
	if st == nil {
		return result, fmt.Errorf("store is required")
	}
	if err := validateCommitRequest(c, parsed); err != nil {
		return result, fmt.Errorf("%w: %v", ErrInvalidCommit, err)
	}

	txResult := result
	if err := st.WithTx(ctx, func(tx ports.Store) error {
		if err := validateStoredSettings(ctx, tx, c); err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := supersedeExistingActiveData(ctx, tx, c, parsed.Rows, now); err != nil {
			return err
		}

		batchID, err := newID("paste_batch")
		if err != nil {
			return err
		}
		contextJSON, err := marshalContext(c)
		if err != nil {
			return err
		}

		batch := domain.PasteBatch{
			ID:                batchID,
			OrganizationID:    domain.ID(c.OrganizationID),
			ReportingPeriodID: domain.ID(c.ReportingPeriodID),
			InputKind:         string(c.InputKind),
			ContextJSON:       contextJSON,
			RawText:           parsed.RawText,
			RawHash:           RawHash(parsed.RawText),
			Status:            domain.PasteBatchStatusParsed,
			RowsTotal:         parsed.RowsTotal,
			RowsValid:         parsed.RowsValid,
			RowsError:         parsed.RowsError,
			CreatedAt:         now,
		}
		if err := tx.CreatePasteBatch(ctx, batch); err != nil {
			return err
		}

		rowActivityIDs, activityRecordIDs, err := createActivityRecords(ctx, tx, c, parsed.Rows, now)
		if err != nil {
			return err
		}
		if err := createPasteRows(ctx, tx, batch.ID, parsed.Rows, rowActivityIDs); err != nil {
			return err
		}
		if err := tx.MarkPasteBatchCommitted(ctx, batch.ID, now); err != nil {
			return err
		}

		txResult.PasteBatchID = batch.ID
		txResult.ActivityRecordIDs = activityRecordIDs
		txResult.Committed = true
		return nil
	}); err != nil {
		return result, err
	}

	return txResult, nil
}

func validateStoredSettings(ctx context.Context, st ports.Store, c CommitContext) error {
	expected, ok := expectedMobileMethod(c.InputKind)
	if !ok {
		return nil
	}

	settings, err := st.GetReportingPeriodSettings(ctx, domain.ID(c.ReportingPeriodID))
	if errors.Is(err, ports.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if settings.MobileMethod != expected {
		return fmt.Errorf("reporting period is set to %s, so %s data cannot be saved", mobileMethodDisplayLabel(settings.MobileMethod), inputKindDisplayLabel(c.InputKind))
	}
	return nil
}

func expectedMobileMethod(inputKind vocab.InputKind) (domain.MobileMethod, bool) {
	switch inputKind {
	case vocab.InputMobileFuelLitres:
		return domain.MobileMethodFuelBased, true
	case vocab.InputVehicleDistanceKm:
		return domain.MobileMethodDistanceBased, true
	}
	return "", false
}

func supersedeExistingActiveData(ctx context.Context, st ports.Store, c CommitContext, rows []ParsedRow, updatedAt time.Time) error {
	sourceKind := sourceKindForInput(c.InputKind)
	switch c.InputKind {
	case vocab.InputElectricityMonthlyKWh, vocab.InputNaturalGasMonthlySmc:
		facilityID := domain.ID(*c.FacilityID)
		for _, row := range rows {
			month, err := normalizedMonthNumber(row)
			if err != nil {
				return err
			}
			periodStart, _ := monthBounds(c.ReportingYear, month)
			if err := st.SupersedeActiveActivityRecordsByMonthlyKey(ctx, domain.ID(c.ReportingPeriodID), facilityID, sourceKind, periodStart, updatedAt); err != nil {
				return err
			}
		}
	case vocab.InputMobileFuelLitres, vocab.InputVehicleDistanceKm, vocab.InputRefrigerantsAnnualKg:
		facilityID := domainIDPtr(c.FacilityID)
		if err := st.SupersedeActiveActivityRecordsByPeriodFacilitySource(ctx, domain.ID(c.ReportingPeriodID), facilityID, sourceKind, updatedAt); err != nil {
			return err
		}
	}
	return nil
}

func inputKindDisplayLabel(inputKind vocab.InputKind) string {
	switch inputKind {
	case vocab.InputElectricityMonthlyKWh:
		return "electricity"
	case vocab.InputNaturalGasMonthlySmc:
		return "natural gas"
	case vocab.InputMobileFuelLitres:
		return "mobile fuel"
	case vocab.InputVehicleDistanceKm:
		return "vehicle distance"
	case vocab.InputRefrigerantsAnnualKg:
		return "refrigerants"
	}
	return "input"
}

func mobileMethodDisplayLabel(method domain.MobileMethod) string {
	switch method {
	case domain.MobileMethodFuelBased:
		return "Fuel consumed"
	case domain.MobileMethodDistanceBased:
		return "Distance travelled"
	}
	return "the selected method"
}

func createActivityRecords(ctx context.Context, st ports.Store, c CommitContext, rows []ParsedRow, now time.Time) ([]*domain.ID, []string, error) {
	switch c.InputKind {
	case vocab.InputElectricityMonthlyKWh, vocab.InputNaturalGasMonthlySmc:
		return createMonthlyActivityRecords(ctx, st, c, rows, now)
	case vocab.InputMobileFuelLitres:
		return createMobileFuelActivityRecords(ctx, st, c, rows, now)
	case vocab.InputVehicleDistanceKm:
		return createVehicleDistanceActivityRecords(ctx, st, c, rows, now)
	case vocab.InputRefrigerantsAnnualKg:
		return createRefrigerantActivityRecords(ctx, st, c, rows, now)
	}
	return nil, nil, fmt.Errorf("unsupported input kind %q", c.InputKind)
}

func createMonthlyActivityRecords(ctx context.Context, st ports.Store, c CommitContext, rows []ParsedRow, now time.Time) ([]*domain.ID, []string, error) {
	rowActivityIDs := make([]*domain.ID, len(rows))
	activityRecordIDs := make([]string, 0, len(rows))
	sourceKind := sourceKindForInput(c.InputKind)
	facilityID := domainIDPtr(c.FacilityID)

	for i, row := range rows {
		month, err := normalizedMonthNumber(row)
		if err != nil {
			return nil, nil, err
		}
		amount, err := normalizedAmount(row)
		if err != nil {
			return nil, nil, err
		}
		periodStart, periodEnd := monthBounds(c.ReportingYear, month)

		recordID, err := newID("activity_record")
		if err != nil {
			return nil, nil, err
		}
		record := domain.ActivityRecord{
			ID:                recordID,
			OrganizationID:    domain.ID(c.OrganizationID),
			FacilityID:        facilityID,
			ReportingPeriodID: domain.ID(c.ReportingPeriodID),
			SourceKind:        sourceKind,
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			Amount:            amount,
			Status:            domain.ActivityRecordStatusActive,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if c.InputKind == vocab.InputElectricityMonthlyKWh {
			record.Scope = domain.Scope2
			record.Vector = domain.ActivityVectorElectricity
			record.Category = "purchased_electricity"
			record.Method = domain.ActivityMethodLocationBased
			record.ActivityType = "purchased_electricity"
			record.Unit = string(vocab.UnitKWh)
		} else {
			record.Scope = domain.Scope1
			record.Vector = domain.ActivityVectorNaturalGas
			record.Category = "stationary_combustion"
			record.Method = domain.ActivityMethodFuelBased
			record.ActivityType = "natural_gas"
			record.Unit = string(vocab.UnitSmc)
		}
		record.SourceHash = SourceHash(record)

		if err := st.CreateActivityRecord(ctx, record); err != nil {
			return nil, nil, err
		}
		rowActivityIDs[i] = idPtr(record.ID)
		activityRecordIDs = append(activityRecordIDs, record.ID)
	}

	return rowActivityIDs, activityRecordIDs, nil
}

type aggregateRows struct {
	amount float64
	rows   []int
}

func createMobileFuelActivityRecords(ctx context.Context, st ports.Store, c CommitContext, rows []ParsedRow, now time.Time) ([]*domain.ID, []string, error) {
	rowActivityIDs := make([]*domain.ID, len(rows))
	aggregates := make(map[string]*aggregateRows)
	var order []string

	for i, row := range rows {
		fuelType := row.Normalized["fuel_type"]
		amount, err := normalizedAmount(row)
		if err != nil {
			return nil, nil, err
		}
		if aggregates[fuelType] == nil {
			aggregates[fuelType] = &aggregateRows{}
			order = append(order, fuelType)
		}
		aggregates[fuelType].amount += amount
		aggregates[fuelType].rows = append(aggregates[fuelType].rows, i)
	}

	activityRecordIDs := make([]string, 0, len(order))
	for _, fuelType := range order {
		aggregate := aggregates[fuelType]
		recordID, err := newID("activity_record")
		if err != nil {
			return nil, nil, err
		}
		record := domain.ActivityRecord{
			ID:                recordID,
			OrganizationID:    domain.ID(c.OrganizationID),
			FacilityID:        domainIDPtr(c.FacilityID),
			ReportingPeriodID: domain.ID(c.ReportingPeriodID),
			SourceKind:        domain.ActivitySourceKindMobileFuelLitres,
			Scope:             domain.Scope1,
			Vector:            domain.ActivityVectorMobileCombustion,
			Category:          "mobile_combustion",
			Method:            domain.ActivityMethodFuelBased,
			ActivityType:      fuelType + "_mobile",
			PeriodStart:       c.PeriodStart,
			PeriodEnd:         c.PeriodEnd,
			Amount:            aggregate.amount,
			Unit:              string(vocab.UnitLitre),
			FuelType:          fuelType,
			Status:            domain.ActivityRecordStatusActive,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		record.SourceHash = SourceHash(record)

		if err := st.CreateActivityRecord(ctx, record); err != nil {
			return nil, nil, err
		}
		for _, rowIndex := range aggregate.rows {
			rowActivityIDs[rowIndex] = idPtr(record.ID)
		}
		activityRecordIDs = append(activityRecordIDs, record.ID)
	}

	return rowActivityIDs, activityRecordIDs, nil
}

func createVehicleDistanceActivityRecords(ctx context.Context, st ports.Store, c CommitContext, rows []ParsedRow, now time.Time) ([]*domain.ID, []string, error) {
	rowActivityIDs := make([]*domain.ID, len(rows))
	activityRecordIDs := make([]string, 0, len(rows))

	for i, row := range rows {
		amount, err := normalizedAmount(row)
		if err != nil {
			return nil, nil, err
		}
		recordID, err := newID("activity_record")
		if err != nil {
			return nil, nil, err
		}
		record := domain.ActivityRecord{
			ID:                recordID,
			OrganizationID:    domain.ID(c.OrganizationID),
			FacilityID:        domainIDPtr(c.FacilityID),
			ReportingPeriodID: domain.ID(c.ReportingPeriodID),
			SourceKind:        domain.ActivitySourceKindVehicleDistanceKM,
			Scope:             domain.Scope1,
			Vector:            domain.ActivityVectorMobileCombustion,
			Category:          "mobile_combustion",
			Method:            domain.ActivityMethodDistanceBased,
			ActivityType:      "vehicle_distance",
			PeriodStart:       c.PeriodStart,
			PeriodEnd:         c.PeriodEnd,
			Amount:            amount,
			Unit:              string(vocab.UnitKm),
			FuelType:          row.Normalized["fuel_type"],
			VehicleName:       row.Normalized["vehicle_name"],
			Plate:             row.Normalized["plate"],
			VehicleType:       row.Normalized["vehicle_type"],
			VehicleSizeClass:  row.Normalized["vehicle_size_class"],
			Status:            domain.ActivityRecordStatusActive,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		record.SourceHash = SourceHash(record)

		if err := st.CreateActivityRecord(ctx, record); err != nil {
			return nil, nil, err
		}
		rowActivityIDs[i] = idPtr(record.ID)
		activityRecordIDs = append(activityRecordIDs, record.ID)
	}

	return rowActivityIDs, activityRecordIDs, nil
}

func createRefrigerantActivityRecords(ctx context.Context, st ports.Store, c CommitContext, rows []ParsedRow, now time.Time) ([]*domain.ID, []string, error) {
	rowActivityIDs := make([]*domain.ID, len(rows))
	aggregates := make(map[string]*aggregateRows)
	var order []string

	for i, row := range rows {
		substance := row.Normalized["substance"]
		amount, err := normalizedAmount(row)
		if err != nil {
			return nil, nil, err
		}
		if aggregates[substance] == nil {
			aggregates[substance] = &aggregateRows{}
			order = append(order, substance)
		}
		aggregates[substance].amount += amount
		aggregates[substance].rows = append(aggregates[substance].rows, i)
	}

	activityRecordIDs := make([]string, 0, len(order))
	for _, substance := range order {
		aggregate := aggregates[substance]
		recordID, err := newID("activity_record")
		if err != nil {
			return nil, nil, err
		}
		record := domain.ActivityRecord{
			ID:                recordID,
			OrganizationID:    domain.ID(c.OrganizationID),
			FacilityID:        domainIDPtr(c.FacilityID),
			ReportingPeriodID: domain.ID(c.ReportingPeriodID),
			SourceKind:        domain.ActivitySourceKindRefrigerantsAnnualKG,
			Scope:             domain.Scope1,
			Vector:            domain.ActivityVectorRefrigerants,
			Category:          "fugitive_emissions",
			Method:            domain.ActivityMethodDirectGWP,
			ActivityType:      "refrigerant_leakage",
			PeriodStart:       c.PeriodStart,
			PeriodEnd:         c.PeriodEnd,
			Amount:            aggregate.amount,
			Unit:              string(vocab.UnitKg),
			Substance:         substance,
			Status:            domain.ActivityRecordStatusActive,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		record.SourceHash = SourceHash(record)

		if err := st.CreateActivityRecord(ctx, record); err != nil {
			return nil, nil, err
		}
		for _, rowIndex := range aggregate.rows {
			rowActivityIDs[rowIndex] = idPtr(record.ID)
		}
		activityRecordIDs = append(activityRecordIDs, record.ID)
	}

	return rowActivityIDs, activityRecordIDs, nil
}

func createPasteRows(ctx context.Context, st ports.Store, pasteBatchID domain.ID, rows []ParsedRow, rowActivityIDs []*domain.ID) error {
	for i, parsedRow := range rows {
		pasteRowID, err := newID("paste_row")
		if err != nil {
			return err
		}
		rawJSON, err := marshalRawFields(parsedRow.RawFields)
		if err != nil {
			return err
		}
		normalizedJSON, err := marshalNormalizedFields(parsedRow.Normalized)
		if err != nil {
			return err
		}
		errorsJSON, err := marshalIssues(parsedRow.Errors)
		if err != nil {
			return err
		}
		warningsJSON, err := marshalIssues(parsedRow.Warnings)
		if err != nil {
			return err
		}

		row := domain.PasteRow{
			ID:               pasteRowID,
			PasteBatchID:     pasteBatchID,
			RowNumber:        parsedRow.RowNumber,
			RawJSON:          rawJSON,
			NormalizedJSON:   normalizedJSON,
			Status:           domain.PasteRowStatusCommitted,
			ErrorsJSON:       errorsJSON,
			WarningsJSON:     warningsJSON,
			ActivityRecordID: rowActivityIDs[i],
		}
		if err := st.CreatePasteRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func sourceKindForInput(inputKind vocab.InputKind) domain.ActivitySourceKind {
	switch inputKind {
	case vocab.InputElectricityMonthlyKWh:
		return domain.ActivitySourceKindElectricityMonthlyKWh
	case vocab.InputNaturalGasMonthlySmc:
		return domain.ActivitySourceKindNaturalGasMonthlySMC
	case vocab.InputMobileFuelLitres:
		return domain.ActivitySourceKindMobileFuelLitres
	case vocab.InputVehicleDistanceKm:
		return domain.ActivitySourceKindVehicleDistanceKM
	case vocab.InputRefrigerantsAnnualKg:
		return domain.ActivitySourceKindRefrigerantsAnnualKG
	}
	return ""
}

func monthBounds(year int, month int) (time.Time, time.Time) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	return start, end
}

func marshalContext(c CommitContext) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("marshal commit context: %w", err)
	}
	return string(data), nil
}

func marshalRawFields(fields []string) (string, error) {
	if fields == nil {
		fields = []string{}
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return "", fmt.Errorf("marshal raw fields: %w", err)
	}
	return string(data), nil
}

func marshalNormalizedFields(fields map[string]string) (string, error) {
	if fields == nil {
		fields = map[string]string{}
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return "", fmt.Errorf("marshal normalized fields: %w", err)
	}
	return string(data), nil
}

func marshalIssues(issues []ParseIssue) (string, error) {
	if issues == nil {
		issues = []ParseIssue{}
	}
	data, err := json.Marshal(issues)
	if err != nil {
		return "", fmt.Errorf("marshal issues: %w", err)
	}
	return string(data), nil
}

func domainIDPtr(value *string) *domain.ID {
	if value == nil {
		return nil
	}
	id := domain.ID(*value)
	return &id
}

func idPtr(id domain.ID) *domain.ID {
	value := id
	return &value
}

func newID(prefix string) (domain.ID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate ID: %w", err)
	}
	return domain.ID(prefix + "_" + hex.EncodeToString(b[:])), nil
}
