package store

import (
	"database/sql"
	"errors"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) UpsertElectricitySettings(settings domain.ElectricitySettings) error {
	_, err := s.exec(
		`INSERT INTO electricity_settings (
  id, organization_id, reporting_period_id, facility_id, has_guarantees_of_origin,
  go_coverage, go_reference, go_market, go_cancelled_at, evidence_file_id, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(reporting_period_id, facility_id) DO UPDATE SET
  id = excluded.id,
  organization_id = excluded.organization_id,
  has_guarantees_of_origin = excluded.has_guarantees_of_origin,
  go_coverage = excluded.go_coverage,
  go_reference = excluded.go_reference,
  go_market = excluded.go_market,
  go_cancelled_at = excluded.go_cancelled_at,
  evidence_file_id = excluded.evidence_file_id,
  updated_at = excluded.updated_at`,
		settings.ID,
		settings.OrganizationID,
		settings.ReportingPeriodID,
		settings.FacilityID,
		boolToInt(settings.HasGuaranteesOfOrigin),
		string(settings.GOCoverage),
		nullString(settings.GOReference),
		nullString(settings.GOMarket),
		nullTime(settings.GOCancelledAt),
		nullStringPtr(settings.EvidenceFileID),
		formatTime(settings.CreatedAt),
		formatTime(settings.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert electricity settings: %w", err)
	}
	return nil
}

func (s *Store) GetElectricitySettings(reportingPeriodID, facilityID domain.ID) (*domain.ElectricitySettings, error) {
	settings, err := scanElectricitySettings(s.queryRow(
		`SELECT id, organization_id, reporting_period_id, facility_id, has_guarantees_of_origin,
go_coverage, go_reference, go_market, go_cancelled_at, evidence_file_id, created_at, updated_at
FROM electricity_settings WHERE reporting_period_id = ? AND facility_id = ?`,
		reportingPeriodID,
		facilityID,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get electricity settings: %w", err)
	}
	return settings, nil
}

func scanElectricitySettings(row scanner) (*domain.ElectricitySettings, error) {
	var settings domain.ElectricitySettings
	var hasGO int
	var coverage string
	var goReference sql.NullString
	var goMarket sql.NullString
	var goCancelledAt sql.NullString
	var evidenceFileID sql.NullString
	var createdAt string
	var updatedAt string
	if err := row.Scan(
		&settings.ID,
		&settings.OrganizationID,
		&settings.ReportingPeriodID,
		&settings.FacilityID,
		&hasGO,
		&coverage,
		&goReference,
		&goMarket,
		&goCancelledAt,
		&evidenceFileID,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	settings.HasGuaranteesOfOrigin = intToBool(hasGO)
	settings.GOCoverage = domain.GOCoverage(coverage)
	settings.GOReference = stringFromNull(goReference)
	settings.GOMarket = stringFromNull(goMarket)
	settings.GOCancelledAt, err = timePtrFromNull(goCancelledAt)
	if err != nil {
		return nil, err
	}
	settings.EvidenceFileID = stringPtrFromNull(evidenceFileID)
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
