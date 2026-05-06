package domain

import "time"

type CalculationRunStatus string

const (
	CalculationRunStatusRunning   CalculationRunStatus = "running"
	CalculationRunStatusCompleted CalculationRunStatus = "completed"
	CalculationRunStatusFailed    CalculationRunStatus = "failed"
)

func (s CalculationRunStatus) Valid() bool {
	switch s {
	case CalculationRunStatusRunning, CalculationRunStatusCompleted, CalculationRunStatusFailed:
		return true
	}
	return false
}

type CalculationRun struct {
	ID                   ID
	OrganizationID       ID
	ReportingPeriodID    ID
	FactorSetID          ID
	Status               CalculationRunStatus
	StartedAt            time.Time
	CompletedAt          *time.Time
	SettingsSnapshotJSON string
}

type CalculationResult struct {
	ID               ID
	CalculationRunID ID
	ActivityRecordID ID

	Scope  Scope
	Vector ActivityVector
	Method ActivityMethod

	ActivityAmount float64
	ActivityUnit   string

	FactorID     *ID
	FactorValue  float64
	FactorUnit   string
	FactorSource string

	EmissionsKgCO2e float64
	IsPrimary       bool

	NotesJSON string
}
