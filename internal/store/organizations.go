package store

import (
	"database/sql"
	"errors"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateOrganization(o domain.Organization) error {
	_, err := s.exec(
		`INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		o.ID,
		o.Name,
		formatTime(o.CreatedAt),
		formatTime(o.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create organization: %w", err)
	}
	return nil
}

func (s *Store) GetOrganization(id domain.ID) (*domain.Organization, error) {
	o, err := scanOrganization(s.queryRow(
		`SELECT id, name, created_at, updated_at FROM organizations WHERE id = ?`,
		id,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return o, nil
}

func (s *Store) ListOrganizations() ([]domain.Organization, error) {
	rows, err := s.query(`SELECT id, name, created_at, updated_at FROM organizations ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var organizations []domain.Organization
	for rows.Next() {
		o, err := scanOrganization(rows)
		if err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		organizations = append(organizations, *o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list organizations rows: %w", err)
	}

	return organizations, nil
}

func scanOrganization(row scanner) (*domain.Organization, error) {
	var o domain.Organization
	var createdAt string
	var updatedAt string
	if err := row.Scan(&o.ID, &o.Name, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	o.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	o.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return nil, err
	}

	return &o, nil
}
