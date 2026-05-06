package domain

import "time"

type FactorSet struct {
	ID           ID
	Name         string
	Source       string
	Year         int
	Version      string
	ImportedAt   time.Time
	MetadataJSON string
}

type EmissionFactor struct {
	ID          ID
	FactorSetID ID

	Source     string
	Scope      Scope
	Level1     string
	Level2     string
	Level3     string
	Level4     string
	ColumnText string

	ActivityType     string
	FuelType         string
	VehicleType      string
	VehicleSizeClass string
	Substance        string

	InputUnit   string
	FactorUnit  string
	GHG         string
	FactorValue float64

	MetadataJSON string
}
