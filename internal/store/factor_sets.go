package store

import (
	"database/sql"
	"errors"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateFactorSet(factorSet domain.FactorSet) error {
	_, err := s.exec(
		`INSERT INTO factor_sets (id, name, source, year, version, imported_at, metadata_json) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		factorSet.ID,
		factorSet.Name,
		factorSet.Source,
		factorSet.Year,
		factorSet.Version,
		formatTime(factorSet.ImportedAt),
		factorSet.MetadataJSON,
	)
	if err != nil {
		return fmt.Errorf("create factor set: %w", err)
	}
	return nil
}

func (s *Store) GetFactorSet(id domain.ID) (*domain.FactorSet, error) {
	factorSet, err := scanFactorSet(s.queryRow(
		`SELECT id, name, source, year, version, imported_at, metadata_json FROM factor_sets WHERE id = ?`,
		id,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get factor set: %w", err)
	}
	return factorSet, nil
}

func (s *Store) FindFactorSetBySourceYearVersion(source string, year int, version string) (*domain.FactorSet, error) {
	factorSet, err := scanFactorSet(s.queryRow(
		`SELECT id, name, source, year, version, imported_at, metadata_json FROM factor_sets WHERE source = ? AND year = ? AND version = ?`,
		source,
		year,
		version,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find factor set: %w", err)
	}
	return factorSet, nil
}

func (s *Store) ListFactorSets() ([]domain.FactorSet, error) {
	rows, err := s.query(`SELECT id, name, source, year, version, imported_at, metadata_json FROM factor_sets ORDER BY source, year, version, id`)
	if err != nil {
		return nil, fmt.Errorf("list factor sets: %w", err)
	}
	defer rows.Close()

	var factorSets []domain.FactorSet
	for rows.Next() {
		factorSet, err := scanFactorSet(rows)
		if err != nil {
			return nil, fmt.Errorf("scan factor set: %w", err)
		}
		factorSets = append(factorSets, *factorSet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list factor sets rows: %w", err)
	}

	return factorSets, nil
}

func scanFactorSet(row scanner) (*domain.FactorSet, error) {
	var factorSet domain.FactorSet
	var importedAt string
	if err := row.Scan(
		&factorSet.ID,
		&factorSet.Name,
		&factorSet.Source,
		&factorSet.Year,
		&factorSet.Version,
		&importedAt,
		&factorSet.MetadataJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	factorSet.ImportedAt, err = parseTime(importedAt)
	if err != nil {
		return nil, err
	}

	return &factorSet, nil
}
