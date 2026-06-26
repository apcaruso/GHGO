package store

import (
	"context"
	"database/sql"
	"fmt"

	"ghgo/internal/domain"
)

type ReportResultRow struct {
	CalculationResult domain.CalculationResult
	ActivityRecord    domain.ActivityRecord
}

func (s *Store) ListReportResultRows(ctx context.Context, calculationRunID domain.ID) ([]ReportResultRow, error) {
	rows, err := s.queryContext(ctx, `SELECT
cr.id, cr.calculation_run_id, cr.activity_record_id, cr.scope, cr.vector, cr.method,
cr.activity_amount, cr.activity_unit, cr.factor_id, cr.factor_value, cr.factor_unit, cr.factor_source,
cr.emissions_kgco2e, cr.is_primary, cr.notes_json,
ar.id, ar.organization_id, ar.facility_id, ar.reporting_period_id,
ar.source_kind, ar.scope, ar.vector, ar.category, ar.method, ar.activity_type,
ar.period_start, ar.period_end, ar.amount, ar.unit,
ar.fuel_type, ar.vehicle_name, ar.plate, ar.vehicle_type, ar.vehicle_size_class, ar.substance,
ar.status, ar.source_hash, ar.created_at, ar.updated_at
FROM calculation_results cr
JOIN activity_records ar ON ar.id = cr.activity_record_id
WHERE cr.calculation_run_id = ?
ORDER BY ar.source_kind, COALESCE(ar.facility_id, ''), ar.period_start, ar.id, cr.method, cr.id`, calculationRunID)
	if err != nil {
		return nil, fmt.Errorf("list report result rows: %w", err)
	}
	defer rows.Close()

	var resultRows []ReportResultRow
	for rows.Next() {
		row, err := scanReportResultRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan report result row: %w", err)
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list report result rows rows: %w", err)
	}

	return resultRows, nil
}

func scanReportResultRow(row scanner) (ReportResultRow, error) {
	var result domain.CalculationResult
	var record domain.ActivityRecord
	var resultScope int
	var resultVector string
	var resultMethod string
	var factorID sql.NullString
	var isPrimary int
	var facilityID sql.NullString
	var sourceKind string
	var recordScope int
	var recordVector string
	var recordMethod string
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
		&result.ID,
		&result.CalculationRunID,
		&result.ActivityRecordID,
		&resultScope,
		&resultVector,
		&resultMethod,
		&result.ActivityAmount,
		&result.ActivityUnit,
		&factorID,
		&result.FactorValue,
		&result.FactorUnit,
		&result.FactorSource,
		&result.EmissionsKgCO2e,
		&isPrimary,
		&result.NotesJSON,
		&record.ID,
		&record.OrganizationID,
		&facilityID,
		&record.ReportingPeriodID,
		&sourceKind,
		&recordScope,
		&recordVector,
		&record.Category,
		&recordMethod,
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
		return ReportResultRow{}, err
	}

	var err error
	result.Scope = domain.Scope(resultScope)
	result.Vector = domain.ActivityVector(resultVector)
	result.Method = domain.ActivityMethod(resultMethod)
	result.FactorID = stringPtrFromNull(factorID)
	result.IsPrimary = intToBool(isPrimary)

	record.FacilityID = stringPtrFromNull(facilityID)
	record.SourceKind = domain.ActivitySourceKind(sourceKind)
	record.Scope = domain.Scope(recordScope)
	record.Vector = domain.ActivityVector(recordVector)
	record.Method = domain.ActivityMethod(recordMethod)
	record.PeriodStart, err = parseTime(periodStart)
	if err != nil {
		return ReportResultRow{}, err
	}
	record.PeriodEnd, err = parseTime(periodEnd)
	if err != nil {
		return ReportResultRow{}, err
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
		return ReportResultRow{}, err
	}
	record.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return ReportResultRow{}, err
	}

	return ReportResultRow{
		CalculationResult: result,
		ActivityRecord:    record,
	}, nil
}
