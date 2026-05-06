package domain

import "time"

type MobileMethod string

const (
	MobileMethodFuelBased     MobileMethod = "fuel_based"
	MobileMethodDistanceBased MobileMethod = "distance_based"
)

func (m MobileMethod) Valid() bool {
	switch m {
	case MobileMethodFuelBased, MobileMethodDistanceBased:
		return true
	}
	return false
}

type ReportingPeriodSettings struct {
	ID                ID
	OrganizationID    ID
	ReportingPeriodID ID
	MobileMethod      MobileMethod
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
