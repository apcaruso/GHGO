package app

import (
	"context"

	"ghgo/internal/calc"
	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/ports"
)

type CalculationService struct {
	store ports.Store
}

type RunCalculationOptions struct {
	OrganizationID    string
	ReportingPeriodID string
	FactorSetID       string
}

func (s *CalculationService) Run(ctx context.Context, opts RunCalculationOptions) (calc.RunResult, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return calc.RunResult{}, err
	}

	organizationID, err := requiredID("organization id", opts.OrganizationID)
	if err != nil {
		return calc.RunResult{}, err
	}
	reportingPeriodID, err := requiredID("reporting period id", opts.ReportingPeriodID)
	if err != nil {
		return calc.RunResult{}, err
	}
	factorSetID := domain.ID(cleanText(opts.FactorSetID))
	if factorSetID == "" {
		factorSet, err := (&FactorService{store: s.store}).Default(ctx)
		if err != nil {
			return calc.RunResult{}, err
		}
		factorSetID = factorSet.ID
	}

	lookup := factors.NewLookup(s.store, string(factorSetID))
	engine := calc.NewEngine(s.store, lookup)
	return engine.Run(ctx, calc.RunOptions{
		OrganizationID:    string(organizationID),
		ReportingPeriodID: string(reportingPeriodID),
		FactorSetID:       string(factorSetID),
	})
}

func (s *CalculationService) GetRun(ctx context.Context, id string) (*domain.CalculationRun, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	runID, err := requiredID("calculation run id", id)
	if err != nil {
		return nil, err
	}
	return s.store.GetCalculationRun(ctx, runID)
}

func (s *CalculationService) ListRunsByPeriod(ctx context.Context, reportingPeriodID string) ([]domain.CalculationRun, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	periodID, err := requiredID("reporting period id", reportingPeriodID)
	if err != nil {
		return nil, err
	}
	return s.store.ListCalculationRunsByPeriod(ctx, periodID)
}
