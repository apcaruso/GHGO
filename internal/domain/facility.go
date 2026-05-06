package domain

import "time"

type Facility struct {
	ID             ID
	OrganizationID ID
	Name           string
	CountryCode    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
