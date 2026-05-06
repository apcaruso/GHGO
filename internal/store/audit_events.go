package store

import (
	"fmt"

	"ghgo/internal/domain"
)

func (s *Store) CreateAuditEvent(event domain.AuditEvent) error {
	_, err := s.exec(
		`INSERT INTO audit_events (id, organization_id, entity_type, entity_id, action, payload_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		event.OrganizationID,
		event.EntityType,
		event.EntityID,
		event.Action,
		event.PayloadJSON,
		formatTime(event.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}

func (s *Store) ListAuditEventsByEntity(entityType string, entityID domain.ID) ([]domain.AuditEvent, error) {
	rows, err := s.query(
		`SELECT id, organization_id, entity_type, entity_id, action, payload_json, created_at
FROM audit_events WHERE entity_type = ? AND entity_id = ? ORDER BY created_at, id`,
		entityType,
		entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	var events []domain.AuditEvent
	for rows.Next() {
		event, err := scanAuditEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		events = append(events, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list audit events rows: %w", err)
	}

	return events, nil
}

func scanAuditEvent(row scanner) (*domain.AuditEvent, error) {
	var event domain.AuditEvent
	var createdAt string
	if err := row.Scan(
		&event.ID,
		&event.OrganizationID,
		&event.EntityType,
		&event.EntityID,
		&event.Action,
		&event.PayloadJSON,
		&createdAt,
	); err != nil {
		return nil, err
	}

	var err error
	event.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
