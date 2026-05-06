package vocab

type InputKind string

const (
	InputElectricityMonthlyKWh InputKind = "electricity_monthly_kwh"
	InputNaturalGasMonthlySmc  InputKind = "natural_gas_monthly_smc"
	InputMobileFuelLitres      InputKind = "mobile_fuel_litres"
	InputVehicleDistanceKm     InputKind = "vehicle_distance_km"
	InputRefrigerantsAnnualKg  InputKind = "refrigerants_annual_kg"
)

func (i InputKind) Valid() bool {
	switch i {
	case InputElectricityMonthlyKWh,
		InputNaturalGasMonthlySmc,
		InputMobileFuelLitres,
		InputVehicleDistanceKm,
		InputRefrigerantsAnnualKg:
		return true
	}
	return false
}
