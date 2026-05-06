package store

import (
	"database/sql"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateCalculationResult(result domain.CalculationResult) error {
	_, err := s.exec(
		`INSERT INTO calculation_results (
  id, calculation_run_id, activity_record_id, scope, vector, method,
  activity_amount, activity_unit, factor_id, factor_value, factor_unit, factor_source,
  emissions_kgco2e, is_primary, notes_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID,
		result.CalculationRunID,
		result.ActivityRecordID,
		int(result.Scope),
		string(result.Vector),
		string(result.Method),
		result.ActivityAmount,
		result.ActivityUnit,
		nullStringPtr(result.FactorID),
		result.FactorValue,
		result.FactorUnit,
		result.FactorSource,
		result.EmissionsKgCO2e,
		boolToInt(result.IsPrimary),
		result.NotesJSON,
	)
	if err != nil {
		return fmt.Errorf("create calculation result: %w", err)
	}
	return nil
}

func (s *Store) ListCalculationResultsByRun(calculationRunID domain.ID) ([]domain.CalculationResult, error) {
	rows, err := s.query(
		`SELECT id, calculation_run_id, activity_record_id, scope, vector, method,
activity_amount, activity_unit, factor_id, factor_value, factor_unit, factor_source,
emissions_kgco2e, is_primary, notes_json
FROM calculation_results WHERE calculation_run_id = ? ORDER BY id`,
		calculationRunID,
	)
	if err != nil {
		return nil, fmt.Errorf("list calculation results: %w", err)
	}
	defer rows.Close()

	var results []domain.CalculationResult
	for rows.Next() {
		result, err := scanCalculationResult(rows)
		if err != nil {
			return nil, fmt.Errorf("scan calculation result: %w", err)
		}
		results = append(results, *result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list calculation results rows: %w", err)
	}

	return results, nil
}

func scanCalculationResult(row scanner) (*domain.CalculationResult, error) {
	var result domain.CalculationResult
	var scope int
	var vector string
	var method string
	var factorID sql.NullString
	var isPrimary int
	if err := row.Scan(
		&result.ID,
		&result.CalculationRunID,
		&result.ActivityRecordID,
		&scope,
		&vector,
		&method,
		&result.ActivityAmount,
		&result.ActivityUnit,
		&factorID,
		&result.FactorValue,
		&result.FactorUnit,
		&result.FactorSource,
		&result.EmissionsKgCO2e,
		&isPrimary,
		&result.NotesJSON,
	); err != nil {
		return nil, err
	}

	result.Scope = domain.Scope(scope)
	result.Vector = domain.ActivityVector(vector)
	result.Method = domain.ActivityMethod(method)
	result.FactorID = stringPtrFromNull(factorID)
	result.IsPrimary = intToBool(isPrimary)

	return &result, nil
}
