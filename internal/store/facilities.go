package store

import (
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateFacility(f domain.Facility) error {
	_, err := s.exec(
		`INSERT INTO facilities (id, organization_id, name, country_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		f.ID,
		f.OrganizationID,
		f.Name,
		f.CountryCode,
		formatTime(f.CreatedAt),
		formatTime(f.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create facility: %w", err)
	}
	return nil
}

func (s *Store) ListFacilitiesByOrganization(organizationID domain.ID) ([]domain.Facility, error) {
	rows, err := s.query(
		`SELECT id, organization_id, name, country_code, created_at, updated_at FROM facilities WHERE organization_id = ? ORDER BY name, id`,
		organizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list facilities: %w", err)
	}
	defer rows.Close()

	var facilities []domain.Facility
	for rows.Next() {
		f, err := scanFacility(rows)
		if err != nil {
			return nil, fmt.Errorf("scan facility: %w", err)
		}
		facilities = append(facilities, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list facilities rows: %w", err)
	}

	return facilities, nil
}

func scanFacility(row scanner) (*domain.Facility, error) {
	var f domain.Facility
	var createdAt string
	var updatedAt string
	if err := row.Scan(&f.ID, &f.OrganizationID, &f.Name, &f.CountryCode, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	var err error
	f.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	f.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return nil, err
	}

	return &f, nil
}
