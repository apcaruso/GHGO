package store

import (
	"database/sql"
	"fmt"
	"time"

	"ghgo/internal/domain"
)

func (s *Store) CreateActivityRecord(record domain.ActivityRecord) error {
	_, err := s.exec(
		`INSERT INTO activity_records (
  id, organization_id, facility_id, reporting_period_id,
  source_kind, scope, vector, category, method, activity_type,
  period_start, period_end, amount, unit,
  fuel_type, vehicle_name, plate, vehicle_type, vehicle_size_class, substance,
  status, source_hash, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.OrganizationID,
		nullStringPtr(record.FacilityID),
		record.ReportingPeriodID,
		string(record.SourceKind),
		int(record.Scope),
		string(record.Vector),
		record.Category,
		string(record.Method),
		record.ActivityType,
		formatTime(record.PeriodStart),
		formatTime(record.PeriodEnd),
		record.Amount,
		record.Unit,
		nullString(record.FuelType),
		nullString(record.VehicleName),
		nullString(record.Plate),
		nullString(record.VehicleType),
		nullString(record.VehicleSizeClass),
		nullString(record.Substance),
		string(record.Status),
		record.SourceHash,
		formatTime(record.CreatedAt),
		formatTime(record.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create activity record: %w", err)
	}
	return nil
}

func (s *Store) ListActivityRecordsByPeriod(reportingPeriodID domain.ID) ([]domain.ActivityRecord, error) {
	return s.listActivityRecords(
		`SELECT id, organization_id, facility_id, reporting_period_id,
source_kind, scope, vector, category, method, activity_type,
period_start, period_end, amount, unit,
fuel_type, vehicle_name, plate, vehicle_type, vehicle_size_class, substance,
status, source_hash, created_at, updated_at
FROM activity_records WHERE reporting_period_id = ? ORDER BY period_start, id`,
		reportingPeriodID,
	)
}

func (s *Store) ListActiveActivityRecordsByPeriod(reportingPeriodID domain.ID) ([]domain.ActivityRecord, error) {
	return s.listActivityRecords(
		`SELECT id, organization_id, facility_id, reporting_period_id,
source_kind, scope, vector, category, method, activity_type,
period_start, period_end, amount, unit,
fuel_type, vehicle_name, plate, vehicle_type, vehicle_size_class, substance,
status, source_hash, created_at, updated_at
FROM activity_records WHERE reporting_period_id = ? AND status = ? ORDER BY period_start, id`,
		reportingPeriodID,
		string(domain.ActivityRecordStatusActive),
	)
}

func (s *Store) ListActiveActivityRecordsByPeriodFacilitySource(reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind) ([]domain.ActivityRecord, error) {
	args := []any{reportingPeriodID, string(sourceKind), string(domain.ActivityRecordStatusActive)}
	facilityFilter := "facility_id IS NULL"
	if facilityID != nil {
		facilityFilter = "facility_id = ?"
		args = append(args, *facilityID)
	}

	return s.listActivityRecords(
		`SELECT id, organization_id, facility_id, reporting_period_id,
source_kind, scope, vector, category, method, activity_type,
period_start, period_end, amount, unit,
fuel_type, vehicle_name, plate, vehicle_type, vehicle_size_class, substance,
status, source_hash, created_at, updated_at
FROM activity_records WHERE reporting_period_id = ? AND source_kind = ? AND status = ? AND `+facilityFilter+` ORDER BY period_start, fuel_type, vehicle_name, plate, substance, id`,
		args...,
	)
}

func (s *Store) listActivityRecords(query string, args ...any) ([]domain.ActivityRecord, error) {
	rows, err := s.query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list activity records: %w", err)
	}
	defer rows.Close()

	var records []domain.ActivityRecord
	for rows.Next() {
		record, err := scanActivityRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan activity record: %w", err)
		}
		records = append(records, *record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list activity records rows: %w", err)
	}

	return records, nil
}

func (s *Store) CountActiveActivityRecordsByPeriodFacilitySource(reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind) (int, error) {
	args := []any{reportingPeriodID, string(sourceKind), string(domain.ActivityRecordStatusActive)}
	query := `SELECT COUNT(*) FROM activity_records WHERE reporting_period_id = ? AND source_kind = ? AND status = ? AND facility_id IS NULL`
	if facilityID != nil {
		query = `SELECT COUNT(*) FROM activity_records WHERE reporting_period_id = ? AND source_kind = ? AND status = ? AND facility_id = ?`
		args = append(args, *facilityID)
	}

	var count int
	if err := s.queryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active activity records: %w", err)
	}
	return count, nil
}

func (s *Store) CountActiveActivityRecordsByMonthlyKey(reportingPeriodID domain.ID, facilityID domain.ID, sourceKind domain.ActivitySourceKind, periodStart time.Time) (int, error) {
	var count int
	if err := s.queryRow(
		`SELECT COUNT(*) FROM activity_records
WHERE reporting_period_id = ? AND facility_id = ? AND source_kind = ? AND period_start = ? AND status = ?`,
		reportingPeriodID,
		facilityID,
		string(sourceKind),
		formatTime(periodStart),
		string(domain.ActivityRecordStatusActive),
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active monthly activity records: %w", err)
	}
	return count, nil
}

func (s *Store) SupersedeActiveActivityRecordsByMonthlyKey(reportingPeriodID domain.ID, facilityID domain.ID, sourceKind domain.ActivitySourceKind, periodStart time.Time, updatedAt time.Time) error {
	_, err := s.exec(
		`UPDATE activity_records SET status = ?, updated_at = ?
WHERE reporting_period_id = ? AND facility_id = ? AND source_kind = ? AND period_start = ? AND status = ?`,
		string(domain.ActivityRecordStatusSuperseded),
		formatTime(updatedAt),
		reportingPeriodID,
		facilityID,
		string(sourceKind),
		formatTime(periodStart),
		string(domain.ActivityRecordStatusActive),
	)
	if err != nil {
		return fmt.Errorf("supersede active monthly activity records: %w", err)
	}
	return nil
}

func (s *Store) SupersedeActiveActivityRecordsByPeriodFacilitySource(reportingPeriodID domain.ID, facilityID *domain.ID, sourceKind domain.ActivitySourceKind, updatedAt time.Time) error {
	args := []any{string(domain.ActivityRecordStatusSuperseded), formatTime(updatedAt), reportingPeriodID, string(sourceKind), string(domain.ActivityRecordStatusActive)}
	facilityFilter := "facility_id IS NULL"
	if facilityID != nil {
		facilityFilter = "facility_id = ?"
		args = append(args, *facilityID)
	}

	_, err := s.exec(
		`UPDATE activity_records SET status = ?, updated_at = ?
WHERE reporting_period_id = ? AND source_kind = ? AND status = ? AND `+facilityFilter,
		args...,
	)
	if err != nil {
		return fmt.Errorf("supersede active activity records: %w", err)
	}
	return nil
}

func scanActivityRecord(row scanner) (*domain.ActivityRecord, error) {
	var record domain.ActivityRecord
	var facilityID sql.NullString
	var sourceKind string
	var scope int
	var vector string
	var method string
	var periodStart string
	var periodEnd string
	var fuelType sql.NullString
	var vehicleName sql.NullString
	var plate sql.NullString
	var vehicleType sql.NullString
	var vehicleSizeClass sql.NullString
	var substance sql.NullString
	var status string
	var createdAt string
	var updatedAt string

	if err := row.Scan(
		&record.ID,
		&record.OrganizationID,
		&facilityID,
		&record.ReportingPeriodID,
		&sourceKind,
		&scope,
		&vector,
		&record.Category,
		&method,
		&record.ActivityType,
		&periodStart,
		&periodEnd,
		&record.Amount,
		&record.Unit,
		&fuelType,
		&vehicleName,
		&plate,
		&vehicleType,
		&vehicleSizeClass,
		&substance,
		&status,
		&record.SourceHash,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	var err error
	record.FacilityID = stringPtrFromNull(facilityID)
	record.SourceKind = domain.ActivitySourceKind(sourceKind)
	record.Scope = domain.Scope(scope)
	record.Vector = domain.ActivityVector(vector)
	record.Method = domain.ActivityMethod(method)
	record.PeriodStart, err = parseTime(periodStart)
	if err != nil {
		return nil, err
	}
	record.PeriodEnd, err = parseTime(periodEnd)
	if err != nil {
		return nil, err
	}
	record.FuelType = stringFromNull(fuelType)
	record.VehicleName = stringFromNull(vehicleName)
	record.Plate = stringFromNull(plate)
	record.VehicleType = stringFromNull(vehicleType)
	record.VehicleSizeClass = stringFromNull(vehicleSizeClass)
	record.Substance = stringFromNull(substance)
	record.Status = domain.ActivityRecordStatus(status)
	record.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	record.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return nil, err
	}

	return &record, nil
}
