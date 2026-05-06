package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ghgo/internal/domain"
)

func (s *Store) CreatePasteBatch(batch domain.PasteBatch) error {
	_, err := s.exec(
		`INSERT INTO paste_batches (
  id, organization_id, reporting_period_id, input_kind, context_json, raw_text, raw_hash,
  status, rows_total, rows_valid, rows_error, created_at, committed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		batch.ID,
		batch.OrganizationID,
		batch.ReportingPeriodID,
		batch.InputKind,
		batch.ContextJSON,
		batch.RawText,
		batch.RawHash,
		string(batch.Status),
		batch.RowsTotal,
		batch.RowsValid,
		batch.RowsError,
		formatTime(batch.CreatedAt),
		nullTime(batch.CommittedAt),
	)
	if err != nil {
		return fmt.Errorf("create paste batch: %w", err)
	}
	return nil
}

func (s *Store) GetPasteBatch(id domain.ID) (*domain.PasteBatch, error) {
	batch, err := scanPasteBatch(s.queryRow(
		`SELECT id, organization_id, reporting_period_id, input_kind, context_json, raw_text, raw_hash,
status, rows_total, rows_valid, rows_error, created_at, committed_at
FROM paste_batches WHERE id = ?`,
		id,
	))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get paste batch: %w", err)
	}
	return batch, nil
}

func (s *Store) ListPasteBatchesByPeriod(reportingPeriodID domain.ID) ([]domain.PasteBatch, error) {
	rows, err := s.query(
		`SELECT id, organization_id, reporting_period_id, input_kind, context_json, raw_text, raw_hash,
status, rows_total, rows_valid, rows_error, created_at, committed_at
FROM paste_batches WHERE reporting_period_id = ? ORDER BY created_at, id`,
		reportingPeriodID,
	)
	if err != nil {
		return nil, fmt.Errorf("list paste batches: %w", err)
	}
	defer rows.Close()

	var batches []domain.PasteBatch
	for rows.Next() {
		batch, err := scanPasteBatch(rows)
		if err != nil {
			return nil, fmt.Errorf("scan paste batch: %w", err)
		}
		batches = append(batches, *batch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list paste batches rows: %w", err)
	}

	return batches, nil
}

func (s *Store) MarkPasteBatchCommitted(id domain.ID, committedAt time.Time) error {
	_, err := s.exec(
		`UPDATE paste_batches SET status = ?, committed_at = ? WHERE id = ?`,
		string(domain.PasteBatchStatusCommitted),
		formatTime(committedAt),
		id,
	)
	if err != nil {
		return fmt.Errorf("mark paste batch committed: %w", err)
	}
	return nil
}

func scanPasteBatch(row scanner) (*domain.PasteBatch, error) {
	var batch domain.PasteBatch
	var status string
	var createdAt string
	var committedAt sql.NullString
	if err := row.Scan(
		&batch.ID,
		&batch.OrganizationID,
		&batch.ReportingPeriodID,
		&batch.InputKind,
		&batch.ContextJSON,
		&batch.RawText,
		&batch.RawHash,
		&status,
		&batch.RowsTotal,
		&batch.RowsValid,
		&batch.RowsError,
		&createdAt,
		&committedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var err error
	batch.Status = domain.PasteBatchStatus(status)
	batch.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	batch.CommittedAt, err = timePtrFromNull(committedAt)
	if err != nil {
		return nil, err
	}

	return &batch, nil
}
