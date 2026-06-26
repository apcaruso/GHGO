package report

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"ghgo/internal/domain"
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

type Builder struct {
	Store *store.Store
}

type BuildOptions struct {
	CalculationRunID string
}

func NewBuilder(st *store.Store) *Builder {
	return &Builder{Store: st}
}

type ReportTables struct {
	CalculationRunID  string
	OrganizationID    string
	ReportingPeriodID string
	FactorSetID       string

	ExecutiveSummary ExecutiveSummaryTable
	MonthlyEmissions MonthlyEmissionsTable

	ElectricityDetail  ElectricityDetailTable
	NaturalGasDetail   NaturalGasDetailTable
	MobileDetail       MobileDetailTable
	RefrigerantsDetail RefrigerantsDetailTable

	ScopeSummary  ScopeSummaryTable
	VectorSummary VectorSummaryTable
	Methodology   MethodologyTable

	ValidationNotes ValidationNotesTable
}

type reportData struct {
	run       *domain.CalculationRun
	factorSet *domain.FactorSet
	rows      []store.ReportResultRow

	hasLocationBasedComparison bool
	scope1PrimaryKgCO2e        float64
	scope2PrimaryKgCO2e        float64
	scope2LocationBasedKgCO2e  float64
	primaryTotalKgCO2e         float64
	locationBasedTotalKgCO2e   float64
	electricityPrimaryMethod   domain.ActivityMethod
}

func (b *Builder) BuildTables(ctx context.Context, opts BuildOptions) (*ReportTables, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if b == nil || b.Store == nil {
		return nil, invalidOptions("store is required")
	}
	if opts.CalculationRunID == "" {
		return nil, invalidOptions("calculation run id is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	run, err := b.Store.GetCalculationRun(domain.ID(opts.CalculationRunID))
	if errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("calculation run %q: %w", opts.CalculationRunID, store.ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	factorSet, err := b.Store.GetFactorSet(run.FactorSetID)
	if err != nil {
		return nil, err
	}
	rows, err := b.Store.ListReportResultRows(ctx, run.ID)
	if err != nil {
		return nil, err
	}

	data := newReportData(run, factorSet, rows)
	electricityDetail := buildElectricityDetail(data)
	naturalGasDetail := buildNaturalGasDetail(data)
	mobileDetail, err := buildMobileDetail(data)
	if err != nil {
		return nil, err
	}
	refrigerantsDetail := buildRefrigerantsDetail(data)

	return &ReportTables{
		CalculationRunID:   run.ID,
		OrganizationID:     run.OrganizationID,
		ReportingPeriodID:  run.ReportingPeriodID,
		FactorSetID:        run.FactorSetID,
		ExecutiveSummary:   buildExecutiveSummary(data),
		MonthlyEmissions:   buildMonthlyEmissions(data),
		ElectricityDetail:  electricityDetail,
		NaturalGasDetail:   naturalGasDetail,
		MobileDetail:       mobileDetail,
		RefrigerantsDetail: refrigerantsDetail,
		ScopeSummary:       buildScopeSummary(data),
		VectorSummary:      buildVectorSummary(data),
		Methodology:        buildMethodology(data),
		ValidationNotes:    buildValidationNotes(data),
	}, nil
}

func newReportData(run *domain.CalculationRun, factorSet *domain.FactorSet, rows []store.ReportResultRow) reportData {
	data := reportData{
		run:                      run,
		factorSet:                factorSet,
		rows:                     rows,
		electricityPrimaryMethod: domain.ActivityMethodLocationBased,
	}
	for _, row := range rows {
		result := row.CalculationResult
		if result.IsPrimary {
			data.primaryTotalKgCO2e += result.EmissionsKgCO2e
			switch result.Scope {
			case domain.Scope1:
				data.scope1PrimaryKgCO2e += result.EmissionsKgCO2e
			case domain.Scope2:
				data.scope2PrimaryKgCO2e += result.EmissionsKgCO2e
			}
		}
		if result.Scope == domain.Scope2 && result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodLocationBased {
			data.scope2LocationBasedKgCO2e += result.EmissionsKgCO2e
		}
		if result.IsPrimary && result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodMarketBased {
			data.hasLocationBasedComparison = true
			data.electricityPrimaryMethod = domain.ActivityMethodMarketBased
		}
	}
	if data.hasLocationBasedComparison {
		data.locationBasedTotalKgCO2e = data.scope1PrimaryKgCO2e + data.scope2LocationBasedKgCO2e
	}
	return data
}

func monthNumber(record domain.ActivityRecord) int {
	return int(record.PeriodStart.Month())
}

func monthName(month int) string {
	return vocab.Month(month).EnglishName()
}

func facilityID(record domain.ActivityRecord) string {
	if record.FacilityID == nil {
		return ""
	}
	return *record.FacilityID
}

func kgToT(value float64) float64 {
	return value / 1000
}

func floatPtr(value float64) *float64 {
	return &value
}

func share(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return value / total
}

func activitySummary(value float64, unit string) string {
	return strconv.FormatFloat(value, 'f', -1, 64) + " " + unit
}
