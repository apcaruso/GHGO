package store

import (
	"database/sql"
	"errors"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateReportingPeriod(p domain.ReportingPeriod) error {
	_, err := s.exec(
		`INSERT INTO reporting_periods (id, organization_id, year, starts_on, ends_on, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID,
		p.OrganizationID,
		p.Year,
		formatTime(p.StartsOn),
		formatTime(p.EndsOn),
		string(p.Status),
		formatTime(p.CreatedAt),
		formatTime(p.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create reporting period: %w", err)
	}
	return nil
}

func (s *Store) GetReportingPeriod(id domain.ID) (*domain.ReportingPeriod, error) {
	p, err := scanReportingPeriod(s.queryRow(
		`SELECT id, organization_id, year, starts_on, ends_on, status, created_at, updated_at FROM reporting_periods WHERE id = ?`,
		id,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reporting period: %w", err)
	}
	return p, nil
}

func (s *Store) ListReportingPeriodsByOrganization(organizationID domain.ID) ([]domain.ReportingPeriod, error) {
	rows, err := s.query(
		`SELECT id, organization_id, year, starts_on, ends_on, status, created_at, updated_at FROM reporting_periods WHERE organization_id = ? ORDER BY year, id`,
		organizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reporting periods: %w", err)
	}
	defer rows.Close()

	var periods []domain.ReportingPeriod
	for rows.Next() {
		p, err := scanReportingPeriod(rows)
		if err != nil {
			return nil, fmt.Errorf("scan reporting period: %w", err)
		}
		periods = append(periods, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list reporting periods rows: %w", err)
	}

	return periods, nil
}

func scanReportingPeriod(row scanner) (*domain.ReportingPeriod, error) {
	var p domain.ReportingPeriod
	var startsOn string
	var endsOn string
	var status string
	var createdAt string
	var updatedAt string
	if err := row.Scan(&p.ID, &p.OrganizationID, &p.Year, &startsOn, &endsOn, &status, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	p.StartsOn, err = parseTime(startsOn)
	if err != nil {
		return nil, err
	}
	p.EndsOn, err = parseTime(endsOn)
	if err != nil {
		return nil, err
	}
	p.Status = domain.ReportingPeriodStatus(status)
	p.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	p.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
