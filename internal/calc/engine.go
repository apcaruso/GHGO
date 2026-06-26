package calc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/store"
)

type Engine struct {
	Store  *store.Store
	Lookup *factors.Lookup
}

type RunOptions struct {
	OrganizationID    string
	ReportingPeriodID string
	FactorSetID       string
}

type RunResult struct {
	CalculationRunID         string
	ResultsCreated           int
	PrimaryTotalKgCO2e       float64
	LocationBasedTotalKgCO2e *float64
}

func NewEngine(st *store.Store, lookup *factors.Lookup) *Engine {
	return &Engine{Store: st, Lookup: lookup}
}

func (e *Engine) Run(ctx context.Context, opts RunOptions) (RunResult, error) {
	var result RunResult
	if ctx == nil {
		ctx = context.Background()
	}
	if err := e.validateOptions(opts); err != nil {
		return result, err
	}

	err := e.Store.WithTx(ctx, func(tx *store.Store) error {
		factorSet, records, periodSettings, electricitySettings, err := e.validateAndLoadInputs(ctx, tx, opts)
		if err != nil {
			return err
		}

		startedAt := time.Now().UTC()
		snapshotJSON, err := buildSettingsSnapshot(opts, startedAt, periodSettings, electricitySettings)
		if err != nil {
			return err
		}

		runID, err := newID("calculation_run")
		if err != nil {
			return err
		}
		run := domain.CalculationRun{
			ID:                   runID,
			OrganizationID:       domain.ID(opts.OrganizationID),
			ReportingPeriodID:    domain.ID(opts.ReportingPeriodID),
			FactorSetID:          domain.ID(opts.FactorSetID),
			StartedAt:            startedAt,
			SettingsSnapshotJSON: snapshotJSON,
		}
		if err := tx.CreateCalculationRun(run); err != nil {
			return err
		}

		lookup := *e.Lookup
		lookup.Store = tx
		lookup.FactorSetID = opts.FactorSetID
		results, err := calculateRecords(ctx, run.ID, factorSet, records, periodSettings, electricitySettings, &lookup)
		if err != nil {
			return err
		}
		for _, calculationResult := range results {
			if err := tx.CreateCalculationResult(calculationResult); err != nil {
				return err
			}
		}

		completedAt := time.Now().UTC()
		if err := tx.CompleteCalculationRun(run.ID, completedAt); err != nil {
			return err
		}

		result = summarizeResults(run.ID, results)
		return nil
	})
	if err != nil {
		return RunResult{}, err
	}

	return result, nil
}

func (e *Engine) validateOptions(opts RunOptions) error {
	if e == nil {
		return invalidOptions("engine is required")
	}
	if e.Store == nil {
		return invalidOptions("store is required")
	}
	if e.Lookup == nil {
		return invalidOptions("lookup is required")
	}
	if opts.OrganizationID == "" {
		return invalidOptions("organization id is required")
	}
	if opts.ReportingPeriodID == "" {
		return invalidOptions("reporting period id is required")
	}
	if opts.FactorSetID == "" {
		return invalidOptions("factor set id is required")
	}
	return nil
}

func (e *Engine) validateAndLoadInputs(ctx context.Context, tx *store.Store, opts RunOptions) (*domain.FactorSet, []domain.ActivityRecord, *domain.ReportingPeriodSettings, map[string]*domain.ElectricitySettings, error) {
	factorSet, err := tx.GetFactorSet(domain.ID(opts.FactorSetID))
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil, nil, nil, invalidOptions("factor set %q does not exist", opts.FactorSetID)
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}

	records, err := tx.ListActiveActivityRecordsByPeriod(domain.ID(opts.ReportingPeriodID))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, nil, nil, ErrNoActiveRecords
	}

	periodSettings, err := tx.GetReportingPeriodSettings(domain.ID(opts.ReportingPeriodID))
	if errors.Is(err, store.ErrNotFound) {
		periodSettings = nil
	} else if err != nil {
		return nil, nil, nil, nil, err
	}

	if err := validateActivityRecords(opts, records, periodSettings); err != nil {
		return nil, nil, nil, nil, err
	}

	electricitySettings, err := loadElectricitySettings(ctx, tx, opts, records)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return factorSet, records, periodSettings, electricitySettings, nil
}

func validateActivityRecords(opts RunOptions, records []domain.ActivityRecord, periodSettings *domain.ReportingPeriodSettings) error {
	hasMobileRecords := false
	for _, record := range records {
		if record.OrganizationID != domain.ID(opts.OrganizationID) {
			return invalidOptions("activity record %q belongs to organization %q, not %q", record.ID, record.OrganizationID, opts.OrganizationID)
		}
		if record.ReportingPeriodID != domain.ID(opts.ReportingPeriodID) {
			return invalidOptions("activity record %q belongs to reporting period %q, not %q", record.ID, record.ReportingPeriodID, opts.ReportingPeriodID)
		}
		if record.Status != domain.ActivityRecordStatusActive {
			return invalidOptions("activity record %q is not active", record.ID)
		}
		if isMobileActivityRecord(record) {
			hasMobileRecords = true
		}
	}

	if periodSettings != nil {
		if periodSettings.OrganizationID != domain.ID(opts.OrganizationID) {
			return invalidSettings("reporting period settings belong to organization %q, not %q", periodSettings.OrganizationID, opts.OrganizationID)
		}
		if periodSettings.ReportingPeriodID != domain.ID(opts.ReportingPeriodID) {
			return invalidSettings("reporting period settings belong to reporting period %q, not %q", periodSettings.ReportingPeriodID, opts.ReportingPeriodID)
		}
	}

	if !hasMobileRecords {
		return nil
	}
	if periodSettings == nil {
		return invalidSettings("reporting period settings are required when mobile combustion records exist")
	}
	if !periodSettings.MobileMethod.Valid() {
		return invalidSettings("mobile method %q is not valid", periodSettings.MobileMethod)
	}

	for _, record := range records {
		if !isMobileActivityRecord(record) {
			continue
		}
		switch periodSettings.MobileMethod {
		case domain.MobileMethodFuelBased:
			if record.SourceKind != domain.ActivitySourceKindMobileFuelLitres || record.Method != domain.ActivityMethodFuelBased {
				return invalidSettings("mobile method %q does not allow activity record %q with source kind %q and method %q", periodSettings.MobileMethod, record.ID, record.SourceKind, record.Method)
			}
		case domain.MobileMethodDistanceBased:
			if record.SourceKind != domain.ActivitySourceKindVehicleDistanceKM || record.Method != domain.ActivityMethodDistanceBased {
				return invalidSettings("mobile method %q does not allow activity record %q with source kind %q and method %q", periodSettings.MobileMethod, record.ID, record.SourceKind, record.Method)
			}
		}
	}

	return nil
}

func isMobileActivityRecord(record domain.ActivityRecord) bool {
	return record.Vector == domain.ActivityVectorMobileCombustion ||
		record.SourceKind == domain.ActivitySourceKindMobileFuelLitres ||
		record.SourceKind == domain.ActivitySourceKindVehicleDistanceKM
}

func loadElectricitySettings(ctx context.Context, tx *store.Store, opts RunOptions, records []domain.ActivityRecord) (map[string]*domain.ElectricitySettings, error) {
	settingsByFacility := map[string]*domain.ElectricitySettings{}
	for _, record := range records {
		if record.SourceKind != domain.ActivitySourceKindElectricityMonthlyKWh {
			continue
		}
		if record.FacilityID == nil || *record.FacilityID == "" {
			return nil, unsupportedRecord("electricity activity record %q is missing facility id", record.ID)
		}

		facilityID := *record.FacilityID
		if _, ok := settingsByFacility[facilityID]; ok {
			continue
		}
		settings, err := tx.GetElectricitySettings(domain.ID(opts.ReportingPeriodID), domain.ID(facilityID))
		if errors.Is(err, store.ErrNotFound) {
			settingsByFacility[facilityID] = nil
			continue
		}
		if err != nil {
			return nil, err
		}
		if settings.OrganizationID != domain.ID(opts.OrganizationID) {
			return nil, invalidSettings("electricity settings for facility %q belong to organization %q, not %q", facilityID, settings.OrganizationID, opts.OrganizationID)
		}
		settingsByFacility[facilityID] = settings
	}
	return settingsByFacility, nil
}

func calculateRecords(ctx context.Context, runID domain.ID, factorSet *domain.FactorSet, records []domain.ActivityRecord, periodSettings *domain.ReportingPeriodSettings, electricitySettings map[string]*domain.ElectricitySettings, lookup *factors.Lookup) ([]domain.CalculationResult, error) {
	results := make([]domain.CalculationResult, 0, len(records))
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var recordResults []domain.CalculationResult
		var err error
		switch record.SourceKind {
		case domain.ActivitySourceKindElectricityMonthlyKWh:
			var settings *domain.ElectricitySettings
			if record.FacilityID != nil {
				settings = electricitySettings[*record.FacilityID]
			}
			recordResults, err = calculateElectricity(ctx, runID, factorSet, record, settings, lookup)
		case domain.ActivitySourceKindNaturalGasMonthlySMC,
			domain.ActivitySourceKindMobileFuelLitres,
			domain.ActivitySourceKindVehicleDistanceKM,
			domain.ActivitySourceKindRefrigerantsAnnualKG:
			recordResults, err = calculateScope1(ctx, runID, factorSet, record, periodSettings, lookup)
		default:
			err = unsupportedRecord("activity record %q has unsupported source kind %q", record.ID, record.SourceKind)
		}
		if err != nil {
			return nil, err
		}
		results = append(results, recordResults...)
	}
	return results, nil
}

func calculationResultFromFactor(runID domain.ID, record domain.ActivityRecord, method domain.ActivityMethod, factorSet *domain.FactorSet, factor *domain.EmissionFactor, isPrimary bool, notesJSON string) (domain.CalculationResult, error) {
	id, err := newID("calculation_result")
	if err != nil {
		return domain.CalculationResult{}, err
	}
	factorID := domain.ID(factor.ID)
	return domain.CalculationResult{
		ID:               id,
		CalculationRunID: runID,
		ActivityRecordID: record.ID,
		Scope:            record.Scope,
		Vector:           record.Vector,
		Method:           method,
		ActivityAmount:   record.Amount,
		ActivityUnit:     record.Unit,
		FactorID:         &factorID,
		FactorValue:      factor.FactorValue,
		FactorUnit:       factor.FactorUnit,
		FactorSource:     factorSource(factorSet, factor),
		EmissionsKgCO2e:  record.Amount * factor.FactorValue,
		IsPrimary:        isPrimary,
		NotesJSON:        notesJSON,
	}, nil
}

func factorSource(factorSet *domain.FactorSet, factor *domain.EmissionFactor) string {
	if factor != nil && factor.Source != "" {
		return factor.Source
	}
	if factorSet != nil {
		return factorSet.Source
	}
	return ""
}

func wrapFactorError(record domain.ActivityRecord, err error) error {
	if errors.Is(err, factors.ErrFactorNotFound) {
		return fmt.Errorf("activity record %q: %w: %v", record.ID, ErrMissingFactor, err)
	}
	if errors.Is(err, factors.ErrUnsupportedActivity) {
		return fmt.Errorf("activity record %q: %w: %v", record.ID, ErrUnsupportedRecord, err)
	}
	return fmt.Errorf("activity record %q: factor lookup: %w", record.ID, err)
}

type settingsSnapshot struct {
	ReportingPeriodSettings *domain.ReportingPeriodSettings `json:"reporting_period_settings"`
	ElectricitySettings     []electricitySettingsSnapshot   `json:"electricity_settings"`
	FactorSetID             string                          `json:"factor_set_id"`
	CalculationTimestamp    string                          `json:"calculation_timestamp"`
	MobileMethod            string                          `json:"mobile_method"`
}

type electricitySettingsSnapshot struct {
	FacilityID               string                      `json:"facility_id"`
	Settings                 *domain.ElectricitySettings `json:"settings"`
	DefaultedToLocationBased bool                        `json:"defaulted_to_location_based"`
}

func buildSettingsSnapshot(opts RunOptions, calculationTime time.Time, periodSettings *domain.ReportingPeriodSettings, electricitySettings map[string]*domain.ElectricitySettings) (string, error) {
	facilityIDs := make([]string, 0, len(electricitySettings))
	for facilityID := range electricitySettings {
		facilityIDs = append(facilityIDs, facilityID)
	}
	sort.Strings(facilityIDs)

	electricity := make([]electricitySettingsSnapshot, 0, len(facilityIDs))
	for _, facilityID := range facilityIDs {
		settings := electricitySettings[facilityID]
		electricity = append(electricity, electricitySettingsSnapshot{
			FacilityID:               facilityID,
			Settings:                 settings,
			DefaultedToLocationBased: settings == nil,
		})
	}

	mobileMethod := ""
	if periodSettings != nil {
		mobileMethod = string(periodSettings.MobileMethod)
	}

	data, err := json.Marshal(settingsSnapshot{
		ReportingPeriodSettings: periodSettings,
		ElectricitySettings:     electricity,
		FactorSetID:             opts.FactorSetID,
		CalculationTimestamp:    calculationTime.UTC().Format(time.RFC3339Nano),
		MobileMethod:            mobileMethod,
	})
	if err != nil {
		return "", fmt.Errorf("marshal calculation settings snapshot: %w", err)
	}
	return string(data), nil
}

func newID(prefix string) (domain.ID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return domain.ID(prefix + "_" + hex.EncodeToString(b[:])), nil
}
