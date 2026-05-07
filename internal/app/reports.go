package app

import (
	"context"

	"ghgo/internal/ports"
	"ghgo/internal/report"
)

type ReportService struct {
	store ports.Store
}

type BuildReportTablesOptions struct {
	CalculationRunID string
}

func (s *ReportService) BuildTables(ctx context.Context, opts BuildReportTablesOptions) (*report.ReportTables, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	if cleanText(opts.CalculationRunID) == "" {
		return nil, invalidOptions("calculation run id is required")
	}
	return report.NewBuilder(s.store).BuildTables(ctx, report.BuildOptions{CalculationRunID: cleanText(opts.CalculationRunID)})
}
