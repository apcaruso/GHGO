package domain

import "time"

type Organization struct {
	ID        ID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
