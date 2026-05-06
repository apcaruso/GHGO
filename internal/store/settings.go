package store

import (
	"database/sql"
	"errors"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) UpsertReportingPeriodSettings(settings domain.ReportingPeriodSettings) error {
	_, err := s.exec(
		`INSERT INTO reporting_period_settings (id, organization_id, reporting_period_id, mobile_method, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(reporting_period_id) DO UPDATE SET
  id = excluded.id,
  organization_id = excluded.organization_id,
  mobile_method = excluded.mobile_method,
  updated_at = excluded.updated_at`,
		settings.ID,
		settings.OrganizationID,
		settings.ReportingPeriodID,
		string(settings.MobileMethod),
		formatTime(settings.CreatedAt),
		formatTime(settings.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert reporting period settings: %w", err)
	}
	return nil
}

func (s *Store) GetReportingPeriodSettings(reportingPeriodID domain.ID) (*domain.ReportingPeriodSettings, error) {
	settings, err := scanReportingPeriodSettings(s.queryRow(
		`SELECT id, organization_id, reporting_period_id, mobile_method, created_at, updated_at FROM reporting_period_settings WHERE reporting_period_id = ?`,
		reportingPeriodID,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reporting period settings: %w", err)
	}
	return settings, nil
}

func scanReportingPeriodSettings(row scanner) (*domain.ReportingPeriodSettings, error) {
	var settings domain.ReportingPeriodSettings
	var mobileMethod string
	var createdAt string
	var updatedAt string
	if err := row.Scan(&settings.ID, &settings.OrganizationID, &settings.ReportingPeriodID, &mobileMethod, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	settings.MobileMethod = domain.MobileMethod(mobileMethod)
	settings.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	settings.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}
