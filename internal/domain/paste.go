package domain

import "time"

type PasteBatchStatus string

const (
	PasteBatchStatusParsed    PasteBatchStatus = "parsed"
	PasteBatchStatusCommitted PasteBatchStatus = "committed"
	PasteBatchStatusFailed    PasteBatchStatus = "failed"
)

func (s PasteBatchStatus) Valid() bool {
	switch s {
	case PasteBatchStatusParsed, PasteBatchStatusCommitted, PasteBatchStatusFailed:
		return true
	}
	return false
}

type PasteRowStatus string

const (
	PasteRowStatusValid     PasteRowStatus = "valid"
	PasteRowStatusWarning   PasteRowStatus = "warning"
	PasteRowStatusError     PasteRowStatus = "error"
	PasteRowStatusCommitted PasteRowStatus = "committed"
)

func (s PasteRowStatus) Valid() bool {
	switch s {
	case PasteRowStatusValid, PasteRowStatusWarning, PasteRowStatusError, PasteRowStatusCommitted:
		return true
	}
	return false
}

type PasteBatch struct {
	ID                ID
	OrganizationID    ID
	ReportingPeriodID ID
	InputKind         string
	ContextJSON       string
	RawText           string
	RawHash           string
	Status            PasteBatchStatus
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
	Status           PasteRowStatus
	ErrorsJSON       string
	WarningsJSON     string
	ActivityRecordID *ID
}
