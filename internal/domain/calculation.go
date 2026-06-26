package domain

import "time"

type CalculationRun struct {
	ID                   ID
	OrganizationID       ID
	ReportingPeriodID    ID
	FactorSetID          ID
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
