package store

import (
	"database/sql"
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreatePasteRow(row domain.PasteRow) error {
	_, err := s.exec(
		`INSERT INTO paste_rows (
  id, paste_batch_id, row_number, raw_json, normalized_json, status,
  errors_json, warnings_json, activity_record_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.PasteBatchID,
		row.RowNumber,
		row.RawJSON,
		row.NormalizedJSON,
		string(row.Status),
		row.ErrorsJSON,
		row.WarningsJSON,
		nullStringPtr(row.ActivityRecordID),
	)
	if err != nil {
		return fmt.Errorf("create paste row: %w", err)
	}
	return nil
}

func (s *Store) ListPasteRowsByBatch(pasteBatchID domain.ID) ([]domain.PasteRow, error) {
	rows, err := s.query(
		`SELECT id, paste_batch_id, row_number, raw_json, normalized_json, status, errors_json, warnings_json, activity_record_id
FROM paste_rows WHERE paste_batch_id = ? ORDER BY row_number, id`,
		pasteBatchID,
	)
	if err != nil {
		return nil, fmt.Errorf("list paste rows: %w", err)
	}
	defer rows.Close()

	var pasteRows []domain.PasteRow
	for rows.Next() {
		pasteRow, err := scanPasteRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan paste row: %w", err)
		}
		pasteRows = append(pasteRows, *pasteRow)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list paste rows rows: %w", err)
	}

	return pasteRows, nil
}

func scanPasteRow(row scanner) (*domain.PasteRow, error) {
	var pasteRow domain.PasteRow
	var status string
	var activityRecordID sql.NullString
	if err := row.Scan(
		&pasteRow.ID,
		&pasteRow.PasteBatchID,
		&pasteRow.RowNumber,
		&pasteRow.RawJSON,
		&pasteRow.NormalizedJSON,
		&status,
		&pasteRow.ErrorsJSON,
		&pasteRow.WarningsJSON,
		&activityRecordID,
	); err != nil {
		return nil, err
	}

	pasteRow.Status = domain.PasteRowStatus(status)
	pasteRow.ActivityRecordID = stringPtrFromNull(activityRecordID)

	return &pasteRow, nil
}
