package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ghgo/internal/domain"
)

func (s *Store) CreateCalculationRun(run domain.CalculationRun) error {
	_, err := s.exec(
		`INSERT INTO calculation_runs (id, organization_id, reporting_period_id, factor_set_id, status, started_at, completed_at, settings_snapshot_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID,
		run.OrganizationID,
		run.ReportingPeriodID,
		run.FactorSetID,
		string(run.Status),
		formatTime(run.StartedAt),
		nullTime(run.CompletedAt),
		run.SettingsSnapshotJSON,
	)
	if err != nil {
		return fmt.Errorf("create calculation run: %w", err)
	}
	return nil
}

func (s *Store) CompleteCalculationRun(id domain.ID, completedAt time.Time) error {
	result, err := s.exec(
		`UPDATE calculation_runs SET status = ?, completed_at = ? WHERE id = ?`,
		string(domain.CalculationRunStatusCompleted),
		formatTime(completedAt),
		id,
	)
	if err != nil {
		return fmt.Errorf("complete calculation run: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete calculation run rows affected: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) GetCalculationRun(id domain.ID) (*domain.CalculationRun, error) {
	run, err := scanCalculationRun(s.queryRow(
		`SELECT id, organization_id, reporting_period_id, factor_set_id, status, started_at, completed_at, settings_snapshot_json
FROM calculation_runs WHERE id = ?`,
		id,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get calculation run: %w", err)
	}
	return run, nil
}

func (s *Store) ListCalculationRunsByPeriod(reportingPeriodID domain.ID) ([]domain.CalculationRun, error) {
	rows, err := s.query(
		`SELECT id, organization_id, reporting_period_id, factor_set_id, status, started_at, completed_at, settings_snapshot_json
FROM calculation_runs WHERE reporting_period_id = ? ORDER BY started_at, id`,
		reportingPeriodID,
	)
	if err != nil {
		return nil, fmt.Errorf("list calculation runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.CalculationRun
	for rows.Next() {
		run, err := scanCalculationRun(rows)
		if err != nil {
			return nil, fmt.Errorf("scan calculation run: %w", err)
		}
		runs = append(runs, *run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list calculation runs rows: %w", err)
	}

	return runs, nil
}

func scanCalculationRun(row scanner) (*domain.CalculationRun, error) {
	var run domain.CalculationRun
	var status string
	var startedAt string
	var completedAt sql.NullString
	if err := row.Scan(
		&run.ID,
		&run.OrganizationID,
		&run.ReportingPeriodID,
		&run.FactorSetID,
		&status,
		&startedAt,
		&completedAt,
		&run.SettingsSnapshotJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	run.Status = domain.CalculationRunStatus(status)
	run.StartedAt, err = parseTime(startedAt)
	if err != nil {
		return nil, err
	}
	run.CompletedAt, err = timePtrFromNull(completedAt)
	if err != nil {
		return nil, err
	}

	return &run, nil
}
