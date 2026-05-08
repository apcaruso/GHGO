package input

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"testing"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

func TestCommitElectricity(t *testing.T) {
	fixture := newCommitFixture(t, true)
	raw := "January\t100\nFebruary\t200"
	parsed := Parse(vocab.InputElectricityMonthlyKWh, raw)

	result, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(vocab.InputElectricityMonthlyKWh), parsed)
	if err != nil {
		t.Fatalf("commit electricity: %v", err)
	}
	if !result.Committed || len(result.ActivityRecordIDs) != 2 || result.RowsTotal != 2 || result.RowsValid != 2 || result.RowsError != 0 {
		t.Fatalf("result = %#v, want committed with two records", result)
	}

	batch, err := fixture.st.GetPasteBatch(result.PasteBatchID)
	if err != nil {
		t.Fatalf("get paste batch: %v", err)
	}
	if batch.Status != domain.PasteBatchStatusCommitted || batch.CommittedAt == nil {
		t.Fatalf("batch = %#v, want committed", batch)
	}
	pasteRows := requirePasteRows(t, fixture.st, result.PasteBatchID, 2)
	if pasteRows[0].Status != domain.PasteRowStatusCommitted || pasteRows[0].ErrorsJSON != "[]" {
		t.Fatalf("paste row = %#v, want committed with [] errors", pasteRows[0])
	}

	records := requireActivityRecords(t, fixture.st, fixture.periodID, 2)
	for _, record := range records {
		if record.Unit != "kWh" || record.Scope != domain.Scope2 || record.Vector != domain.ActivityVectorElectricity {
			t.Fatalf("record = %#v, want kWh scope 2 electricity", record)
		}
	}

	duplicate := Parse(vocab.InputElectricityMonthlyKWh, "Jan\t100\nJanuary\t200")
	if duplicate.RowsError == 0 {
		t.Fatalf("duplicate month RowsError = 0, want parser error")
	}
}

func TestCommitNaturalGas(t *testing.T) {
	fixture := newCommitFixture(t, true)
	parsed := Parse(vocab.InputNaturalGasMonthlySmc, "January\t30\nFebruary\t40")

	_, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(vocab.InputNaturalGasMonthlySmc), parsed)
	if err != nil {
		t.Fatalf("commit natural gas: %v", err)
	}

	records := requireActivityRecords(t, fixture.st, fixture.periodID, 2)
	for _, record := range records {
		if record.Unit != "Smc" || record.Scope != domain.Scope1 || record.Vector != domain.ActivityVectorNaturalGas {
			t.Fatalf("record = %#v, want Smc scope 1 natural_gas", record)
		}
	}
}

func TestCommitMobileFuel(t *testing.T) {
	fixture := newCommitFixture(t, true)
	parsed := Parse(vocab.InputMobileFuelLitres, "Diesel\t2000\ngasolio\t2200\nPetrol\t1000")
	ctx := fixture.commitContext(vocab.InputMobileFuelLitres)

	result, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, parsed)
	if err != nil {
		t.Fatalf("commit mobile fuel: %v", err)
	}
	if len(result.ActivityRecordIDs) != 2 {
		t.Fatalf("activity record IDs = %#v, want 2 aggregated records", result.ActivityRecordIDs)
	}

	records := requireActivityRecords(t, fixture.st, fixture.periodID, 2)
	diesel := findRecordByFuel(t, records, "diesel")
	if !floatEqual(diesel.Amount, 4200) || diesel.Unit != "L" || diesel.Method != domain.ActivityMethodFuelBased {
		t.Fatalf("diesel record = %#v, want 4200 L fuel_based", diesel)
	}
	pasteRows := requirePasteRows(t, fixture.st, result.PasteBatchID, 3)
	if pasteRows[0].ActivityRecordID == nil || pasteRows[1].ActivityRecordID == nil || *pasteRows[0].ActivityRecordID != *pasteRows[1].ActivityRecordID {
		t.Fatalf("paste rows = %#v, want duplicate diesel rows pointing to same activity record", pasteRows)
	}

	badCtx := ctx
	badCtx.MobileMethod = domain.MobileMethodDistanceBased
	if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), badCtx, parsed); err == nil {
		t.Fatalf("distance_based context accepted mobile_fuel_litres")
	}
}

func TestCommitVehicleDistance(t *testing.T) {
	fixture := newCommitFixture(t, true)
	parsed := Parse(vocab.InputVehicleDistanceKm, "Car 1\tAA111AA\tCar\tSmall\tPetrol\t100\nVan 1\tBB222BB\tVan\tClass II\tDiesel\t200")
	ctx := fixture.commitContext(vocab.InputVehicleDistanceKm)

	_, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, parsed)
	if err != nil {
		t.Fatalf("commit vehicle distance: %v", err)
	}

	records := requireActivityRecords(t, fixture.st, fixture.periodID, 2)
	for _, record := range records {
		if record.Method != domain.ActivityMethodDistanceBased || record.Unit != "km" || record.ActivityType != "vehicle_distance" {
			t.Fatalf("record = %#v, want distance_based km vehicle_distance", record)
		}
	}

	badCtx := ctx
	badCtx.MobileMethod = domain.MobileMethodFuelBased
	if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), badCtx, parsed); err == nil {
		t.Fatalf("fuel_based context accepted vehicle_distance_km")
	}
}

func TestCommitRefrigerants(t *testing.T) {
	fixture := newCommitFixture(t, true)
	parsed := Parse(vocab.InputRefrigerantsAnnualKg, "R410A\t3.2\nr-410a\t1.0")

	result, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(vocab.InputRefrigerantsAnnualKg), parsed)
	if err != nil {
		t.Fatalf("commit refrigerants: %v", err)
	}
	if len(result.ActivityRecordIDs) != 1 {
		t.Fatalf("activity record IDs = %#v, want one aggregated record", result.ActivityRecordIDs)
	}

	records := requireActivityRecords(t, fixture.st, fixture.periodID, 1)
	record := records[0]
	if !floatEqual(record.Amount, 4.2) || record.Unit != "kg" || record.Method != domain.ActivityMethodDirectGWP || record.Vector != domain.ActivityVectorRefrigerants {
		t.Fatalf("record = %#v, want 4.2 kg direct_gwp refrigerants", record)
	}
	pasteRows := requirePasteRows(t, fixture.st, result.PasteBatchID, 2)
	if pasteRows[0].ActivityRecordID == nil || pasteRows[1].ActivityRecordID == nil || *pasteRows[0].ActivityRecordID != *pasteRows[1].ActivityRecordID {
		t.Fatalf("paste rows = %#v, want duplicate refrigerants pointing to same activity record", pasteRows)
	}
}

func TestParserErrorsBlockCommit(t *testing.T) {
	fixture := newCommitFixture(t, true)
	parsed := Parse(vocab.InputElectricityMonthlyKWh, "not-a-month\t100")

	if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(vocab.InputElectricityMonthlyKWh), parsed); err == nil {
		t.Fatalf("commit with parser errors succeeded")
	}
	assertNoBatchesOrRecords(t, fixture)
}

func TestExistingActiveDataCanBeReplaced(t *testing.T) {
	t.Run("monthly data replaces matching months", func(t *testing.T) {
		fixture := newCommitFixture(t, true)
		ctx := fixture.commitContext(vocab.InputElectricityMonthlyKWh)
		if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputElectricityMonthlyKWh, "January\t100\nFebruary\t200")); err != nil {
			t.Fatalf("first commit: %v", err)
		}
		if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputElectricityMonthlyKWh, "January\t300")); err != nil {
			t.Fatalf("replacement commit: %v", err)
		}

		facilityID := domain.ID(fixture.facilityID)
		active, err := fixture.st.ListActiveActivityRecordsByPeriodFacilitySource(fixture.periodID, &facilityID, domain.ActivitySourceKindElectricityMonthlyKWh)
		if err != nil {
			t.Fatalf("list active activity records: %v", err)
		}
		if len(active) != 2 {
			t.Fatalf("active records = %#v, want January replacement and original February", active)
		}
		amounts := map[time.Month]float64{}
		for _, record := range active {
			amounts[record.PeriodStart.Month()] = record.Amount
		}
		if !floatEqual(amounts[time.January], 300) || !floatEqual(amounts[time.February], 200) {
			t.Fatalf("active amounts = %#v, want January 300 and February 200", amounts)
		}
		records := requireActivityRecords(t, fixture.st, fixture.periodID, 3)
		superseded := 0
		for _, record := range records {
			if record.Status == domain.ActivityRecordStatusSuperseded {
				superseded++
			}
		}
		if superseded != 1 {
			t.Fatalf("superseded records = %d, want 1; records = %#v", superseded, records)
		}
	})

	t.Run("period data replaces existing vector data", func(t *testing.T) {
		fixture := newCommitFixture(t, true)
		ctx := fixture.commitContext(vocab.InputMobileFuelLitres)
		if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputMobileFuelLitres, "Diesel\t200\nPetrol\t50")); err != nil {
			t.Fatalf("first commit: %v", err)
		}
		if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputMobileFuelLitres, "Diesel\t300")); err != nil {
			t.Fatalf("replacement commit: %v", err)
		}

		active, err := fixture.st.ListActiveActivityRecordsByPeriodFacilitySource(fixture.periodID, nil, domain.ActivitySourceKindMobileFuelLitres)
		if err != nil {
			t.Fatalf("list active activity records: %v", err)
		}
		if len(active) != 1 || active[0].FuelType != "diesel" || !floatEqual(active[0].Amount, 300) {
			t.Fatalf("active records = %#v, want replacement diesel 300", active)
		}
		records := requireActivityRecords(t, fixture.st, fixture.periodID, 3)
		superseded := 0
		for _, record := range records {
			if record.Status == domain.ActivityRecordStatusSuperseded {
				superseded++
			}
		}
		if superseded != 2 {
			t.Fatalf("superseded records = %d, want 2; records = %#v", superseded, records)
		}
	})
}

func TestCommitHashesAreStable(t *testing.T) {
	fixtureA := newCommitFixture(t, true)
	fixtureB := newCommitFixture(t, true)
	rawA := "January\t100\nFebruary\t200"
	rawB := "January\t100\r\nFebruary\t200"

	resultA, err := CommitParsedInput(context.Background(), store.NewRepository(fixtureA.st), fixtureA.commitContext(vocab.InputElectricityMonthlyKWh), Parse(vocab.InputElectricityMonthlyKWh, rawA))
	if err != nil {
		t.Fatalf("commit A: %v", err)
	}
	resultB, err := CommitParsedInput(context.Background(), store.NewRepository(fixtureB.st), fixtureB.commitContext(vocab.InputElectricityMonthlyKWh), Parse(vocab.InputElectricityMonthlyKWh, rawB))
	if err != nil {
		t.Fatalf("commit B: %v", err)
	}

	batchA, err := fixtureA.st.GetPasteBatch(resultA.PasteBatchID)
	if err != nil {
		t.Fatalf("get batch A: %v", err)
	}
	batchB, err := fixtureB.st.GetPasteBatch(resultB.PasteBatchID)
	if err != nil {
		t.Fatalf("get batch B: %v", err)
	}
	if batchA.RawHash != batchB.RawHash || batchA.RawHash != RawHash(rawA) {
		t.Fatalf("raw hashes = %q and %q, want stable normalized raw hash", batchA.RawHash, batchB.RawHash)
	}

	recordsA := requireActivityRecords(t, fixtureA.st, fixtureA.periodID, 2)
	recordsB := requireActivityRecords(t, fixtureB.st, fixtureB.periodID, 2)
	if recordsA[0].SourceHash != recordsB[0].SourceHash || recordsA[1].SourceHash != recordsB[1].SourceHash {
		t.Fatalf("source hashes A=%#v B=%#v, want stable source hashes", recordsA, recordsB)
	}
}

func TestCommitRollsBackOnInsertFailure(t *testing.T) {
	fixture := newCommitFixture(t, false)
	parsed := Parse(vocab.InputElectricityMonthlyKWh, "January\t100")
	ctx := fixture.commitContext(vocab.InputElectricityMonthlyKWh)

	if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, parsed); err == nil {
		t.Fatalf("commit with missing facility succeeded")
	}
	assertNoBatchesOrRecords(t, fixture)
}

func TestStoredMobileMethodRejectsMismatch(t *testing.T) {
	fixture := newCommitFixture(t, true)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := fixture.st.UpsertReportingPeriodSettings(domain.ReportingPeriodSettings{
		ID:                "settings-1",
		OrganizationID:    fixture.organizationID,
		ReportingPeriodID: fixture.periodID,
		MobileMethod:      domain.MobileMethodDistanceBased,
		CreatedAt:         now,
		UpdatedAt:         now,
	}); err != nil {
		t.Fatalf("upsert reporting period settings: %v", err)
	}

	parsed := Parse(vocab.InputMobileFuelLitres, "Diesel\t100")
	if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(vocab.InputMobileFuelLitres), parsed); err == nil {
		t.Fatalf("stored distance_based method accepted mobile_fuel_litres")
	}
	assertNoBatchesOrRecords(t, fixture)
}

func TestCommitUsesTrustedStoredContext(t *testing.T) {
	t.Run("reparses raw text before writing", func(t *testing.T) {
		fixture := newCommitFixture(t, true)
		parsed := Parse(vocab.InputElectricityMonthlyKWh, "January\t100")
		parsed.Rows[0].Normalized["amount"] = "999"
		parsed.RowsTotal = 99
		parsed.RowsValid = 99

		result, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(vocab.InputElectricityMonthlyKWh), parsed)
		if err != nil {
			t.Fatalf("commit electricity: %v", err)
		}
		if result.RowsTotal != 1 || result.RowsValid != 1 {
			t.Fatalf("commit counts = %#v, want reparsed counts", result)
		}

		records := requireActivityRecords(t, fixture.st, fixture.periodID, 1)
		if !floatEqual(records[0].Amount, 100) {
			t.Fatalf("record amount = %v, want raw text amount 100", records[0].Amount)
		}
	})

	t.Run("derives reporting period dates", func(t *testing.T) {
		fixture := newCommitFixture(t, true)
		ctx := fixture.commitContext(vocab.InputElectricityMonthlyKWh)
		ctx.ReportingYear = 2030
		ctx.PeriodStart = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
		ctx.PeriodEnd = time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)

		if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputElectricityMonthlyKWh, "January\t100")); err != nil {
			t.Fatalf("commit electricity: %v", err)
		}

		records := requireActivityRecords(t, fixture.st, fixture.periodID, 1)
		wantStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		wantEnd := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		if !records[0].PeriodStart.Equal(wantStart) || !records[0].PeriodEnd.Equal(wantEnd) {
			t.Fatalf("record period = %s..%s, want %s..%s", records[0].PeriodStart, records[0].PeriodEnd, wantStart, wantEnd)
		}
	})

	t.Run("rejects mismatched organization", func(t *testing.T) {
		fixture := newCommitFixture(t, true)
		ctx := fixture.commitContext(vocab.InputElectricityMonthlyKWh)
		ctx.OrganizationID = "other-org"

		_, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputElectricityMonthlyKWh, "January\t100"))
		if !errors.Is(err, ErrInvalidCommit) {
			t.Fatalf("commit error = %v, want ErrInvalidCommit", err)
		}
		assertNoBatchesOrRecords(t, fixture)
	})

	t.Run("rejects cross organization facility", func(t *testing.T) {
		fixture := newCommitFixture(t, true)
		now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		otherOrganizationID := domain.ID("org-2")
		otherFacilityID := "facility-2"
		if err := fixture.st.CreateOrganization(domain.Organization{ID: otherOrganizationID, Name: "Other", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatalf("create other organization: %v", err)
		}
		if err := fixture.st.CreateFacility(domain.Facility{ID: domain.ID(otherFacilityID), OrganizationID: otherOrganizationID, Name: "Other Facility", CountryCode: "IT", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatalf("create other facility: %v", err)
		}

		ctx := fixture.commitContext(vocab.InputNaturalGasMonthlySmc)
		ctx.FacilityID = &otherFacilityID
		_, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), ctx, Parse(vocab.InputNaturalGasMonthlySmc, "January\t100"))
		if !errors.Is(err, ErrInvalidCommit) {
			t.Fatalf("commit error = %v, want ErrInvalidCommit", err)
		}
		assertNoBatchesOrRecords(t, fixture)
	})
}

func TestSavedDataDisplayQueries(t *testing.T) {
	tests := []struct {
		name       string
		kind       vocab.InputKind
		raw        string
		sourceKind domain.ActivitySourceKind
		facility   bool
		want       int
	}{
		{"electricity", vocab.InputElectricityMonthlyKWh, "January\t100\nFebruary\t200", domain.ActivitySourceKindElectricityMonthlyKWh, true, 2},
		{"natural gas", vocab.InputNaturalGasMonthlySmc, "January\t30\nFebruary\t40", domain.ActivitySourceKindNaturalGasMonthlySMC, true, 2},
		{"mobile fuel", vocab.InputMobileFuelLitres, "Diesel\t200\nPetrol\t50", domain.ActivitySourceKindMobileFuelLitres, false, 2},
		{"vehicle distance", vocab.InputVehicleDistanceKm, "Fiat Panda\tAB123CD\tCar\tSmall\tPetrol\t1000\nFiat Doblo\tCD456EF\tVan\tClass II\tDiesel\t2000", domain.ActivitySourceKindVehicleDistanceKM, false, 2},
		{"refrigerants", vocab.InputRefrigerantsAnnualKg, "R410A\t1.5\nR134a\t0.5", domain.ActivitySourceKindRefrigerantsAnnualKG, true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newCommitFixture(t, true)
			parsed := Parse(tt.kind, tt.raw)
			if _, err := CommitParsedInput(context.Background(), store.NewRepository(fixture.st), fixture.commitContext(tt.kind), parsed); err != nil {
				t.Fatalf("commit %s: %v", tt.name, err)
			}

			var facilityID *domain.ID
			if tt.facility {
				id := domain.ID(fixture.facilityID)
				facilityID = &id
			}
			records, err := fixture.st.ListActiveActivityRecordsByPeriodFacilitySource(fixture.periodID, facilityID, tt.sourceKind)
			if err != nil {
				t.Fatalf("list saved records: %v", err)
			}
			if len(records) != tt.want {
				t.Fatalf("saved records count = %d, want %d; records = %#v", len(records), tt.want, records)
			}
		})
	}
}

type commitFixture struct {
	st             *store.Store
	organizationID domain.ID
	periodID       domain.ID
	facilityID     string
	periodStart    time.Time
	periodEnd      time.Time
}

func newCommitFixture(t *testing.T, createFacility bool) commitFixture {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "commit.sqlite"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	st := store.New(db)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	organizationID := domain.ID("org-1")
	periodID := domain.ID("period-2026")
	facilityID := "facility-1"
	periodStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	if err := st.CreateOrganization(domain.Organization{ID: organizationID, Name: "Acme", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if createFacility {
		if err := st.CreateFacility(domain.Facility{ID: facilityID, OrganizationID: organizationID, Name: "Milan", CountryCode: "IT", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatalf("create facility: %v", err)
		}
	}
	if err := st.CreateReportingPeriod(domain.ReportingPeriod{
		ID:             periodID,
		OrganizationID: organizationID,
		Year:           2026,
		StartsOn:       periodStart,
		EndsOn:         periodEnd,
		Status:         domain.ReportingPeriodStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create reporting period: %v", err)
	}

	return commitFixture{
		st:             st,
		organizationID: organizationID,
		periodID:       periodID,
		facilityID:     facilityID,
		periodStart:    periodStart,
		periodEnd:      periodEnd,
	}
}

func (f commitFixture) commitContext(kind vocab.InputKind) CommitContext {
	ctx := CommitContext{
		OrganizationID:    f.organizationID,
		ReportingPeriodID: f.periodID,
		ReportingYear:     2026,
		PeriodStart:       f.periodStart,
		PeriodEnd:         f.periodEnd,
		InputKind:         kind,
	}
	switch kind {
	case vocab.InputElectricityMonthlyKWh:
		ctx.FacilityID = &f.facilityID
		ctx.GOCoverage = domain.GOCoverageNone
	case vocab.InputNaturalGasMonthlySmc, vocab.InputRefrigerantsAnnualKg:
		ctx.FacilityID = &f.facilityID
	case vocab.InputMobileFuelLitres:
		ctx.MobileMethod = domain.MobileMethodFuelBased
	case vocab.InputVehicleDistanceKm:
		ctx.MobileMethod = domain.MobileMethodDistanceBased
	}
	return ctx
}

func requirePasteRows(t *testing.T, st *store.Store, batchID domain.ID, want int) []domain.PasteRow {
	t.Helper()
	rows, err := st.ListPasteRowsByBatch(batchID)
	if err != nil {
		t.Fatalf("list paste rows: %v", err)
	}
	if len(rows) != want {
		t.Fatalf("paste rows count = %d, want %d; rows = %#v", len(rows), want, rows)
	}
	return rows
}

func requireActivityRecords(t *testing.T, st *store.Store, periodID domain.ID, want int) []domain.ActivityRecord {
	t.Helper()
	records, err := st.ListActivityRecordsByPeriod(periodID)
	if err != nil {
		t.Fatalf("list activity records: %v", err)
	}
	if len(records) != want {
		t.Fatalf("activity records count = %d, want %d; records = %#v", len(records), want, records)
	}
	return records
}

func findRecordByFuel(t *testing.T, records []domain.ActivityRecord, fuelType string) domain.ActivityRecord {
	t.Helper()
	for _, record := range records {
		if record.FuelType == fuelType {
			return record
		}
	}
	t.Fatalf("records = %#v, want fuel type %q", records, fuelType)
	return domain.ActivityRecord{}
}

func assertNoBatchesOrRecords(t *testing.T, fixture commitFixture) {
	t.Helper()
	batches, err := fixture.st.ListPasteBatchesByPeriod(fixture.periodID)
	if err != nil {
		t.Fatalf("list paste batches: %v", err)
	}
	if len(batches) != 0 {
		t.Fatalf("paste batches = %#v, want none", batches)
	}
	records, err := fixture.st.ListActivityRecordsByPeriod(fixture.periodID)
	if err != nil {
		t.Fatalf("list activity records: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("activity records = %#v, want none", records)
	}
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0000001
}
