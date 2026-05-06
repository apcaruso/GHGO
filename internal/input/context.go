package input

import (
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/vocab"
)

type CommitContext struct {
	OrganizationID    string
	ReportingPeriodID string
	FacilityID        *string

	ReportingYear int
	PeriodStart   time.Time
	PeriodEnd     time.Time

	InputKind vocab.InputKind

	MobileMethod domain.MobileMethod

	HasGuaranteesOfOrigin bool
	GOCoverage            domain.GOCoverage
}

type CommitResult struct {
	PasteBatchID      string
	ActivityRecordIDs []string
	RowsTotal         int
	RowsValid         int
	RowsError         int
	Committed         bool
}
