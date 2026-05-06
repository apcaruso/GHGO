package store

import (
	"path/filepath"
	"testing"
	"time"

	"ghgo/internal/domain"
)

func TestStoreCoreEntities(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	organization := domain.Organization{
		ID:        "org-1",
		Name:      "Acme Ltd",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateOrganization(organization); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	gotOrganization, err := store.GetOrganization(organization.ID)
	if err != nil {
		t.Fatalf("get organization: %v", err)
	}
	if gotOrganization.Name != organization.Name {
		t.Fatalf("organization name = %q, want %q", gotOrganization.Name, organization.Name)
	}
	organizations, err := store.ListOrganizations()
	if err != nil {
		t.Fatalf("list organizations: %v", err)
	}
	if len(organizations) != 1 {
		t.Fatalf("organizations count = %d, want 1", len(organizations))
	}

	facility := domain.Facility{
		ID:             "facility-1",
		OrganizationID: organization.ID,
		Name:           "Milan Office",
		CountryCode:    "IT",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := store.CreateFacility(facility); err != nil {
		t.Fatalf("create facility: %v", err)
	}
	facilities, err := store.ListFacilitiesByOrganization(organization.ID)
	if err != nil {
		t.Fatalf("list facilities: %v", err)
	}
	if len(facilities) != 1 || facilities[0].ID != facility.ID {
		t.Fatalf("facilities = %#v, want facility %q", facilities, facility.ID)
	}

	period := domain.ReportingPeriod{
		ID:             "period-2026",
		OrganizationID: organization.ID,
		Year:           2026,
		StartsOn:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndsOn:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:         domain.ReportingPeriodStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := store.CreateReportingPeriod(period); err != nil {
		t.Fatalf("create reporting period: %v", err)
	}
	gotPeriod, err := store.GetReportingPeriod(period.ID)
	if err != nil {
		t.Fatalf("get reporting period: %v", err)
	}
	if gotPeriod.Year != period.Year {
		t.Fatalf("reporting period year = %d, want %d", gotPeriod.Year, period.Year)
	}
	periods, err := store.ListReportingPeriodsByOrganization(organization.ID)
	if err != nil {
		t.Fatalf("list reporting periods: %v", err)
	}
	if len(periods) != 1 || periods[0].ID != period.ID {
		t.Fatalf("periods = %#v, want period %q", periods, period.ID)
	}

	periodSettings := domain.ReportingPeriodSettings{
		ID:                "settings-1",
		OrganizationID:    organization.ID,
		ReportingPeriodID: period.ID,
		MobileMethod:      domain.MobileMethodFuelBased,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := store.UpsertReportingPeriodSettings(periodSettings); err != nil {
		t.Fatalf("upsert reporting period settings: %v", err)
	}
	gotPeriodSettings, err := store.GetReportingPeriodSettings(period.ID)
	if err != nil {
		t.Fatalf("get reporting period settings: %v", err)
	}
	if gotPeriodSettings.MobileMethod != domain.MobileMethodFuelBased {
		t.Fatalf("mobile method = %q, want %q", gotPeriodSettings.MobileMethod, domain.MobileMethodFuelBased)
	}

	goCancelledAt := now.Add(time.Hour)
	evidenceFileID := domain.ID("evidence-1")
	electricitySettings := domain.ElectricitySettings{
		ID:                    "electricity-settings-1",
		OrganizationID:        organization.ID,
		ReportingPeriodID:     period.ID,
		FacilityID:            facility.ID,
		HasGuaranteesOfOrigin: true,
		GOCoverage:            domain.GOCoverageFull,
		GOReference:           "GO-123",
		GOMarket:              "IT",
		GOCancelledAt:         &goCancelledAt,
		EvidenceFileID:        &evidenceFileID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := store.UpsertElectricitySettings(electricitySettings); err != nil {
		t.Fatalf("upsert electricity settings: %v", err)
	}
	electricitySettings.GOReference = "GO-456"
	electricitySettings.UpdatedAt = now.Add(2 * time.Hour)
	if err := store.UpsertElectricitySettings(electricitySettings); err != nil {
		t.Fatalf("upsert electricity settings second time: %v", err)
	}
	gotElectricitySettings, err := store.GetElectricitySettings(period.ID, facility.ID)
	if err != nil {
		t.Fatalf("get electricity settings: %v", err)
	}
	if gotElectricitySettings.GOReference != "GO-456" || !gotElectricitySettings.HasGuaranteesOfOrigin {
		t.Fatalf("electricity settings = %#v, want updated GO reference", gotElectricitySettings)
	}

	activityFacilityID := domain.ID(facility.ID)
	activeRecord := domain.ActivityRecord{
		ID:                "activity-active",
		OrganizationID:    organization.ID,
		FacilityID:        &activityFacilityID,
		ReportingPeriodID: period.ID,
		SourceKind:        domain.ActivitySourceKindElectricityMonthlyKWh,
		Scope:             domain.Scope2,
		Vector:            domain.ActivityVectorElectricity,
		Category:          "purchased electricity",
		Method:            domain.ActivityMethodLocationBased,
		ActivityType:      "electricity",
		PeriodStart:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:         time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		Amount:            100,
		Unit:              "kWh",
		Status:            domain.ActivityRecordStatusActive,
		SourceHash:        "activity-hash-1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := store.CreateActivityRecord(activeRecord); err != nil {
		t.Fatalf("create active activity record: %v", err)
	}
	supersededRecord := activeRecord
	supersededRecord.ID = "activity-superseded"
	supersededRecord.Status = domain.ActivityRecordStatusSuperseded
	supersededRecord.SourceHash = "activity-hash-2"
	if err := store.CreateActivityRecord(supersededRecord); err != nil {
		t.Fatalf("create superseded activity record: %v", err)
	}
	allRecords, err := store.ListActivityRecordsByPeriod(period.ID)
	if err != nil {
		t.Fatalf("list activity records: %v", err)
	}
	if len(allRecords) != 2 {
		t.Fatalf("activity records count = %d, want 2", len(allRecords))
	}
	activeRecords, err := store.ListActiveActivityRecordsByPeriod(period.ID)
	if err != nil {
		t.Fatalf("list active activity records: %v", err)
	}
	if len(activeRecords) != 1 || activeRecords[0].ID != activeRecord.ID {
		t.Fatalf("active records = %#v, want %q", activeRecords, activeRecord.ID)
	}

	factorSet := domain.FactorSet{
		ID:           "factor-set-1",
		Name:         "Example Factors",
		Source:       "example",
		Year:         2026,
		Version:      "v1",
		ImportedAt:   now,
		MetadataJSON: `{}`,
	}
	if err := store.CreateFactorSet(factorSet); err != nil {
		t.Fatalf("create factor set: %v", err)
	}
	gotFactorSet, err := store.GetFactorSet(factorSet.ID)
	if err != nil {
		t.Fatalf("get factor set: %v", err)
	}
	if gotFactorSet.Source != factorSet.Source {
		t.Fatalf("factor set source = %q, want %q", gotFactorSet.Source, factorSet.Source)
	}
	factorSets, err := store.ListFactorSets()
	if err != nil {
		t.Fatalf("list factor sets: %v", err)
	}
	if len(factorSets) != 1 {
		t.Fatalf("factor sets count = %d, want 1", len(factorSets))
	}

	emissionFactor := domain.EmissionFactor{
		ID:           "factor-1",
		FactorSetID:  factorSet.ID,
		Source:       "example",
		Scope:        domain.Scope2,
		Level1:       "electricity",
		InputUnit:    "kWh",
		FactorUnit:   "kgCO2e/kWh",
		GHG:          "CO2e",
		FactorValue:  0.2,
		MetadataJSON: `{}`,
	}
	if err := store.CreateEmissionFactor(emissionFactor); err != nil {
		t.Fatalf("create emission factor: %v", err)
	}
	emissionFactors, err := store.ListEmissionFactorsBySet(factorSet.ID)
	if err != nil {
		t.Fatalf("list emission factors: %v", err)
	}
	if len(emissionFactors) != 1 || emissionFactors[0].ID != emissionFactor.ID {
		t.Fatalf("emission factors = %#v, want %q", emissionFactors, emissionFactor.ID)
	}

	calculationRun := domain.CalculationRun{
		ID:                   "calculation-run-1",
		OrganizationID:       organization.ID,
		ReportingPeriodID:    period.ID,
		FactorSetID:          factorSet.ID,
		Status:               domain.CalculationRunStatusRunning,
		StartedAt:            now,
		SettingsSnapshotJSON: `{}`,
	}
	if err := store.CreateCalculationRun(calculationRun); err != nil {
		t.Fatalf("create calculation run: %v", err)
	}
	completedAt := now.Add(3 * time.Hour)
	if err := store.CompleteCalculationRun(calculationRun.ID, completedAt); err != nil {
		t.Fatalf("complete calculation run: %v", err)
	}
	gotCalculationRun, err := store.GetCalculationRun(calculationRun.ID)
	if err != nil {
		t.Fatalf("get calculation run: %v", err)
	}
	if gotCalculationRun.Status != domain.CalculationRunStatusCompleted || gotCalculationRun.CompletedAt == nil {
		t.Fatalf("calculation run = %#v, want completed", gotCalculationRun)
	}

	factorID := domain.ID(emissionFactor.ID)
	calculationResult := domain.CalculationResult{
		ID:               "calculation-result-1",
		CalculationRunID: calculationRun.ID,
		ActivityRecordID: activeRecord.ID,
		Scope:            domain.Scope2,
		Vector:           domain.ActivityVectorElectricity,
		Method:           domain.ActivityMethodLocationBased,
		ActivityAmount:   100,
		ActivityUnit:     "kWh",
		FactorID:         &factorID,
		FactorValue:      0.2,
		FactorUnit:       "kgCO2e/kWh",
		FactorSource:     "example",
		EmissionsKgCO2e:  20,
		IsPrimary:        true,
		NotesJSON:        `{}`,
	}
	if err := store.CreateCalculationResult(calculationResult); err != nil {
		t.Fatalf("create calculation result: %v", err)
	}
	calculationResults, err := store.ListCalculationResultsByRun(calculationRun.ID)
	if err != nil {
		t.Fatalf("list calculation results: %v", err)
	}
	if len(calculationResults) != 1 || !calculationResults[0].IsPrimary {
		t.Fatalf("calculation results = %#v, want one primary result", calculationResults)
	}

	pasteBatch := domain.PasteBatch{
		ID:                "paste-batch-1",
		OrganizationID:    organization.ID,
		ReportingPeriodID: period.ID,
		InputKind:         "electricity",
		ContextJSON:       `{}`,
		RawText:           "raw",
		RawHash:           "paste-hash-1",
		Status:            domain.PasteBatchStatusParsed,
		RowsTotal:         1,
		RowsValid:         1,
		RowsError:         0,
		CreatedAt:         now,
	}
	if err := store.CreatePasteBatch(pasteBatch); err != nil {
		t.Fatalf("create paste batch: %v", err)
	}
	gotPasteBatch, err := store.GetPasteBatch(pasteBatch.ID)
	if err != nil {
		t.Fatalf("get paste batch: %v", err)
	}
	if gotPasteBatch.RawHash != pasteBatch.RawHash {
		t.Fatalf("paste batch hash = %q, want %q", gotPasteBatch.RawHash, pasteBatch.RawHash)
	}
	activityRecordID := domain.ID(activeRecord.ID)
	pasteRow := domain.PasteRow{
		ID:               "paste-row-1",
		PasteBatchID:     pasteBatch.ID,
		RowNumber:        1,
		RawJSON:          `{}`,
		NormalizedJSON:   `{}`,
		Status:           domain.PasteRowStatusCommitted,
		ErrorsJSON:       `[]`,
		WarningsJSON:     `[]`,
		ActivityRecordID: &activityRecordID,
	}
	if err := store.CreatePasteRow(pasteRow); err != nil {
		t.Fatalf("create paste row: %v", err)
	}
	pasteRows, err := store.ListPasteRowsByBatch(pasteBatch.ID)
	if err != nil {
		t.Fatalf("list paste rows: %v", err)
	}
	if len(pasteRows) != 1 || pasteRows[0].ActivityRecordID == nil || *pasteRows[0].ActivityRecordID != activeRecord.ID {
		t.Fatalf("paste rows = %#v, want linked activity record", pasteRows)
	}

	auditEvent := domain.AuditEvent{
		ID:             "audit-1",
		OrganizationID: organization.ID,
		EntityType:     "activity_record",
		EntityID:       activeRecord.ID,
		Action:         "create",
		PayloadJSON:    `{}`,
		CreatedAt:      now,
	}
	if err := store.CreateAuditEvent(auditEvent); err != nil {
		t.Fatalf("create audit event: %v", err)
	}
	auditEvents, err := store.ListAuditEventsByEntity(auditEvent.EntityType, auditEvent.EntityID)
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(auditEvents) != 1 || auditEvents[0].ID != auditEvent.ID {
		t.Fatalf("audit events = %#v, want %q", auditEvents, auditEvent.ID)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := Open(filepath.Join(t.TempDir(), "store.sqlite"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return New(db)
}
