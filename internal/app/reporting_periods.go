package app

import (
	"context"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/ports"
)

type ReportingPeriodService struct {
	store ports.Store
}

type CreateReportingPeriodOptions struct {
	ID             string
	OrganizationID string
	Year           int
	StartsOn       time.Time
	EndsOn         time.Time
	Status         domain.ReportingPeriodStatus
}

type UpsertReportingPeriodSettingsOptions struct {
	ID                string
	OrganizationID    string
	ReportingPeriodID string
	MobileMethod      domain.MobileMethod
}

type UpsertElectricitySettingsOptions struct {
	ID                    string
	OrganizationID        string
	ReportingPeriodID     string
	FacilityID            string
	HasGuaranteesOfOrigin bool
	GOCoverage            domain.GOCoverage
	GOReference           string
	GOMarket              string
	GOCancelledAt         *time.Time
	EvidenceFileID        string
}

func (s *ReportingPeriodService) Create(ctx context.Context, opts CreateReportingPeriodOptions) (*domain.ReportingPeriod, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}

	organizationID, err := requiredID("organization id", opts.OrganizationID)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.GetOrganization(ctx, organizationID); err != nil {
		return nil, err
	}
	if opts.Year <= 0 {
		return nil, invalidOptions("reporting year must be positive")
	}

	startsOn := opts.StartsOn
	endsOn := opts.EndsOn
	if startsOn.IsZero() && endsOn.IsZero() {
		startsOn, endsOn = yearBounds(opts.Year)
	} else if startsOn.IsZero() || endsOn.IsZero() {
		return nil, invalidOptions("reporting period start and end must both be provided")
	}
	if startsOn.After(endsOn) {
		return nil, invalidOptions("reporting period start is after end")
	}

	status := opts.Status
	if status == "" {
		status = domain.ReportingPeriodStatusDraft
	}
	if !status.Valid() {
		return nil, invalidOptions("invalid reporting period status %q", status)
	}

	id := domain.ID(cleanText(opts.ID))
	if id == "" {
		id, err = newID("reporting_period")
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	period := domain.ReportingPeriod{
		ID:             id,
		OrganizationID: organizationID,
		Year:           opts.Year,
		StartsOn:       startsOn.UTC(),
		EndsOn:         endsOn.UTC(),
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.store.CreateReportingPeriod(ctx, period); err != nil {
		return nil, err
	}
	return &period, nil
}

func (s *ReportingPeriodService) Get(ctx context.Context, id string) (*domain.ReportingPeriod, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	periodID, err := requiredID("reporting period id", id)
	if err != nil {
		return nil, err
	}
	return s.store.GetReportingPeriod(ctx, periodID)
}

func (s *ReportingPeriodService) ListByOrganization(ctx context.Context, organizationID string) ([]domain.ReportingPeriod, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	id, err := requiredID("organization id", organizationID)
	if err != nil {
		return nil, err
	}
	return s.store.ListReportingPeriodsByOrganization(ctx, id)
}

func (s *ReportingPeriodService) UpsertSettings(ctx context.Context, opts UpsertReportingPeriodSettingsOptions) (*domain.ReportingPeriodSettings, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}

	reportingPeriodID, err := requiredID("reporting period id", opts.ReportingPeriodID)
	if err != nil {
		return nil, err
	}
	period, err := s.store.GetReportingPeriod(ctx, reportingPeriodID)
	if err != nil {
		return nil, err
	}

	organizationID := domain.ID(cleanText(opts.OrganizationID))
	if organizationID == "" {
		organizationID = period.OrganizationID
	}
	if organizationID != period.OrganizationID {
		return nil, invalidOptions("reporting period belongs to organization %q, not %q", period.OrganizationID, organizationID)
	}
	if !opts.MobileMethod.Valid() {
		return nil, invalidOptions("invalid mobile method %q", opts.MobileMethod)
	}

	id := domain.ID(cleanText(opts.ID))
	if id == "" {
		id = domain.ID("reporting_period_settings_" + reportingPeriodID)
	}

	now := time.Now().UTC()
	settings := domain.ReportingPeriodSettings{
		ID:                id,
		OrganizationID:    organizationID,
		ReportingPeriodID: reportingPeriodID,
		MobileMethod:      opts.MobileMethod,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.store.UpsertReportingPeriodSettings(ctx, settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *ReportingPeriodService) GetSettings(ctx context.Context, reportingPeriodID string) (*domain.ReportingPeriodSettings, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	id, err := requiredID("reporting period id", reportingPeriodID)
	if err != nil {
		return nil, err
	}
	return s.store.GetReportingPeriodSettings(ctx, id)
}

func (s *ReportingPeriodService) UpsertElectricitySettings(ctx context.Context, opts UpsertElectricitySettingsOptions) (*domain.ElectricitySettings, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}

	reportingPeriodID, err := requiredID("reporting period id", opts.ReportingPeriodID)
	if err != nil {
		return nil, err
	}
	facilityID, err := requiredID("facility id", opts.FacilityID)
	if err != nil {
		return nil, err
	}
	period, err := s.store.GetReportingPeriod(ctx, reportingPeriodID)
	if err != nil {
		return nil, err
	}

	organizationID := domain.ID(cleanText(opts.OrganizationID))
	if organizationID == "" {
		organizationID = period.OrganizationID
	}
	if organizationID != period.OrganizationID {
		return nil, invalidOptions("reporting period belongs to organization %q, not %q", period.OrganizationID, organizationID)
	}

	coverage := opts.GOCoverage
	if coverage == "" {
		coverage = domain.GOCoverageNone
		if opts.HasGuaranteesOfOrigin {
			coverage = domain.GOCoverageFull
		}
	}
	if !coverage.Valid() {
		return nil, invalidOptions("invalid GO coverage %q", coverage)
	}
	if opts.HasGuaranteesOfOrigin && coverage != domain.GOCoverageFull {
		return nil, invalidOptions("guarantees of origin require full GO coverage")
	}
	if !opts.HasGuaranteesOfOrigin && coverage != domain.GOCoverageNone {
		return nil, invalidOptions("missing guarantees of origin require GO coverage none")
	}

	id := domain.ID(cleanText(opts.ID))
	if id == "" {
		id = domain.ID("electricity_settings_" + reportingPeriodID + "_" + facilityID)
	}
	var evidenceFileID *domain.ID
	if cleanText(opts.EvidenceFileID) != "" {
		id := domain.ID(cleanText(opts.EvidenceFileID))
		evidenceFileID = &id
	}

	now := time.Now().UTC()
	settings := domain.ElectricitySettings{
		ID:                    id,
		OrganizationID:        organizationID,
		ReportingPeriodID:     reportingPeriodID,
		FacilityID:            facilityID,
		HasGuaranteesOfOrigin: opts.HasGuaranteesOfOrigin,
		GOCoverage:            coverage,
		GOReference:           cleanText(opts.GOReference),
		GOMarket:              cleanText(opts.GOMarket),
		GOCancelledAt:         opts.GOCancelledAt,
		EvidenceFileID:        evidenceFileID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.store.UpsertElectricitySettings(ctx, settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *ReportingPeriodService) GetElectricitySettings(ctx context.Context, reportingPeriodID string, facilityID string) (*domain.ElectricitySettings, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	periodID, err := requiredID("reporting period id", reportingPeriodID)
	if err != nil {
		return nil, err
	}
	facility, err := requiredID("facility id", facilityID)
	if err != nil {
		return nil, err
	}
	return s.store.GetElectricitySettings(ctx, periodID, facility)
}
