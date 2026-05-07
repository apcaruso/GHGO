package store

import (
	"context"
	"fmt"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/ports"
)

type Repository struct {
	store *Store
}

var _ ports.Store = (*Repository)(nil)

func NewRepository(st *Store) *Repository {
	return &Repository{store: st}
}

func (r *Repository) RawStore() *Store {
	if r == nil {
		return nil
	}
	return r.store
}

func (r *Repository) check(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.store == nil {
		return fmt.Errorf("store is required")
	}
	return nil
}

func (r *Repository) WithTx(ctx context.Context, fn func(ports.Store) error) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("transaction function is required")
	}
	return r.store.WithTx(ctx, func(tx *Store) error {
		return fn(NewRepository(tx))
	})
}

func (r *Repository) CreateOrganization(ctx context.Context, o domain.Organization) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateOrganization(o)
}

func (r *Repository) GetOrganization(ctx context.Context, id domain.ID) (*domain.Organization, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetOrganization(id)
}

func (r *Repository) ListOrganizations(ctx context.Context) ([]domain.Organization, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListOrganizations()
}

func (r *Repository) CreateFacility(ctx context.Context, f domain.Facility) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateFacility(f)
}

func (r *Repository) ListFacilitiesByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.Facility, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListFacilitiesByOrganization(organizationID)
}

func (r *Repository) CreateReportingPeriod(ctx context.Context, p domain.ReportingPeriod) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateReportingPeriod(p)
}

func (r *Repository) GetReportingPeriod(ctx context.Context, id domain.ID) (*domain.ReportingPeriod, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetReportingPeriod(id)
}

func (r *Repository) ListReportingPeriodsByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.ReportingPeriod, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListReportingPeriodsByOrganization(organizationID)
}

func (r *Repository) UpsertReportingPeriodSettings(ctx context.Context, settings domain.ReportingPeriodSettings) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.UpsertReportingPeriodSettings(settings)
}

func (r *Repository) GetReportingPeriodSettings(ctx context.Context, reportingPeriodID domain.ID) (*domain.ReportingPeriodSettings, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetReportingPeriodSettings(reportingPeriodID)
}

func (r *Repository) UpsertElectricitySettings(ctx context.Context, settings domain.ElectricitySettings) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.UpsertElectricitySettings(settings)
}

func (r *Repository) GetElectricitySettings(ctx context.Context, reportingPeriodID, facilityID domain.ID) (*domain.ElectricitySettings, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetElectricitySettings(reportingPeriodID, facilityID)
}

func (r *Repository) CreateActivityRecord(ctx context.Context, record domain.ActivityRecord) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateActivityRecord(record)
}

func (r *Repository) ListActivityRecordsByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.ActivityRecord, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListActivityRecordsByPeriod(reportingPeriodID)
}

func (r *Repository) ListActiveActivityRecordsByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.ActivityRecord, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListActiveActivityRecordsByPeriod(reportingPeriodID)
}

func (r *Repository) ListActiveActivityRecordsByPeriodFacilitySource(ctx context.Context, reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind) ([]domain.ActivityRecord, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListActiveActivityRecordsByPeriodFacilitySource(reportingPeriodID, facilityID, sourceKind)
}

func (r *Repository) CountActiveActivityRecordsByPeriodFacilitySource(ctx context.Context, reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind) (int, error) {
	if err := r.check(ctx); err != nil {
		return 0, err
	}
	return r.store.CountActiveActivityRecordsByPeriodFacilitySource(reportingPeriodID, facilityID, sourceKind)
}

func (r *Repository) CountActiveActivityRecordsByMonthlyKey(ctx context.Context, reportingPeriodID domain.ID, facilityID domain.ID, sourceKind domain.ActivitySourceKind, periodStart time.Time) (int, error) {
	if err := r.check(ctx); err != nil {
		return 0, err
	}
	return r.store.CountActiveActivityRecordsByMonthlyKey(reportingPeriodID, facilityID, sourceKind, periodStart)
}

func (r *Repository) SupersedeActiveActivityRecordsByMonthlyKey(ctx context.Context, reportingPeriodID domain.ID, facilityID domain.ID, sourceKind domain.ActivitySourceKind, periodStart time.Time, updatedAt time.Time) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.SupersedeActiveActivityRecordsByMonthlyKey(reportingPeriodID, facilityID, sourceKind, periodStart, updatedAt)
}

func (r *Repository) SupersedeActiveActivityRecordsByPeriodFacilitySource(ctx context.Context, reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind, updatedAt time.Time) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.SupersedeActiveActivityRecordsByPeriodFacilitySource(reportingPeriodID, facilityID, sourceKind, updatedAt)
}

func (r *Repository) CreatePasteBatch(ctx context.Context, batch domain.PasteBatch) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreatePasteBatch(batch)
}

func (r *Repository) GetPasteBatch(ctx context.Context, id domain.ID) (*domain.PasteBatch, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetPasteBatch(id)
}

func (r *Repository) ListPasteBatchesByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.PasteBatch, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListPasteBatchesByPeriod(reportingPeriodID)
}

func (r *Repository) MarkPasteBatchCommitted(ctx context.Context, id domain.ID, committedAt time.Time) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.MarkPasteBatchCommitted(id, committedAt)
}

func (r *Repository) CreatePasteRow(ctx context.Context, row domain.PasteRow) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreatePasteRow(row)
}

func (r *Repository) ListPasteRowsByBatch(ctx context.Context, pasteBatchID domain.ID) ([]domain.PasteRow, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListPasteRowsByBatch(pasteBatchID)
}

func (r *Repository) CreateFactorSet(ctx context.Context, factorSet domain.FactorSet) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateFactorSet(factorSet)
}

func (r *Repository) GetFactorSet(ctx context.Context, id domain.ID) (*domain.FactorSet, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetFactorSet(id)
}

func (r *Repository) FindFactorSetBySourceYearVersion(ctx context.Context, source string, year int, version string) (*domain.FactorSet, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.FindFactorSetBySourceYearVersion(source, year, version)
}

func (r *Repository) ListFactorSets(ctx context.Context) ([]domain.FactorSet, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListFactorSets()
}

func (r *Repository) CreateEmissionFactor(ctx context.Context, factor domain.EmissionFactor) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateEmissionFactor(factor)
}

func (r *Repository) DeleteEmissionFactorsBySet(ctx context.Context, factorSetID domain.ID) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.DeleteEmissionFactorsBySet(factorSetID)
}

func (r *Repository) CountEmissionFactorsBySet(ctx context.Context, factorSetID domain.ID) (int, error) {
	if err := r.check(ctx); err != nil {
		return 0, err
	}
	return r.store.CountEmissionFactorsBySet(factorSetID)
}

func (r *Repository) ListEmissionFactorsBySet(ctx context.Context, factorSetID domain.ID) ([]domain.EmissionFactor, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListEmissionFactorsBySet(factorSetID)
}

func (r *Repository) FindEmissionFactors(ctx context.Context, q ports.EmissionFactorQuery) ([]domain.EmissionFactor, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.FindEmissionFactors(ctx, q)
}

func (r *Repository) CreateCalculationRun(ctx context.Context, run domain.CalculationRun) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateCalculationRun(run)
}

func (r *Repository) CompleteCalculationRun(ctx context.Context, id domain.ID, completedAt time.Time) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CompleteCalculationRun(id, completedAt)
}

func (r *Repository) GetCalculationRun(ctx context.Context, id domain.ID) (*domain.CalculationRun, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.GetCalculationRun(id)
}

func (r *Repository) ListCalculationRunsByPeriod(ctx context.Context, reportingPeriodID domain.ID) ([]domain.CalculationRun, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListCalculationRunsByPeriod(reportingPeriodID)
}

func (r *Repository) CreateCalculationResult(ctx context.Context, result domain.CalculationResult) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateCalculationResult(result)
}

func (r *Repository) ListCalculationResultsByRun(ctx context.Context, calculationRunID domain.ID) ([]domain.CalculationResult, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListCalculationResultsByRun(calculationRunID)
}

func (r *Repository) ListReportResultRows(ctx context.Context, calculationRunID domain.ID) ([]ports.ReportResultRow, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListReportResultRows(ctx, calculationRunID)
}

func (r *Repository) CreateAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	return r.store.CreateAuditEvent(event)
}

func (r *Repository) ListAuditEventsByEntity(ctx context.Context, entityType string, entityID domain.ID) ([]domain.AuditEvent, error) {
	if err := r.check(ctx); err != nil {
		return nil, err
	}
	return r.store.ListAuditEventsByEntity(entityType, entityID)
}
