package report

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
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

func TestElectricityWithoutGO(t *testing.T) {
	f := newReportFixture(t)
	record := f.addElectricityRecord(t, "electricity-jan", 1, 1000)
	f.addResult(t, "result-electricity-location", record, domain.ActivityMethodLocationBased, 0.3, "kgCO2e/kWh", "test source", 300, true)

	tables := f.build(t)
	assertFloat(t, tables.ExecutiveSummary.PrimaryTotalKgCO2e, 300)
	if tables.ExecutiveSummary.ElectricityPrimaryMethod != string(domain.ActivityMethodLocationBased) {
		t.Fatalf("electricity primary method = %q, want location_based", tables.ExecutiveSummary.ElectricityPrimaryMethod)
	}
	if tables.ExecutiveSummary.HasLocationBasedComparison {
		t.Fatalf("has location-based comparison = true, want false")
	}
	if tables.ExecutiveSummary.LocationBasedTotalKgCO2e != nil {
		t.Fatalf("location-based total = %v, want nil", *tables.ExecutiveSummary.LocationBasedTotalKgCO2e)
	}

	if len(tables.ElectricityDetail.Rows) != 1 {
		t.Fatalf("electricity detail rows = %d, want 1", len(tables.ElectricityDetail.Rows))
	}
	row := tables.ElectricityDetail.Rows[0]
	if row.PrimaryMethod != string(domain.ActivityMethodLocationBased) || row.LocationBasedKgCO2e == nil {
		t.Fatalf("electricity detail row = %#v, want location-based primary", row)
	}
	assertFloat(t, *row.LocationBasedKgCO2e, 300)
	if row.MarketBasedKgCO2e != nil {
		t.Fatalf("market-based kgCO2e = %v, want nil", *row.MarketBasedKgCO2e)
	}
}

func TestElectricityWithGO(t *testing.T) {
	f := newReportFixture(t)
	record := f.addElectricityRecord(t, "electricity-jan", 1, 1000)
	f.addResult(t, "result-electricity-location", record, domain.ActivityMethodLocationBased, 0.3, "kgCO2e/kWh", "test source", 300, false)
	f.addResult(t, "result-electricity-market", record, domain.ActivityMethodMarketBased, 0, "kgCO2e/kWh", "Guarantees of Origin", 0, true)

	tables := f.build(t)
	assertFloat(t, tables.ExecutiveSummary.PrimaryTotalKgCO2e, 0)
	if tables.ExecutiveSummary.LocationBasedTotalKgCO2e == nil {
		t.Fatalf("location-based total = nil, want 300")
	}
	assertFloat(t, *tables.ExecutiveSummary.LocationBasedTotalKgCO2e, 300)
	if tables.ExecutiveSummary.ElectricityPrimaryMethod != string(domain.ActivityMethodMarketBased) {
		t.Fatalf("electricity primary method = %q, want market_based", tables.ExecutiveSummary.ElectricityPrimaryMethod)
	}

	if len(tables.ElectricityDetail.Rows) != 1 {
		t.Fatalf("electricity detail rows = %d, want 1", len(tables.ElectricityDetail.Rows))
	}
	row := tables.ElectricityDetail.Rows[0]
	if !row.IsMarketBasedPrimary || row.LocationBasedKgCO2e == nil || row.MarketBasedKgCO2e == nil {
		t.Fatalf("electricity detail row = %#v, want both methods with market-based primary", row)
	}
	assertFloat(t, *row.LocationBasedKgCO2e, 300)
	assertFloat(t, *row.MarketBasedKgCO2e, 0)

	if !hasValidationCode(tables.ValidationNotes, "go_market_zero") {
		t.Fatalf("validation notes = %#v, want GO zero note", tables.ValidationNotes.Rows)
	}
	for _, scopeRow := range tables.ScopeSummary.Rows {
		if math.IsNaN(scopeRow.PrimaryShare) {
			t.Fatalf("scope row %#v has NaN primary share", scopeRow)
		}
	}
}

func TestMonthlyTable(t *testing.T) {
	f := newReportFixture(t)
	electricity := f.addElectricityRecord(t, "electricity-jan", 1, 1000)
	f.addResult(t, "result-electricity-location", electricity, domain.ActivityMethodLocationBased, 0.3, "kgCO2e/kWh", "test source", 300, false)
	f.addResult(t, "result-electricity-market", electricity, domain.ActivityMethodMarketBased, 0, "kgCO2e/kWh", "Guarantees of Origin", 0, true)
	naturalGas := f.addNaturalGasRecord(t, "natural-gas-jan", 1, 100)
	f.addResult(t, "result-natural-gas", naturalGas, domain.ActivityMethodFuelBased, 2, "kgCO2e/Smc", "test source", 200, true)

	tables := f.build(t)
	if len(tables.MonthlyEmissions.Rows) != 12 {
		t.Fatalf("monthly rows = %d, want 12", len(tables.MonthlyEmissions.Rows))
	}
	january := tables.MonthlyEmissions.Rows[0]
	assertFloat(t, january.ElectricityPrimaryKgCO2e, 0)
	if january.ElectricityLocationBasedKgCO2e == nil {
		t.Fatalf("january electricity location-based = nil, want 300")
	}
	assertFloat(t, *january.ElectricityLocationBasedKgCO2e, 300)
	assertFloat(t, january.NaturalGasKgCO2e, 200)
	assertFloat(t, january.MonthlyPrimaryKgCO2e, 200)
	if january.MonthlyLocationBasedKgCO2e == nil {
		t.Fatalf("january monthly location-based = nil, want 500")
	}
	assertFloat(t, *january.MonthlyLocationBasedKgCO2e, 500)
}

func TestNaturalGasDetail(t *testing.T) {
	f := newReportFixture(t)
	record := f.addNaturalGasRecord(t, "natural-gas-jan", 1, 100)
	f.addResult(t, "result-natural-gas", record, domain.ActivityMethodFuelBased, 2, "kgCO2e/Smc", "test source", 200, true)

	tables := f.build(t)
	if len(tables.NaturalGasDetail.Rows) != 1 {
		t.Fatalf("natural gas detail rows = %d, want 1", len(tables.NaturalGasDetail.Rows))
	}
	row := tables.NaturalGasDetail.Rows[0]
	assertFloat(t, row.ConsumptionSmc, 100)
	assertFloat(t, row.EmissionsKgCO2e, 200)
	if row.Month != 1 || row.MonthName != "January" {
		t.Fatalf("natural gas month = %d/%q, want January", row.Month, row.MonthName)
	}
}

func TestMobileFuelDetail(t *testing.T) {
	f := newReportFixture(t)
	record := f.addMobileFuelRecord(t, "mobile-fuel", 100)
	f.addResult(t, "result-mobile-fuel", record, domain.ActivityMethodFuelBased, 2.5, "kgCO2e/L", "test source", 250, true)

	tables := f.build(t)
	if tables.MobileDetail.Method != string(domain.ActivityMethodFuelBased) {
		t.Fatalf("mobile method = %q, want fuel_based", tables.MobileDetail.Method)
	}
	assertFloat(t, tables.MobileDetail.TotalKgCO2e, 250)
	if len(tables.MobileDetail.FuelRows) != 1 || len(tables.MobileDetail.DistanceRows) != 0 {
		t.Fatalf("mobile detail = %#v, want one fuel row only", tables.MobileDetail)
	}
	assertFloat(t, tables.MobileDetail.FuelRows[0].Litres, 100)
	assertFloat(t, tables.MobileDetail.FuelRows[0].EmissionsKgCO2e, 250)
}

func TestMobileDistanceDetail(t *testing.T) {
	f := newReportFixture(t)
	record := f.addMobileDistanceRecord(t, "mobile-distance", 1000, string(vocab.FuelPetrol))
	f.addResult(t, "result-mobile-distance", record, domain.ActivityMethodDistanceBased, 0.15, "kgCO2e/km", "test source", 150, true)

	tables := f.build(t)
	if tables.MobileDetail.Method != string(domain.ActivityMethodDistanceBased) {
		t.Fatalf("mobile method = %q, want distance_based", tables.MobileDetail.Method)
	}
	assertFloat(t, tables.MobileDetail.TotalKgCO2e, 150)
	if len(tables.MobileDetail.DistanceRows) != 1 || len(tables.MobileDetail.FuelRows) != 0 {
		t.Fatalf("mobile detail = %#v, want one distance row only", tables.MobileDetail)
	}
	assertFloat(t, tables.MobileDetail.DistanceRows[0].Km, 1000)
	assertFloat(t, tables.MobileDetail.DistanceRows[0].EmissionsKgCO2e, 150)
}

func TestMixedMobileMethodReturnsError(t *testing.T) {
	f := newReportFixture(t)
	fuel := f.addMobileFuelRecord(t, "mobile-fuel", 100)
	distance := f.addMobileDistanceRecord(t, "mobile-distance", 1000, string(vocab.FuelPetrol))
	f.addResult(t, "result-mobile-fuel", fuel, domain.ActivityMethodFuelBased, 2.5, "kgCO2e/L", "test source", 250, true)
	f.addResult(t, "result-mobile-distance", distance, domain.ActivityMethodDistanceBased, 0.15, "kgCO2e/km", "test source", 150, true)

	_, err := f.builder.BuildTables(context.Background(), BuildOptions{CalculationRunID: f.runID})
	if err == nil {
		t.Fatalf("BuildTables error = nil, want error")
	}
	if !errors.Is(err, ErrMixedMobileMethods) {
		t.Fatalf("BuildTables error = %v, want ErrMixedMobileMethods", err)
	}
}

func TestRefrigerantsDetail(t *testing.T) {
	f := newReportFixture(t)
	record := f.addRefrigerantRecord(t, "refrigerant", 2)
	f.addResult(t, "result-refrigerant", record, domain.ActivityMethodDirectGWP, 2088, "kgCO2e/kg", "test source", 4176, true)

	tables := f.build(t)
	if len(tables.RefrigerantsDetail.Rows) != 1 {
		t.Fatalf("refrigerant detail rows = %d, want 1", len(tables.RefrigerantsDetail.Rows))
	}
	row := tables.RefrigerantsDetail.Rows[0]
	if row.Substance != string(vocab.RefrigerantR410A) {
		t.Fatalf("substance = %q, want R410A", row.Substance)
	}
	assertFloat(t, row.QuantityKg, 2)
	assertFloat(t, row.EmissionsKgCO2e, 4176)
}

func TestScopeSummary(t *testing.T) {
	f := newReportFixture(t)
	gas := f.addNaturalGasRecord(t, "natural-gas", 1, 100)
	mobile := f.addMobileFuelRecord(t, "mobile-fuel", 100)
	refrigerant := f.addRefrigerantRecord(t, "refrigerant", 2)
	electricity := f.addElectricityRecord(t, "electricity", 1, 1000)
	f.addResult(t, "result-gas", gas, domain.ActivityMethodFuelBased, 2, "kgCO2e/Smc", "test source", 200, true)
	f.addResult(t, "result-mobile", mobile, domain.ActivityMethodFuelBased, 2.5, "kgCO2e/L", "test source", 250, true)
	f.addResult(t, "result-refrigerant", refrigerant, domain.ActivityMethodDirectGWP, 50, "kgCO2e/kg", "test source", 100, true)
	f.addResult(t, "result-electricity-location", electricity, domain.ActivityMethodLocationBased, 0.3, "kgCO2e/kWh", "test source", 300, false)
	f.addResult(t, "result-electricity-market", electricity, domain.ActivityMethodMarketBased, 0, "kgCO2e/kWh", "Guarantees of Origin", 0, true)

	tables := f.build(t)
	scope1 := findScopeRow(t, tables.ScopeSummary, "Scope 1")
	scope2 := findScopeRow(t, tables.ScopeSummary, "Scope 2")
	total := findScopeRow(t, tables.ScopeSummary, "Total")
	assertFloat(t, scope1.PrimaryKgCO2e, 550)
	assertFloat(t, scope2.PrimaryKgCO2e, 0)
	assertFloat(t, total.PrimaryKgCO2e, 550)
	assertFloat(t, scope1.PrimaryShare, 1)
	assertFloat(t, scope2.PrimaryShare, 0)
	if total.LocationBasedKgCO2e == nil {
		t.Fatalf("location-based total = nil, want 850")
	}
	assertFloat(t, *total.LocationBasedKgCO2e, 850)
	assertFloat(t, *scope1.LocationBasedShare, 550.0/850.0)
	assertFloat(t, *scope2.LocationBasedShare, 300.0/850.0)

	zero := newReportFixture(t).build(t)
	for _, row := range zero.ScopeSummary.Rows {
		if math.IsNaN(row.PrimaryShare) || row.PrimaryShare != 0 {
			t.Fatalf("zero-total scope row = %#v, want zero non-NaN share", row)
		}
	}
}

func TestVectorSummary(t *testing.T) {
	f := newReportFixture(t)
	electricity := f.addElectricityRecord(t, "electricity", 1, 1000)
	gas := f.addNaturalGasRecord(t, "natural-gas", 1, 100)
	mobile := f.addMobileFuelRecord(t, "mobile-fuel", 100)
	refrigerant := f.addRefrigerantRecord(t, "refrigerant", 2)
	f.addResult(t, "result-electricity", electricity, domain.ActivityMethodLocationBased, 0.3, "kgCO2e/kWh", "test source", 300, true)
	f.addResult(t, "result-gas", gas, domain.ActivityMethodFuelBased, 2, "kgCO2e/Smc", "test source", 200, true)
	f.addResult(t, "result-mobile", mobile, domain.ActivityMethodFuelBased, 2.5, "kgCO2e/L", "test source", 250, true)
	f.addResult(t, "result-refrigerant", refrigerant, domain.ActivityMethodDirectGWP, 50, "kgCO2e/kg", "test source", 100, true)

	tables := f.build(t)
	electricityRow := findVectorRow(t, tables.VectorSummary, "Electricity")
	gasRow := findVectorRow(t, tables.VectorSummary, "Natural gas")
	mobileRow := findVectorRow(t, tables.VectorSummary, "Mobile combustion")
	refrigerantsRow := findVectorRow(t, tables.VectorSummary, "Refrigerants")
	if electricityRow.ActivitySummary != "1000 kWh" || gasRow.ActivitySummary != "100 Smc" || mobileRow.ActivitySummary != "100 L" || refrigerantsRow.ActivitySummary != "2 kg" {
		t.Fatalf("activity summaries = %q, %q, %q, %q", electricityRow.ActivitySummary, gasRow.ActivitySummary, mobileRow.ActivitySummary, refrigerantsRow.ActivitySummary)
	}
	assertFloat(t, electricityRow.PrimaryShare, 300.0/850.0)
	assertFloat(t, gasRow.PrimaryShare, 200.0/850.0)
	assertFloat(t, mobileRow.PrimaryShare, 250.0/850.0)
	assertFloat(t, refrigerantsRow.PrimaryShare, 100.0/850.0)
}

func TestNoRecalculation(t *testing.T) {
	f := newReportFixture(t)
	record := f.addElectricityRecord(t, "electricity", 1, 1000)
	f.addResult(t, "result-electricity", record, domain.ActivityMethodLocationBased, 0.3, "kgCO2e/kWh", "test source", 123, true)

	tables := f.build(t)
	assertFloat(t, tables.ExecutiveSummary.PrimaryTotalKgCO2e, 123)
	assertFloat(t, *tables.ElectricityDetail.Rows[0].LocationBasedKgCO2e, 123)
}

func TestNoXLSXDependency(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read report package dir: %v", err)
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
		if strings.Contains(source, "xlsx") || strings.Contains(source, "excelize") || strings.Contains(source, "internal/factors") {
			t.Fatalf("%s has a forbidden report dependency", entry.Name())
		}
	}
}

type reportFixture struct {
	store       *store.Store
	builder     *Builder
	orgID       domain.ID
	facilityID  domain.ID
	periodID    domain.ID
	factorSetID domain.ID
	runID       domain.ID
	now         time.Time
}

func newReportFixture(t *testing.T) *reportFixture {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "report.sqlite"))
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
	f := &reportFixture{
		store:       st,
		builder:     NewBuilder(st),
		orgID:       "org-1",
		facilityID:  "facility-1",
		periodID:    "period-2026",
		factorSetID: "factor-set-1",
		runID:       "calculation-run-1",
		now:         now,
	}
	if err := st.CreateOrganization(domain.Organization{ID: f.orgID, Name: "Acme Ltd", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if err := st.CreateFacility(domain.Facility{ID: f.facilityID, OrganizationID: f.orgID, Name: "Milan Office", CountryCode: "IT", CreatedAt: now, UpdatedAt: now}); err != nil {
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
	completedAt := now
	if err := st.CreateCalculationRun(domain.CalculationRun{
		ID:                   f.runID,
		OrganizationID:       f.orgID,
		ReportingPeriodID:    f.periodID,
		FactorSetID:          f.factorSetID,
		StartedAt:            now,
		CompletedAt:          &completedAt,
		SettingsSnapshotJSON: `{"mobile_method":"fuel_based"}`,
	}); err != nil {
		t.Fatalf("create calculation run: %v", err)
	}
	return f
}

func (f *reportFixture) build(t *testing.T) *ReportTables {
	t.Helper()
	tables, err := f.builder.BuildTables(context.Background(), BuildOptions{CalculationRunID: f.runID})
	if err != nil {
		t.Fatalf("BuildTables: %v", err)
	}
	return tables
}

func (f *reportFixture) addElectricityRecord(t *testing.T, id string, month int, amount float64) domain.ActivityRecord {
	t.Helper()
	facilityID := f.facilityID
	return f.addActivityRecord(t, domain.ActivityRecord{
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
		Status:       domain.ActivityRecordStatusActive,
	}, month)
}

func (f *reportFixture) addNaturalGasRecord(t *testing.T, id string, month int, amount float64) domain.ActivityRecord {
	t.Helper()
	facilityID := f.facilityID
	return f.addActivityRecord(t, domain.ActivityRecord{
		ID:           domain.ID(id),
		FacilityID:   &facilityID,
		SourceKind:   domain.ActivitySourceKindNaturalGasMonthlySMC,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorNaturalGas,
		Category:     "stationary_combustion",
		Method:       domain.ActivityMethodFuelBased,
		ActivityType: "natural_gas",
		Amount:       amount,
		Unit:         string(vocab.UnitSmc),
		Status:       domain.ActivityRecordStatusActive,
	}, month)
}

func (f *reportFixture) addMobileFuelRecord(t *testing.T, id string, amount float64) domain.ActivityRecord {
	t.Helper()
	return f.addActivityRecord(t, domain.ActivityRecord{
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
		Status:       domain.ActivityRecordStatusActive,
	}, 1)
}

func (f *reportFixture) addMobileDistanceRecord(t *testing.T, id string, amount float64, fuelType string) domain.ActivityRecord {
	t.Helper()
	return f.addActivityRecord(t, domain.ActivityRecord{
		ID:               domain.ID(id),
		SourceKind:       domain.ActivitySourceKindVehicleDistanceKM,
		Scope:            domain.Scope1,
		Vector:           domain.ActivityVectorMobileCombustion,
		Category:         "mobile_combustion",
		Method:           domain.ActivityMethodDistanceBased,
		ActivityType:     "vehicle_distance",
		Amount:           amount,
		Unit:             string(vocab.UnitKm),
		FuelType:         fuelType,
		VehicleName:      "Car 1",
		Plate:            "AA123BB",
		VehicleType:      string(vocab.VehicleCar),
		VehicleSizeClass: string(vocab.SizeSmall),
		Status:           domain.ActivityRecordStatusActive,
	}, 1)
}

func (f *reportFixture) addRefrigerantRecord(t *testing.T, id string, amount float64) domain.ActivityRecord {
	t.Helper()
	facilityID := f.facilityID
	return f.addActivityRecord(t, domain.ActivityRecord{
		ID:           domain.ID(id),
		FacilityID:   &facilityID,
		SourceKind:   domain.ActivitySourceKindRefrigerantsAnnualKG,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorRefrigerants,
		Category:     "fugitive_emissions",
		Method:       domain.ActivityMethodDirectGWP,
		ActivityType: "refrigerant_leakage",
		Amount:       amount,
		Unit:         string(vocab.UnitKg),
		Substance:    string(vocab.RefrigerantR410A),
		Status:       domain.ActivityRecordStatusActive,
	}, 1)
}

func (f *reportFixture) addActivityRecord(t *testing.T, record domain.ActivityRecord, month int) domain.ActivityRecord {
	t.Helper()
	record.OrganizationID = f.orgID
	record.ReportingPeriodID = f.periodID
	record.PeriodStart = time.Date(2026, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	record.PeriodEnd = record.PeriodStart.AddDate(0, 1, -1)
	record.SourceHash = string(record.ID) + "-hash"
	record.CreatedAt = f.now
	record.UpdatedAt = f.now
	if err := f.store.CreateActivityRecord(record); err != nil {
		t.Fatalf("create activity record %q: %v", record.ID, err)
	}
	return record
}

func (f *reportFixture) addResult(t *testing.T, id string, record domain.ActivityRecord, method domain.ActivityMethod, factorValue float64, factorUnit string, factorSource string, emissions float64, primary bool) {
	t.Helper()
	if err := f.store.CreateCalculationResult(domain.CalculationResult{
		ID:               domain.ID(id),
		CalculationRunID: f.runID,
		ActivityRecordID: record.ID,
		Scope:            record.Scope,
		Vector:           record.Vector,
		Method:           method,
		ActivityAmount:   record.Amount,
		ActivityUnit:     record.Unit,
		FactorValue:      factorValue,
		FactorUnit:       factorUnit,
		FactorSource:     factorSource,
		EmissionsKgCO2e:  emissions,
		IsPrimary:        primary,
		NotesJSON:        `{}`,
	}); err != nil {
		t.Fatalf("create calculation result %q: %v", id, err)
	}
}

func findScopeRow(t *testing.T, table ScopeSummaryTable, scope string) ScopeSummaryRow {
	t.Helper()
	for _, row := range table.Rows {
		if row.Scope == scope {
			return row
		}
	}
	t.Fatalf("scope row %q not found in %#v", scope, table.Rows)
	return ScopeSummaryRow{}
}

func findVectorRow(t *testing.T, table VectorSummaryTable, vector string) VectorSummaryRow {
	t.Helper()
	for _, row := range table.Rows {
		if row.Vector == vector {
			return row
		}
	}
	t.Fatalf("vector row %q not found in %#v", vector, table.Rows)
	return VectorSummaryRow{}
}

func hasValidationCode(table ValidationNotesTable, code string) bool {
	for _, row := range table.Rows {
		if row.Code == code {
			return true
		}
	}
	return false
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}
