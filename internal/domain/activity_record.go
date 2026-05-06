package domain

import "time"

type ActivitySourceKind string

const (
	ActivitySourceKindElectricityMonthlyKWh ActivitySourceKind = "electricity_monthly_kwh"
	ActivitySourceKindNaturalGasMonthlySMC  ActivitySourceKind = "natural_gas_monthly_smc"
	ActivitySourceKindMobileFuelLitres      ActivitySourceKind = "mobile_fuel_litres"
	ActivitySourceKindVehicleDistanceKM     ActivitySourceKind = "vehicle_distance_km"
	ActivitySourceKindRefrigerantsAnnualKG  ActivitySourceKind = "refrigerants_annual_kg"
)

func (k ActivitySourceKind) Valid() bool {
	switch k {
	case ActivitySourceKindElectricityMonthlyKWh,
		ActivitySourceKindNaturalGasMonthlySMC,
		ActivitySourceKindMobileFuelLitres,
		ActivitySourceKindVehicleDistanceKM,
		ActivitySourceKindRefrigerantsAnnualKG:
		return true
	}
	return false
}

type Scope int

const (
	Scope1 Scope = 1
	Scope2 Scope = 2
)

func (s Scope) Valid() bool {
	switch s {
	case Scope1, Scope2:
		return true
	}
	return false
}

type ActivityVector string

const (
	ActivityVectorElectricity      ActivityVector = "electricity"
	ActivityVectorNaturalGas       ActivityVector = "natural_gas"
	ActivityVectorMobileCombustion ActivityVector = "mobile_combustion"
	ActivityVectorRefrigerants     ActivityVector = "refrigerants"
)

func (v ActivityVector) Valid() bool {
	switch v {
	case ActivityVectorElectricity,
		ActivityVectorNaturalGas,
		ActivityVectorMobileCombustion,
		ActivityVectorRefrigerants:
		return true
	}
	return false
}

type ActivityMethod string

const (
	ActivityMethodLocationBased ActivityMethod = "location_based"
	ActivityMethodMarketBased   ActivityMethod = "market_based"
	ActivityMethodFuelBased     ActivityMethod = "fuel_based"
	ActivityMethodDistanceBased ActivityMethod = "distance_based"
	ActivityMethodDirectGWP     ActivityMethod = "direct_gwp"
)

func (m ActivityMethod) Valid() bool {
	switch m {
	case ActivityMethodLocationBased,
		ActivityMethodMarketBased,
		ActivityMethodFuelBased,
		ActivityMethodDistanceBased,
		ActivityMethodDirectGWP:
		return true
	}
	return false
}

type ActivityRecordStatus string

const (
	ActivityRecordStatusActive     ActivityRecordStatus = "active"
	ActivityRecordStatusSuperseded ActivityRecordStatus = "superseded"
	ActivityRecordStatusDeleted    ActivityRecordStatus = "deleted"
)

func (s ActivityRecordStatus) Valid() bool {
	switch s {
	case ActivityRecordStatusActive, ActivityRecordStatusSuperseded, ActivityRecordStatusDeleted:
		return true
	}
	return false
}

type ActivityRecord struct {
	ID                ID
	OrganizationID    ID
	FacilityID        *ID
	ReportingPeriodID ID

	SourceKind   ActivitySourceKind
	Scope        Scope
	Vector       ActivityVector
	Category     string
	Method       ActivityMethod
	ActivityType string

	PeriodStart time.Time
	PeriodEnd   time.Time

	Amount float64
	Unit   string

	FuelType         string
	VehicleName      string
	Plate            string
	VehicleType      string
	VehicleSizeClass string
	Substance        string

	Status     ActivityRecordStatus
	SourceHash string

	CreatedAt time.Time
	UpdatedAt time.Time
}
