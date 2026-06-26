package calc

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/input"
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

func TestNaturalGasCalculation(t *testing.T) {
	f := newCalcFixture(t)
	f.addNaturalGasFactor(t, 2.0)
	f.addNaturalGasRecord(t, "natural-gas-1", 100, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 200)
	if run.LocationBasedTotalKgCO2e != nil {
		t.Fatalf("location-based total = %v, want nil", *run.LocationBasedTotalKgCO2e)
	}

	results := f.results(t, run.CalculationRunID)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	assertResult(t, results[0], domain.ActivityMethodFuelBased, 200, true)
}

func TestElectricityWithoutGO(t *testing.T) {
	f := newCalcFixture(t)
	f.addElectricityFactor(t, 0.3)
	f.addElectricityRecord(t, "electricity-1", 1000, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 300)
	if run.LocationBasedTotalKgCO2e != nil {
		t.Fatalf("location-based total = %v, want nil", *run.LocationBasedTotalKgCO2e)
	}

	results := f.results(t, run.CalculationRunID)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	assertResult(t, results[0], domain.ActivityMethodLocationBased, 300, true)
}

func TestElectricityWithFullGO(t *testing.T) {
	f := newCalcFixture(t)
	f.addElectricityFactor(t, 0.3)
	f.addElectricityRecord(t, "electricity-1", 1000, domain.ActivityRecordStatusActive)
	f.addElectricitySettings(t, true, domain.GOCoverageFull)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 0)
	if run.LocationBasedTotalKgCO2e == nil {
		t.Fatalf("location-based total = nil, want 300")
	}
	assertFloat(t, *run.LocationBasedTotalKgCO2e, 300)

	results := f.results(t, run.CalculationRunID)
	if len(results) != 2 {
		t.Fatalf("results count = %d, want 2", len(results))
	}
	locationBased := findResult(t, results, domain.ActivityMethodLocationBased)
	marketBased := findResult(t, results, domain.ActivityMethodMarketBased)
	assertResult(t, locationBased, domain.ActivityMethodLocationBased, 300, false)
	assertResult(t, marketBased, domain.ActivityMethodMarketBased, 0, true)
	if marketBased.FactorID != nil {
		t.Fatalf("market-based factor id = %q, want nil", *marketBased.FactorID)
	}
	if marketBased.FactorSource != guaranteesOfOriginSource {
		t.Fatalf("market-based factor source = %q, want %q", marketBased.FactorSource, guaranteesOfOriginSource)
	}
	if !strings.Contains(marketBased.NotesJSON, "full GO coverage") {
		t.Fatalf("market-based notes = %q, want full GO coverage note", marketBased.NotesJSON)
	}
}

func TestElectricityGOInconsistencyRollsBack(t *testing.T) {
	f := newCalcFixture(t)
	f.addElectricityFactor(t, 0.3)
	f.addElectricityRecord(t, "electricity-1", 1000, domain.ActivityRecordStatusActive)
	f.addElectricitySettings(t, true, domain.GOCoverageNone)

	_, err := f.engine.Run(context.Background(), f.options())
	if err == nil {
		t.Fatalf("run error = nil, want error")
	}
	if !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("run error = %v, want ErrInvalidSettings", err)
	}
	f.assertNoRuns(t)
}

func TestMobileFuelMethod(t *testing.T) {
	f := newCalcFixture(t)
	f.addReportingPeriodSettings(t, domain.MobileMethodFuelBased)
	f.addMobileFuelFactor(t, 2.5)
	f.addMobileFuelRecord(t, "mobile-fuel-1", 100, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 250)

	results := f.results(t, run.CalculationRunID)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	assertResult(t, results[0], domain.ActivityMethodFuelBased, 250, true)
}

func TestMobileDistanceMethod(t *testing.T) {
	f := newCalcFixture(t)
	f.addReportingPeriodSettings(t, domain.MobileMethodDistanceBased)
	f.addVehicleDistanceFactor(t, 0.15)
	f.addVehicleDistanceRecord(t, "vehicle-distance-1", 1000, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 150)

	results := f.results(t, run.CalculationRunID)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	assertResult(t, results[0], domain.ActivityMethodDistanceBased, 150, true)
}

func TestMobileMethodMismatchRollsBack(t *testing.T) {
	f := newCalcFixture(t)
	f.addReportingPeriodSettings(t, domain.MobileMethodFuelBased)
	f.addVehicleDistanceFactor(t, 0.15)
	f.addVehicleDistanceRecord(t, "vehicle-distance-1", 1000, domain.ActivityRecordStatusActive)

	_, err := f.engine.Run(context.Background(), f.options())
	if err == nil {
		t.Fatalf("run error = nil, want error")
	}
	if !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("run error = %v, want ErrInvalidSettings", err)
	}
	f.assertNoRuns(t)
}

func TestRefrigerantCalculation(t *testing.T) {
	f := newCalcFixture(t)
	f.addRefrigerantFactor(t, 2088)
	f.addRefrigerantRecord(t, "refrigerant-1", 2, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 4176)

	results := f.results(t, run.CalculationRunID)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	assertResult(t, results[0], domain.ActivityMethodDirectGWP, 4176, true)
}

func TestInactiveRecordsIgnored(t *testing.T) {
	f := newCalcFixture(t)
	f.addNaturalGasFactor(t, 2.0)
	f.addNaturalGasRecord(t, "natural-gas-active", 100, domain.ActivityRecordStatusActive)
	f.addNaturalGasRecord(t, "natural-gas-superseded", 999, domain.ActivityRecordStatusSuperseded)

	run := f.run(t)
	if run.ResultsCreated != 1 {
		t.Fatalf("results created = %d, want 1", run.ResultsCreated)
	}
	assertFloat(t, run.PrimaryTotalKgCO2e, 200)
}

func TestMissingFactorRollsBack(t *testing.T) {
	f := newCalcFixture(t)
	f.addNaturalGasRecord(t, "natural-gas-1", 100, domain.ActivityRecordStatusActive)

	_, err := f.engine.Run(context.Background(), f.options())
	if err == nil {
		t.Fatalf("run error = nil, want error")
	}
	if !errors.Is(err, ErrMissingFactor) {
		t.Fatalf("run error = %v, want ErrMissingFactor", err)
	}
	f.assertNoRuns(t)
}

func TestMultipleVectorsTotal(t *testing.T) {
	f := newCalcFixture(t)
	f.addReportingPeriodSettings(t, domain.MobileMethodFuelBased)
	f.addNaturalGasFactor(t, 2.0)
	f.addMobileFuelFactor(t, 2.5)
	f.addRefrigerantFactor(t, 50)
	f.addElectricityFactor(t, 0.3)
	f.addNaturalGasRecord(t, "natural-gas-1", 100, domain.ActivityRecordStatusActive)
	f.addMobileFuelRecord(t, "mobile-fuel-1", 100, domain.ActivityRecordStatusActive)
	f.addRefrigerantRecord(t, "refrigerant-1", 2, domain.ActivityRecordStatusActive)
	f.addElectricityRecord(t, "electricity-1", 1000, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 850)
	if run.LocationBasedTotalKgCO2e != nil {
		t.Fatalf("location-based total = %v, want nil", *run.LocationBasedTotalKgCO2e)
	}
	if run.ResultsCreated != 4 {
		t.Fatalf("results created = %d, want 4", run.ResultsCreated)
	}
}

func TestMultipleVectorsWithGO(t *testing.T) {
	f := newCalcFixture(t)
	f.addReportingPeriodSettings(t, domain.MobileMethodFuelBased)
	f.addNaturalGasFactor(t, 2.0)
	f.addMobileFuelFactor(t, 2.5)
	f.addRefrigerantFactor(t, 50)
	f.addElectricityFactor(t, 0.3)
	f.addElectricitySettings(t, true, domain.GOCoverageFull)
	f.addNaturalGasRecord(t, "natural-gas-1", 100, domain.ActivityRecordStatusActive)
	f.addMobileFuelRecord(t, "mobile-fuel-1", 100, domain.ActivityRecordStatusActive)
	f.addRefrigerantRecord(t, "refrigerant-1", 2, domain.ActivityRecordStatusActive)
	f.addElectricityRecord(t, "electricity-1", 1000, domain.ActivityRecordStatusActive)

	run := f.run(t)
	assertFloat(t, run.PrimaryTotalKgCO2e, 550)
	if run.LocationBasedTotalKgCO2e == nil {
		t.Fatalf("location-based total = nil, want 850")
	}
	assertFloat(t, *run.LocationBasedTotalKgCO2e, 850)
	if run.ResultsCreated != 5 {
		t.Fatalf("results created = %d, want 5", run.ResultsCreated)
	}
}

func TestMultipleCalculationRunsAreKept(t *testing.T) {
	f := newCalcFixture(t)
	f.addNaturalGasFactor(t, 2.0)
	f.addNaturalGasRecord(t, "natural-gas-1", 100, domain.ActivityRecordStatusActive)

	first := f.run(t)
	second := f.run(t)
	if first.CalculationRunID == second.CalculationRunID {
		t.Fatalf("calculation run ids are equal: %q", first.CalculationRunID)
	}

	runs, err := f.store.ListCalculationRunsByPeriod(f.periodID)
	if err != nil {
		t.Fatalf("list calculation runs: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("calculation runs count = %d, want 2", len(runs))
	}
	for _, run := range runs {
		results := f.results(t, run.ID)
		if len(results) != 1 {
			t.Fatalf("run %q results count = %d, want 1", run.ID, len(results))
		}
	}
}

func TestNoXLSXDependency(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read calc package dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		contents, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		source := strings.ToLower(string(contents))
		if strings.Contains(source, "xlsx") || strings.Contains(source, "excelize") {
			t.Fatalf("%s mentions XLSX or Excel dependencies", entry.Name())
		}
	}
}

func TestCalculationWithSeededDefaultFactors(t *testing.T) {
	f := newCalcFixture(t)
	factorSet, err := factors.EnsureDefaultFactors(context.Background(), f.store)
	if err != nil {
		t.Fatalf("ensure default factors: %v", err)
	}

	facilityID := string(f.facilityID)
	parsed := input.Parse(vocab.InputNaturalGasMonthlySmc, "January\t100")
	_, err = input.CommitParsedInput(context.Background(), f.store, input.CommitContext{
		OrganizationID:    string(f.orgID),
		ReportingPeriodID: string(f.periodID),
		FacilityID:        &facilityID,
		ReportingYear:     2026,
		PeriodStart:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		InputKind:         vocab.InputNaturalGasMonthlySmc,
	}, parsed)
	if err != nil {
		t.Fatalf("commit natural gas: %v", err)
	}

	engine := NewEngine(f.store, factors.NewLookup(f.store, factorSet.ID))
	result, err := engine.Run(context.Background(), RunOptions{
		OrganizationID:    string(f.orgID),
		ReportingPeriodID: string(f.periodID),
		FactorSetID:       factorSet.ID,
	})
	if err != nil {
		t.Fatalf("run calculation with seeded factors: %v", err)
	}
	if result.PrimaryTotalKgCO2e <= 0 {
		t.Fatalf("primary total = %v, want positive seeded-factor result", result.PrimaryTotalKgCO2e)
	}
}

type calcFixture struct {
	store       *store.Store
	engine      *Engine
	orgID       domain.ID
	facilityID  domain.ID
	periodID    domain.ID
	factorSetID domain.ID
	now         time.Time
}

func newCalcFixture(t *testing.T) *calcFixture {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "calc.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	st := store.New(db)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	f := &calcFixture{
		store:       st,
		orgID:       "org-1",
		facilityID:  "facility-1",
		periodID:    "period-2026",
		factorSetID: "factor-set-1",
		now:         now,
	}

	if err := st.CreateOrganization(domain.Organization{
		ID:        f.orgID,
		Name:      "Acme Ltd",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if err := st.CreateFacility(domain.Facility{
		ID:             f.facilityID,
		OrganizationID: f.orgID,
		Name:           "Milan Office",
		CountryCode:    "IT",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create facility: %v", err)
	}
	if err := st.CreateReportingPeriod(domain.ReportingPeriod{
		ID:             f.periodID,
		OrganizationID: f.orgID,
		Year:           2026,
		StartsOn:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndsOn:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:         domain.ReportingPeriodStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create reporting period: %v", err)
	}
	if err := st.CreateFactorSet(domain.FactorSet{
		ID:           f.factorSetID,
		Name:         "Test Factors",
		Source:       "test source",
		Year:         2025,
		Version:      "test",
		ImportedAt:   now,
		MetadataJSON: `{}`,
	}); err != nil {
		t.Fatalf("create factor set: %v", err)
	}

	f.engine = NewEngine(st, factors.NewLookup(st, string(f.factorSetID)))
	return f
}

func (f *calcFixture) options() RunOptions {
	return RunOptions{
		OrganizationID:    string(f.orgID),
		ReportingPeriodID: string(f.periodID),
		FactorSetID:       string(f.factorSetID),
	}
}

func (f *calcFixture) run(t *testing.T) RunResult {
	t.Helper()
	result, err := f.engine.Run(context.Background(), f.options())
	if err != nil {
		t.Fatalf("run calculation: %v", err)
	}
	return result
}

func (f *calcFixture) results(t *testing.T, runID domain.ID) []domain.CalculationResult {
	t.Helper()
	results, err := f.store.ListCalculationResultsByRun(runID)
	if err != nil {
		t.Fatalf("list calculation results: %v", err)
	}
	return results
}

func (f *calcFixture) assertNoRuns(t *testing.T) {
	t.Helper()
	runs, err := f.store.ListCalculationRunsByPeriod(f.periodID)
	if err != nil {
		t.Fatalf("list calculation runs: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("calculation runs count = %d, want 0", len(runs))
	}
}

func (f *calcFixture) addReportingPeriodSettings(t *testing.T, method domain.MobileMethod) {
	t.Helper()
	if err := f.store.UpsertReportingPeriodSettings(domain.ReportingPeriodSettings{
		ID:                "reporting-period-settings-1",
		OrganizationID:    f.orgID,
		ReportingPeriodID: f.periodID,
		MobileMethod:      method,
		CreatedAt:         f.now,
		UpdatedAt:         f.now,
	}); err != nil {
		t.Fatalf("upsert reporting period settings: %v", err)
	}
}

func (f *calcFixture) addElectricitySettings(t *testing.T, hasGO bool, coverage domain.GOCoverage) {
	t.Helper()
	if err := f.store.UpsertElectricitySettings(domain.ElectricitySettings{
		ID:                    "electricity-settings-1",
		OrganizationID:        f.orgID,
		ReportingPeriodID:     f.periodID,
		FacilityID:            f.facilityID,
		HasGuaranteesOfOrigin: hasGO,
		GOCoverage:            coverage,
		GOReference:           "GO-123",
		GOMarket:              "IT",
		CreatedAt:             f.now,
		UpdatedAt:             f.now,
	}); err != nil {
		t.Fatalf("upsert electricity settings: %v", err)
	}
}

func (f *calcFixture) addNaturalGasFactor(t *testing.T, value float64) {
	t.Helper()
	f.addFactor(t, domain.EmissionFactor{
		ID:           "factor-natural-gas",
		Scope:        domain.Scope1,
		ActivityType: "natural_gas",
		InputUnit:    string(vocab.UnitSmc),
		FactorUnit:   "kgCO2e/Smc",
		FactorValue:  value,
	})
}

func (f *calcFixture) addElectricityFactor(t *testing.T, value float64) {
	t.Helper()
	f.addFactor(t, domain.EmissionFactor{
		ID:           "factor-electricity",
		Scope:        domain.Scope2,
		ActivityType: "purchased_electricity",
		InputUnit:    string(vocab.UnitKWh),
		FactorUnit:   "kgCO2e/kWh",
		FactorValue:  value,
	})
}

func (f *calcFixture) addMobileFuelFactor(t *testing.T, value float64) {
	t.Helper()
	f.addFactor(t, domain.EmissionFactor{
		ID:           "factor-mobile-diesel",
		Scope:        domain.Scope1,
		ActivityType: "diesel_mobile",
		FuelType:     string(vocab.FuelDiesel),
		InputUnit:    string(vocab.UnitLitre),
		FactorUnit:   "kgCO2e/L",
		FactorValue:  value,
	})
}

func (f *calcFixture) addVehicleDistanceFactor(t *testing.T, value float64) {
	t.Helper()
	f.addFactor(t, domain.EmissionFactor{
		ID:               "factor-vehicle-distance",
		Scope:            domain.Scope1,
		ActivityType:     "vehicle_distance",
		FuelType:         string(vocab.FuelPetrol),
		VehicleType:      string(vocab.VehicleCar),
		VehicleSizeClass: string(vocab.SizeSmall),
		InputUnit:        string(vocab.UnitKm),
		FactorUnit:       "kgCO2e/km",
		FactorValue:      value,
	})
}

func (f *calcFixture) addRefrigerantFactor(t *testing.T, value float64) {
	t.Helper()
	f.addFactor(t, domain.EmissionFactor{
		ID:           "factor-refrigerant-r410a",
		Scope:        domain.Scope1,
		ActivityType: "refrigerant_leakage",
		Substance:    string(vocab.RefrigerantR410A),
		InputUnit:    string(vocab.UnitKg),
		FactorUnit:   "kgCO2e/kg",
		FactorValue:  value,
	})
}

func (f *calcFixture) addFactor(t *testing.T, factor domain.EmissionFactor) {
	t.Helper()
	factor.FactorSetID = f.factorSetID
	factor.Source = "test source"
	factor.GHG = "kgCO2e"
	factor.MetadataJSON = `{}`
	if err := f.store.CreateEmissionFactor(factor); err != nil {
		t.Fatalf("create emission factor %q: %v", factor.ID, err)
	}
}

func (f *calcFixture) addNaturalGasRecord(t *testing.T, id string, amount float64, status domain.ActivityRecordStatus) {
	t.Helper()
	f.addActivityRecord(t, domain.ActivityRecord{
		ID:           domain.ID(id),
		SourceKind:   domain.ActivitySourceKindNaturalGasMonthlySMC,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorNaturalGas,
		Category:     "stationary_combustion",
		Method:       domain.ActivityMethodFuelBased,
		ActivityType: "natural_gas",
		Amount:       amount,
		Unit:         string(vocab.UnitSmc),
		Status:       status,
	})
}

func (f *calcFixture) addElectricityRecord(t *testing.T, id string, amount float64, status domain.ActivityRecordStatus) {
	t.Helper()
	facilityID := f.facilityID
	f.addActivityRecord(t, domain.ActivityRecord{
		ID:           domain.ID(id),
		FacilityID:   &facilityID,
		SourceKind:   domain.ActivitySourceKindElectricityMonthlyKWh,
		Scope:        domain.Scope2,
		Vector:       domain.ActivityVectorElectricity,
		Category:     "purchased_electricity",
		Method:       domain.ActivityMethodLocationBased,
		ActivityType: "purchased_electricity",
		Amount:       amount,
		Unit:         string(vocab.UnitKWh),
		Status:       status,
	})
}

func (f *calcFixture) addMobileFuelRecord(t *testing.T, id string, amount float64, status domain.ActivityRecordStatus) {
	t.Helper()
	f.addActivityRecord(t, domain.ActivityRecord{
		ID:           domain.ID(id),
		SourceKind:   domain.ActivitySourceKindMobileFuelLitres,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorMobileCombustion,
		Category:     "mobile_combustion",
		Method:       domain.ActivityMethodFuelBased,
		ActivityType: "diesel_mobile",
		Amount:       amount,
		Unit:         string(vocab.UnitLitre),
		FuelType:     string(vocab.FuelDiesel),
		Status:       status,
	})
}

func (f *calcFixture) addVehicleDistanceRecord(t *testing.T, id string, amount float64, status domain.ActivityRecordStatus) {
	t.Helper()
	f.addActivityRecord(t, domain.ActivityRecord{
		ID:               domain.ID(id),
		SourceKind:       domain.ActivitySourceKindVehicleDistanceKM,
		Scope:            domain.Scope1,
		Vector:           domain.ActivityVectorMobileCombustion,
		Category:         "mobile_combustion",
		Method:           domain.ActivityMethodDistanceBased,
		ActivityType:     "vehicle_distance",
		Amount:           amount,
		Unit:             string(vocab.UnitKm),
		FuelType:         string(vocab.FuelPetrol),
		VehicleType:      string(vocab.VehicleCar),
		VehicleSizeClass: string(vocab.SizeSmall),
		Status:           status,
	})
}

func (f *calcFixture) addRefrigerantRecord(t *testing.T, id string, amount float64, status domain.ActivityRecordStatus) {
	t.Helper()
	f.addActivityRecord(t, domain.ActivityRecord{
		ID:           domain.ID(id),
		SourceKind:   domain.ActivitySourceKindRefrigerantsAnnualKG,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorRefrigerants,
		Category:     "fugitive_emissions",
		Method:       domain.ActivityMethodDirectGWP,
		ActivityType: "refrigerant_leakage",
		Amount:       amount,
		Unit:         string(vocab.UnitKg),
		Substance:    string(vocab.RefrigerantR410A),
		Status:       status,
	})
}

func (f *calcFixture) addActivityRecord(t *testing.T, record domain.ActivityRecord) {
	t.Helper()
	if record.Status == "" {
		record.Status = domain.ActivityRecordStatusActive
	}
	record.OrganizationID = f.orgID
	record.ReportingPeriodID = f.periodID
	record.PeriodStart = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	record.PeriodEnd = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	record.SourceHash = string(record.ID) + "-hash"
	record.CreatedAt = f.now
	record.UpdatedAt = f.now
	if err := f.store.CreateActivityRecord(record); err != nil {
		t.Fatalf("create activity record %q: %v", record.ID, err)
	}
}

func findResult(t *testing.T, results []domain.CalculationResult, method domain.ActivityMethod) domain.CalculationResult {
	t.Helper()
	for _, result := range results {
		if result.Method == method {
			return result
		}
	}
	t.Fatalf("result method %q not found in %#v", method, results)
	return domain.CalculationResult{}
}

func assertResult(t *testing.T, result domain.CalculationResult, method domain.ActivityMethod, emissions float64, isPrimary bool) {
	t.Helper()
	if result.Method != method {
		t.Fatalf("result method = %q, want %q", result.Method, method)
	}
	assertFloat(t, result.EmissionsKgCO2e, emissions)
	if result.IsPrimary != isPrimary {
		t.Fatalf("result is_primary = %t, want %t", result.IsPrimary, isPrimary)
	}
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}
