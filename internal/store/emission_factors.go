package store

import (
	"database/sql"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateEmissionFactor(factor domain.EmissionFactor) error {
	_, err := s.exec(
		`INSERT INTO emission_factors (
  id, factor_set_id, source, scope, level_1, level_2, level_3, level_4, column_text,
  activity_type, fuel_type, vehicle_type, vehicle_size_class, substance,
  input_unit, factor_unit, ghg, factor_value, metadata_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		factor.ID,
		factor.FactorSetID,
		factor.Source,
		int(factor.Scope),
		nullString(factor.Level1),
		nullString(factor.Level2),
		nullString(factor.Level3),
		nullString(factor.Level4),
		nullString(factor.ColumnText),
		nullString(factor.ActivityType),
		nullString(factor.FuelType),
		nullString(factor.VehicleType),
		nullString(factor.VehicleSizeClass),
		nullString(factor.Substance),
		factor.InputUnit,
		factor.FactorUnit,
		factor.GHG,
		factor.FactorValue,
		factor.MetadataJSON,
	)
	if err != nil {
		return fmt.Errorf("create emission factor: %w", err)
	}
	return nil
}

func (s *Store) DeleteEmissionFactorsBySet(factorSetID domain.ID) error {
	_, err := s.exec(`DELETE FROM emission_factors WHERE factor_set_id = ?`, factorSetID)
	if err != nil {
		return fmt.Errorf("delete emission factors by set: %w", err)
	}
	return nil
}

func (s *Store) CountEmissionFactorsBySet(factorSetID domain.ID) (int, error) {
	var count int
	if err := s.queryRow(`SELECT COUNT(*) FROM emission_factors WHERE factor_set_id = ?`, factorSetID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count emission factors by set: %w", err)
	}
	return count, nil
}

func (s *Store) ListEmissionFactorsBySet(factorSetID domain.ID) ([]domain.EmissionFactor, error) {
	rows, err := s.query(
		`SELECT id, factor_set_id, source, scope, level_1, level_2, level_3, level_4, column_text,
activity_type, fuel_type, vehicle_type, vehicle_size_class, substance,
input_unit, factor_unit, ghg, factor_value, metadata_json
FROM emission_factors WHERE factor_set_id = ? ORDER BY id`,
		factorSetID,
	)
	if err != nil {
		return nil, fmt.Errorf("list emission factors: %w", err)
	}
	defer rows.Close()

	var factors []domain.EmissionFactor
	for rows.Next() {
		factor, err := scanEmissionFactor(rows)
		if err != nil {
			return nil, fmt.Errorf("scan emission factor: %w", err)
		}
		factors = append(factors, *factor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list emission factors rows: %w", err)
	}

	return factors, nil
}

func scanEmissionFactor(row scanner) (*domain.EmissionFactor, error) {
	var factor domain.EmissionFactor
	var scope int
	var level1 sql.NullString
	var level2 sql.NullString
	var level3 sql.NullString
	var level4 sql.NullString
	var columnText sql.NullString
	var activityType sql.NullString
	var fuelType sql.NullString
	var vehicleType sql.NullString
	var vehicleSizeClass sql.NullString
	var substance sql.NullString
	if err := row.Scan(
		&factor.ID,
		&factor.FactorSetID,
		&factor.Source,
		&scope,
		&level1,
		&level2,
		&level3,
		&level4,
		&columnText,
		&activityType,
		&fuelType,
		&vehicleType,
		&vehicleSizeClass,
		&substance,
		&factor.InputUnit,
		&factor.FactorUnit,
		&factor.GHG,
		&factor.FactorValue,
		&factor.MetadataJSON,
	); err != nil {
		return nil, err
	}

	factor.Scope = domain.Scope(scope)
	factor.Level1 = stringFromNull(level1)
	factor.Level2 = stringFromNull(level2)
	factor.Level3 = stringFromNull(level3)
	factor.Level4 = stringFromNull(level4)
	factor.ColumnText = stringFromNull(columnText)
	factor.ActivityType = stringFromNull(activityType)
	factor.FuelType = stringFromNull(fuelType)
	factor.VehicleType = stringFromNull(vehicleType)
	factor.VehicleSizeClass = stringFromNull(vehicleSizeClass)
	factor.Substance = stringFromNull(substance)

	return &factor, nil
}
