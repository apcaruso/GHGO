package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"ghgo/internal/domain"
)

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

func (s *Store) FindEmissionFactors(ctx context.Context, q EmissionFactorQuery) ([]domain.EmissionFactor, error) {
	if q.FactorSetID == "" {
		return nil, fmt.Errorf("factor_set_id is required")
	}

	clauses := []string{"factor_set_id = ?"}
	args := []any{q.FactorSetID}
	addIntFilter(&clauses, &args, "scope", q.Scope)
	addNullableTextFilter(&clauses, &args, "activity_type", q.ActivityType)
	addNullableTextFilter(&clauses, &args, "fuel_type", q.FuelType)
	addNullableTextFilter(&clauses, &args, "vehicle_type", q.VehicleType)
	addNullableTextFilter(&clauses, &args, "vehicle_size_class", q.VehicleSizeClass)
	addNullableTextFilter(&clauses, &args, "substance", q.Substance)
	addTextFilter(&clauses, &args, "input_unit", q.InputUnit)
	addTextFilter(&clauses, &args, "factor_unit", q.FactorUnit)
	addTextFilter(&clauses, &args, "ghg", q.GHG)

	rows, err := s.queryContext(ctx, `SELECT id, factor_set_id, source, scope, level_1, level_2, level_3, level_4, column_text,
activity_type, fuel_type, vehicle_type, vehicle_size_class, substance,
input_unit, factor_unit, ghg, factor_value, metadata_json
FROM emission_factors WHERE `+strings.Join(clauses, " AND ")+` ORDER BY id`, args...)
	if err != nil {
		return nil, fmt.Errorf("find emission factors: %w", err)
	}
	defer rows.Close()

	factors := []domain.EmissionFactor{}
	for rows.Next() {
		factor, err := scanEmissionFactor(rows)
		if err != nil {
			return nil, fmt.Errorf("scan emission factor: %w", err)
		}
		factors = append(factors, *factor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("find emission factors rows: %w", err)
	}

	return factors, nil
}

func (s *Store) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(ctx, query, args...)
	}
	return s.db.QueryContext(ctx, query, args...)
}

func addIntFilter(clauses *[]string, args *[]any, column string, value *int) {
	if value == nil {
		return
	}
	*clauses = append(*clauses, column+" = ?")
	*args = append(*args, *value)
}

func addTextFilter(clauses *[]string, args *[]any, column string, value *string) {
	if value == nil {
		return
	}
	*clauses = append(*clauses, column+" = ?")
	*args = append(*args, *value)
}

func addNullableTextFilter(clauses *[]string, args *[]any, column string, value *string) {
	if value == nil {
		return
	}
	if *value == "" {
		*clauses = append(*clauses, "("+column+" IS NULL OR "+column+" = '')")
		return
	}
	*clauses = append(*clauses, column+" = ?")
	*args = append(*args, *value)
}
