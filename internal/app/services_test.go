package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/ports"
	"ghgo/internal/vocab"
)

func TestOpenSQLiteInitializesBackend(t *testing.T) {
	backend, err := OpenSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open sqlite backend: %v", err)
	}
	defer backend.Close()

	if backend.Store == nil || backend.Services == nil {
		t.Fatalf("backend = %#v, want store and services", backend)
	}
	if backend.DefaultFactorSet == nil || backend.DefaultFactorSet.ID != defaultFactorSetID {
		t.Fatalf("default factor set = %#v, want %q", backend.DefaultFactorSet, defaultFactorSetID)
	}
}

func TestServicesEndToEndNaturalGasWorkflow(t *testing.T) {
	ctx := context.Background()
	backend, err := OpenSQLite(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open sqlite backend: %v", err)
	}
	defer backend.Close()

	services := backend.Services
	organization, err := services.Organizations.Create(ctx, CreateOrganizationOptions{Name: "Acme Ltd"})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	facility, err := services.Facilities.Create(ctx, CreateFacilityOptions{
		OrganizationID: organization.ID,
		Name:           "Milan Office",
		CountryCode:    "it",
	})
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}
	period, err := services.ReportingPeriods.Create(ctx, CreateReportingPeriodOptions{
		OrganizationID: organization.ID,
		Year:           2026,
	})
	if err != nil {
		t.Fatalf("create reporting period: %v", err)
	}

	facilityID := string(facility.ID)
	commitResult, err := services.Inputs.ParseAndCommit(ctx, ParseAndCommitInputOptions{
		Context: input.CommitContext{
			OrganizationID:    organization.ID,
			ReportingPeriodID: period.ID,
			FacilityID:        &facilityID,
			ReportingYear:     period.Year,
			PeriodStart:       period.StartsOn,
			PeriodEnd:         period.EndsOn,
			InputKind:         vocab.InputNaturalGasMonthlySmc,
		},
		RawText: "January\t100",
	})
	if err != nil {
		t.Fatalf("parse and commit input: %v", err)
	}
	if !commitResult.Committed || len(commitResult.ActivityRecordIDs) != 1 {
		t.Fatalf("commit result = %#v, want one committed activity record", commitResult)
	}

	calculation, err := services.Calculations.Run(ctx, RunCalculationOptions{
		OrganizationID:    organization.ID,
		ReportingPeriodID: period.ID,
	})
	if err != nil {
		t.Fatalf("run calculation: %v", err)
	}
	if calculation.PrimaryTotalKgCO2e <= 0 {
		t.Fatalf("primary total = %v, want positive emissions", calculation.PrimaryTotalKgCO2e)
	}

	tables, err := services.Reports.BuildTables(ctx, BuildReportTablesOptions{CalculationRunID: calculation.CalculationRunID})
	if err != nil {
		t.Fatalf("build report tables: %v", err)
	}
	if tables.ExecutiveSummary.PrimaryTotalKgCO2e != calculation.PrimaryTotalKgCO2e {
		t.Fatalf("report total = %v, want calculation total %v", tables.ExecutiveSummary.PrimaryTotalKgCO2e, calculation.PrimaryTotalKgCO2e)
	}
}

func TestReportingPeriodSettingsServices(t *testing.T) {
	ctx := context.Background()
	backend, err := OpenSQLite(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open sqlite backend: %v", err)
	}
	defer backend.Close()

	services := backend.Services
	organization, err := services.Organizations.Create(ctx, CreateOrganizationOptions{Name: "Acme Ltd"})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	facility, err := services.Facilities.Create(ctx, CreateFacilityOptions{OrganizationID: organization.ID, Name: "Milan Office", CountryCode: "IT"})
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}
	period, err := services.ReportingPeriods.Create(ctx, CreateReportingPeriodOptions{OrganizationID: organization.ID, Year: 2026})
	if err != nil {
		t.Fatalf("create reporting period: %v", err)
	}

	periodSettings, err := services.ReportingPeriods.UpsertSettings(ctx, UpsertReportingPeriodSettingsOptions{
		ReportingPeriodID: period.ID,
		MobileMethod:      domain.MobileMethodDistanceBased,
	})
	if err != nil {
		t.Fatalf("upsert reporting period settings: %v", err)
	}
	if periodSettings.OrganizationID != organization.ID || periodSettings.MobileMethod != domain.MobileMethodDistanceBased {
		t.Fatalf("period settings = %#v, want derived organization and distance method", periodSettings)
	}

	cancelledAt := time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC)
	electricitySettings, err := services.ReportingPeriods.UpsertElectricitySettings(ctx, UpsertElectricitySettingsOptions{
		ReportingPeriodID:     period.ID,
		FacilityID:            facility.ID,
		HasGuaranteesOfOrigin: true,
		GOReference:           "GO-123",
		GOMarket:              "IT",
		GOCancelledAt:         &cancelledAt,
	})
	if err != nil {
		t.Fatalf("upsert electricity settings: %v", err)
	}
	if electricitySettings.OrganizationID != organization.ID || electricitySettings.GOCoverage != domain.GOCoverageFull {
		t.Fatalf("electricity settings = %#v, want derived organization and full GO coverage", electricitySettings)
	}

	if _, err := services.ReportingPeriods.UpsertSettings(ctx, UpsertReportingPeriodSettingsOptions{ReportingPeriodID: period.ID}); !errors.Is(err, ErrInvalidOptions) {
		t.Fatalf("invalid mobile method error = %v, want ErrInvalidOptions", err)
	}
}

func TestServiceValidation(t *testing.T) {
	backend, err := OpenSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open sqlite backend: %v", err)
	}
	defer backend.Close()

	_, err = backend.Services.Organizations.Create(context.Background(), CreateOrganizationOptions{Name: ""})
	if !errors.Is(err, ErrInvalidOptions) {
		t.Fatalf("create organization error = %v, want ErrInvalidOptions", err)
	}

	_, err = backend.Services.Organizations.Get(context.Background(), "missing")
	if !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("get missing organization error = %v, want ports.ErrNotFound", err)
	}
}

func TestServicesPropagateCanceledContext(t *testing.T) {
	backend, err := OpenSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open sqlite backend: %v", err)
	}
	defer backend.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = backend.Services.Organizations.List(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("list organizations error = %v, want context.Canceled", err)
	}
}
