package domain

import "time"

type PasteBatch struct {
	ID                ID
	OrganizationID    ID
	ReportingPeriodID ID
	InputKind         string
	ContextJSON       string
	RawText           string
	RawHash           string
	RowsTotal         int
	RowsValid         int
	RowsError         int
	CreatedAt         time.Time
	CommittedAt       *time.Time
}

type PasteRow struct {
	ID               ID
	PasteBatchID     ID
	RowNumber        int
	RawJSON          string
	NormalizedJSON   string
	ErrorsJSON       string
	WarningsJSON     string
	ActivityRecordID *ID
}
