package domain

import "time"

type GOCoverage string

const (
	GOCoverageNone GOCoverage = "none"
	GOCoverageFull GOCoverage = "full"
)

func (c GOCoverage) Valid() bool {
	switch c {
	case GOCoverageNone, GOCoverageFull:
		return true
	}
	return false
}

type ElectricitySettings struct {
	ID                    ID
	OrganizationID        ID
	ReportingPeriodID     ID
	FacilityID            ID
	HasGuaranteesOfOrigin bool
	GOCoverage            GOCoverage
	GOReference           string
	GOMarket              string
	GOCancelledAt         *time.Time
	EvidenceFileID        *ID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
