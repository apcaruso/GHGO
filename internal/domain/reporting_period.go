package domain

import "time"

type ReportingPeriodStatus string

const (
	ReportingPeriodStatusDraft    ReportingPeriodStatus = "draft"
	ReportingPeriodStatusLocked   ReportingPeriodStatus = "locked"
	ReportingPeriodStatusArchived ReportingPeriodStatus = "archived"
)

func (s ReportingPeriodStatus) Valid() bool {
	switch s {
	case ReportingPeriodStatusDraft, ReportingPeriodStatusLocked, ReportingPeriodStatusArchived:
		return true
	}
	return false
}

type ReportingPeriod struct {
	ID             ID
	OrganizationID ID
	Year           int
	StartsOn       time.Time
	EndsOn         time.Time
	Status         ReportingPeriodStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
