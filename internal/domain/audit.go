package domain

import "time"

type AuditEvent struct {
	ID             ID
	OrganizationID ID
	EntityType     string
	EntityID       ID
	Action         string
	PayloadJSON    string
	CreatedAt      time.Time
}
