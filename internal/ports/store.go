package ports

import (
	"context"
	"errors"
	"time"

	"ghgo/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Transactor
	OrganizationStore
	FacilityStore
	ReportingPeriodStore
	SettingsStore
	ElectricitySettingsStore
	ActivityRecordStore
	PasteStore
	FactorStore
	CalculationStore
	ReportStore
	AuditStore
}

type Transactor interface {
	WithTx(ctx context.Context, fn func(Store) error) error
}

type OrganizationStore interface {
	CreateOrganization(ctx context.Context, o domain.Organization) error
	GetOrganization(ctx context.Context, id domain.ID) (*domain.Organization, error)
	ListOrganizations(ctx context.Context) ([]domain.Organization, error)
}

type FacilityStore interface {
	CreateFacility(ctx context.Context, f domain.Facility) error
	ListFacilitiesByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.Facility, error)
}

type ReportingPeriodStore interface {
	CreateReportingPeriod(ctx context.Context, p domain.ReportingPeriod) error
	GetReportingPeriod(ctx context.Context, id domain.ID) (*domain.ReportingPeriod, error)
	ListReportingPeriodsByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.ReportingPeriod, error)
}

type SettingsStore interface {
	UpsertReportingPeriodSettings(ctx context.Context, settings domain.ReportingPeriodSettings) error
	GetReportingPeriodSettings(ctx context.Context, reportingPeriodID domain.ID) (*domain.ReportingPeriodSettings, error)
}

type ElectricitySettingsStore interface {
	UpsertElectricitySettings(ctx context.Context, settings domain.ElectricitySettings) error
	GetElectricitySettings(ctx context.Context, reportingPeriodID, facilityID domain.ID) (*domain.ElectricitySettings, error)
}

type ActivityRecordStore interface {
	CreateActivityRecord(ctx context.Context, record domain.ActivityRecord) error
	ListActivityRecordsByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.ActivityRecord, error)
	ListActiveActivityRecordsByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.ActivityRecord, error)
	ListActiveActivityRecordsByPeriodFacilitySource(ctx context.Context, reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind) ([]domain.ActivityRecord, error)
	CountActiveActivityRecordsByPeriodFacilitySource(ctx context.Context, reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind) (int, error)
	CountActiveActivityRecordsByMonthlyKey(ctx context.Context, reportingPeriodID domain.ID, facilityID domain.ID, sourceKind domain.ActivitySourceKind, periodStart time.Time) (int, error)
	SupersedeActiveActivityRecordsByMonthlyKey(ctx context.Context, reportingPeriodID domain.ID, facilityID domain.ID, sourceKind domain.ActivitySourceKind, periodStart time.Time, updatedAt time.Time) error
	SupersedeActiveActivityRecordsByPeriodFacilitySource(ctx context.Context, reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind, updatedAt time.Time) error
}

type PasteStore interface {
	CreatePasteBatch(ctx context.Context, batch domain.PasteBatch) error
	GetPasteBatch(ctx context.Context, id domain.ID) (*domain.PasteBatch, error)
	ListPasteBatchesByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.PasteBatch, error)
	MarkPasteBatchCommitted(ctx context.Context, id domain.ID, committedAt time.Time) error
	CreatePasteRow(ctx context.Context, row domain.PasteRow) error
	ListPasteRowsByBatch(ctx context.Context, pasteBatchID domain.ID) ([]domain.PasteRow, error)
}

type FactorStore interface {
	CreateFactorSet(ctx context.Context, factorSet domain.FactorSet) error
	GetFactorSet(ctx context.Context, id domain.ID) (*domain.FactorSet, error)
	FindFactorSetBySourceYearVersion(ctx context.Context, source string, year int, version string) (*domain.FactorSet, error)
	ListFactorSets(ctx context.Context) ([]domain.FactorSet, error)
	CreateEmissionFactor(ctx context.Context, factor domain.EmissionFactor) error
	DeleteEmissionFactorsBySet(ctx context.Context, factorSetID domain.ID) error
	CountEmissionFactorsBySet(ctx context.Context, factorSetID domain.ID) (int, error)
	ListEmissionFactorsBySet(ctx context.Context, factorSetID domain.ID) ([]domain.EmissionFactor, error)
	FindEmissionFactors(ctx context.Context, q EmissionFactorQuery) ([]domain.EmissionFactor, error)
}

type CalculationStore interface {
	CreateCalculationRun(ctx context.Context, run domain.CalculationRun) error
	CompleteCalculationRun(ctx context.Context, id domain.ID, completedAt time.Time) error
	GetCalculationRun(ctx context.Context, id domain.ID) (*domain.CalculationRun, error)
	ListCalculationRunsByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.CalculationRun, error)
	CreateCalculationResult(ctx context.Context, result domain.CalculationResult) error
	ListCalculationResultsByRun(ctx context.Context, calculationRunID domain.ID) ([]domain.CalculationResult, error)
}

type ReportStore interface {
	ListReportResultRows(ctx context.Context, calculationRunID domain.ID) ([]ReportResultRow, error)
}

type AuditStore interface {
	CreateAuditEvent(ctx context.Context, event domain.AuditEvent) error
	ListAuditEventsByEntity(ctx context.Context, entityType string, entityID domain.ID) ([]domain.AuditEvent, error)
}

type EmissionFactorQuery struct {
	FactorSetID string

	Scope            *int
	ActivityType     *string
	FuelType         *string
	VehicleType      *string
	VehicleSizeClass *string
	Substance        *string
	InputUnit        *string
	FactorUnit       *string
	GHG              *string
}

type ReportResultRow struct {
	CalculationResult domain.CalculationResult
	ActivityRecord    domain.ActivityRecord
}
