package input

import (
	"fmt"
	"strconv"
	"strings"

	"ghgo/internal/domain"
	"ghgo/internal/vocab"
)

func validateCommitRequest(c CommitContext, parsed ParseResult) error {
	if err := validateCommitContext(c); err != nil {
		return err
	}
	if parsed.InputKind != string(c.InputKind) {
		return fmt.Errorf("parsed input kind %q does not match context input kind %q", parsed.InputKind, c.InputKind)
	}
	if parsed.RowsError > 0 || parsedRowsContainErrors(parsed.Rows) {
		return fmt.Errorf("parsed rows contain blocking errors")
	}
	if validParsedRowCount(parsed.Rows) == 0 {
		return fmt.Errorf("no valid rows")
	}
	return validateParsedRows(c, parsed.Rows)
}

func validateCommitContext(c CommitContext) error {
	if strings.TrimSpace(c.OrganizationID) == "" {
		return fmt.Errorf("missing organization ID")
	}
	if strings.TrimSpace(c.ReportingPeriodID) == "" {
		return fmt.Errorf("missing reporting period ID")
	}
	if c.ReportingYear == 0 {
		return fmt.Errorf("missing reporting year")
	}
	if c.PeriodStart.IsZero() {
		return fmt.Errorf("period start is required")
	}
	if c.PeriodEnd.IsZero() {
		return fmt.Errorf("period end is required")
	}
	if c.PeriodStart.After(c.PeriodEnd) {
		return fmt.Errorf("period start is after period end")
	}
	if !c.InputKind.Valid() {
		return fmt.Errorf("invalid input kind %q", c.InputKind)
	}

	switch c.InputKind {
	case vocab.InputElectricityMonthlyKWh:
		if err := requireFacility(c); err != nil {
			return err
		}
		return validateElectricityGO(c)
	case vocab.InputNaturalGasMonthlySmc, vocab.InputRefrigerantsAnnualKg:
		return requireFacility(c)
	case vocab.InputMobileFuelLitres:
		if err := validateOptionalFacility(c); err != nil {
			return err
		}
		if c.MobileMethod == "" {
			return fmt.Errorf("mobile method is required")
		}
		if c.MobileMethod != domain.MobileMethodFuelBased {
			return fmt.Errorf("mobile_fuel_litres requires fuel_based mobile method")
		}
	case vocab.InputVehicleDistanceKm:
		if err := validateOptionalFacility(c); err != nil {
			return err
		}
		if c.MobileMethod == "" {
			return fmt.Errorf("mobile method is required")
		}
		if c.MobileMethod != domain.MobileMethodDistanceBased {
			return fmt.Errorf("vehicle_distance_km requires distance_based mobile method")
		}
	}

	return nil
}

func requireFacility(c CommitContext) error {
	if c.FacilityID == nil || strings.TrimSpace(*c.FacilityID) == "" {
		return fmt.Errorf("facility ID is required for %s", c.InputKind)
	}
	return nil
}

func validateOptionalFacility(c CommitContext) error {
	if c.FacilityID != nil && strings.TrimSpace(*c.FacilityID) == "" {
		return fmt.Errorf("facility ID cannot be empty")
	}
	return nil
}

func validateElectricityGO(c CommitContext) error {
	if c.HasGuaranteesOfOrigin {
		if c.GOCoverage != domain.GOCoverageFull {
			return fmt.Errorf("guarantees of origin require full GO coverage")
		}
		return nil
	}
	if c.GOCoverage != domain.GOCoverageNone {
		return fmt.Errorf("missing guarantees of origin require GO coverage none")
	}
	return nil
}

func parsedRowsContainErrors(rows []ParsedRow) bool {
	for _, row := range rows {
		if len(row.Errors) > 0 {
			return true
		}
	}
	return false
}

func validParsedRowCount(rows []ParsedRow) int {
	count := 0
	for _, row := range rows {
		if len(row.Errors) == 0 {
			count++
		}
	}
	return count
}

func validateParsedRows(c CommitContext, rows []ParsedRow) error {
	switch c.InputKind {
	case vocab.InputElectricityMonthlyKWh:
		return validateMonthlyRows(rows, vocab.UnitKWh)
	case vocab.InputNaturalGasMonthlySmc:
		return validateMonthlyRows(rows, vocab.UnitSmc)
	case vocab.InputMobileFuelLitres:
		return validateMobileFuelRows(rows)
	case vocab.InputVehicleDistanceKm:
		return validateVehicleDistanceRows(rows)
	case vocab.InputRefrigerantsAnnualKg:
		return validateRefrigerantRows(rows)
	}
	return nil
}

func validateMonthlyRows(rows []ParsedRow, unit vocab.Unit) error {
	seenMonths := make(map[int]bool)
	for _, row := range rows {
		month, err := normalizedMonthNumber(row)
		if err != nil {
			return err
		}
		if seenMonths[month] {
			return fmt.Errorf("row %d duplicates month %d", row.RowNumber, month)
		}
		seenMonths[month] = true

		if _, err := normalizedAmount(row); err != nil {
			return err
		}
		if err := requireNormalizedUnit(row, unit); err != nil {
			return err
		}
	}
	return nil
}

func validateMobileFuelRows(rows []ParsedRow) error {
	for _, row := range rows {
		if _, err := requireNormalizedField(row, "fuel_type"); err != nil {
			return err
		}
		if _, err := normalizedAmount(row); err != nil {
			return err
		}
		if err := requireNormalizedUnit(row, vocab.UnitLitre); err != nil {
			return err
		}
	}
	return nil
}

func validateVehicleDistanceRows(rows []ParsedRow) error {
	for _, row := range rows {
		for _, field := range []string{"vehicle_name", "plate", "vehicle_type", "vehicle_size_class", "fuel_type"} {
			if _, err := requireNormalizedField(row, field); err != nil {
				return err
			}
		}
		if _, err := normalizedAmount(row); err != nil {
			return err
		}
		if err := requireNormalizedUnit(row, vocab.UnitKm); err != nil {
			return err
		}
	}
	return nil
}

func validateRefrigerantRows(rows []ParsedRow) error {
	for _, row := range rows {
		if _, err := requireNormalizedField(row, "substance"); err != nil {
			return err
		}
		if _, err := normalizedAmount(row); err != nil {
			return err
		}
		if err := requireNormalizedUnit(row, vocab.UnitKg); err != nil {
			return err
		}
	}
	return nil
}

func normalizedMonthNumber(row ParsedRow) (int, error) {
	value, err := requireNormalizedField(row, "month_number")
	if err != nil {
		return 0, err
	}
	month, err := strconv.Atoi(value)
	if err != nil || month < 1 || month > 12 {
		return 0, fmt.Errorf("row %d has invalid month_number %q", row.RowNumber, value)
	}
	return month, nil
}

func normalizedAmount(row ParsedRow) (float64, error) {
	value, err := requireNormalizedField(row, "amount")
	if err != nil {
		return 0, err
	}
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("row %d has invalid amount %q", row.RowNumber, value)
	}
	if amount < 0 {
		return 0, fmt.Errorf("row %d has negative amount", row.RowNumber)
	}
	return amount, nil
}

func requireNormalizedUnit(row ParsedRow, want vocab.Unit) error {
	got, err := requireNormalizedField(row, "unit")
	if err != nil {
		return err
	}
	if got != string(want) {
		return fmt.Errorf("row %d has unit %q, want %q", row.RowNumber, got, want)
	}
	return nil
}

func requireNormalizedField(row ParsedRow, field string) (string, error) {
	value := row.Normalized[field]
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("row %d missing normalized field %q", row.RowNumber, field)
	}
	return value, nil
}
