package factors

import (
	"ghgo/internal/domain"
	"ghgo/internal/vocab"
)

func defra2025SeedFactors() []SeedFactor {
	factors := []SeedFactor{
		seedFactor(domain.Scope2, "purchased_electricity", "", "", "", "", string(vocab.UnitKWh), "kgCO2e/kWh", 0.20705, "UK electricity", "Electricity generated", "Electricity: UK", "", "Location-based"),
		seedFactor(domain.Scope1, "natural_gas", "", "", "", "", string(vocab.UnitSmc), "kgCO2e/Smc", 2.02135, "Fuels", "Gaseous fuels", "Natural gas", "", "Gross CV - standard cubic metre"),
	}

	for _, row := range []struct {
		fuel  vocab.FuelType
		value float64
	}{
		{vocab.FuelDiesel, 2.51233},
		{vocab.FuelPetrol, 2.17642},
		{vocab.FuelLPG, 1.55713},
		{vocab.FuelCNG, 0.44341},
	} {
		factors = append(factors, seedFactor(domain.Scope1, string(row.fuel)+"_mobile", string(row.fuel), "", "", "", string(vocab.UnitLitre), "kgCO2e/L", row.value, "Fuels", "Mobile combustion", row.fuel.EnglishLabel(), "", "Litres"))
	}

	factors = append(factors, vehicleSeedFactors()...)
	factors = append(factors, refrigerantSeedFactors()...)
	return factors
}

func vehicleSeedFactors() []SeedFactor {
	type vehicleFactor struct {
		vehicleType vocab.VehicleType
		sizeClass   vocab.VehicleSizeClass
		fuelType    vocab.FuelType
		value       float64
	}

	rows := []vehicleFactor{
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelDiesel, 0.13122},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelPetrol, 0.14112},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelLPG, 0.12134},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelCNG, 0.11543},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelHybrid, 0.09582},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelPHEV, 0.07024},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelBEV, 0},
		{vocab.VehicleCar, vocab.SizeSmall, vocab.FuelUnknown, 0.13617},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelDiesel, 0.15847},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelPetrol, 0.17022},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelLPG, 0.14791},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelCNG, 0.13964},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelHybrid, 0.11872},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelPHEV, 0.08935},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelBEV, 0},
		{vocab.VehicleCar, vocab.SizeMedium, vocab.FuelUnknown, 0.16438},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelDiesel, 0.20714},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelPetrol, 0.24189},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelLPG, 0.20782},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelCNG, 0.19763},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelHybrid, 0.16312},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelPHEV, 0.12241},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelBEV, 0},
		{vocab.VehicleCar, vocab.SizeLarge, vocab.FuelUnknown, 0.22452},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelDiesel, 0.16821},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelPetrol, 0.17491},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelLPG, 0.15185},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelCNG, 0.14352},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelHybrid, 0.12175},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelPHEV, 0.09162},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelBEV, 0},
		{vocab.VehicleCar, vocab.SizeAverage, vocab.FuelUnknown, 0.17156},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelDiesel, 0.17611},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelPetrol, 0.20241},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelLPG, 0.17326},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelCNG, 0.16487},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelHybrid, 0.14532},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelPHEV, 0.11243},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelBEV, 0},
		{vocab.VehicleVan, vocab.SizeClassI, vocab.FuelUnknown, 0.18824},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelDiesel, 0.22517},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelPetrol, 0.25433},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelLPG, 0.21982},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelCNG, 0.20954},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelHybrid, 0.18462},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelPHEV, 0.14234},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelBEV, 0},
		{vocab.VehicleVan, vocab.SizeClassII, vocab.FuelUnknown, 0.23975},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelDiesel, 0.30318},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelPetrol, 0.33142},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelLPG, 0.28654},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelCNG, 0.27361},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelHybrid, 0.24117},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelPHEV, 0.18875},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelBEV, 0},
		{vocab.VehicleVan, vocab.SizeClassIII, vocab.FuelUnknown, 0.31842},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelDiesel, 0.25328},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelPetrol, 0.28144},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelLPG, 0.24291},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelCNG, 0.23182},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelHybrid, 0.20435},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelPHEV, 0.15841},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelBEV, 0},
		{vocab.VehicleVan, vocab.SizeAverage, vocab.FuelUnknown, 0.26642},
	}

	factors := make([]SeedFactor, 0, len(rows)+4)
	for _, row := range rows {
		factors = append(factors, seedFactor(domain.Scope1, "vehicle_distance", string(row.fuelType), string(row.vehicleType), string(row.sizeClass), "", string(vocab.UnitKm), "kgCO2e/km", row.value, "Passenger and delivery vehicles", row.vehicleType.EnglishLabel(), row.sizeClass.EnglishLabel(row.vehicleType), row.fuelType.EnglishLabel(), "Distance travelled"))
	}

	for _, row := range []struct {
		size  vocab.VehicleSizeClass
		value float64
	}{
		{vocab.SizeSmall, 0.08321},
		{vocab.SizeMedium, 0.10048},
		{vocab.SizeLarge, 0.13276},
		{vocab.SizeAverage, 0.11324},
	} {
		factors = append(factors, seedFactor(domain.Scope1, "vehicle_distance", "", string(vocab.VehicleMotorbike), string(row.size), "", string(vocab.UnitKm), "kgCO2e/km", row.value, "Passenger vehicles", vocab.VehicleMotorbike.EnglishLabel(), row.size.EnglishLabel(vocab.VehicleMotorbike), "", "Distance travelled"))
	}

	return factors
}

func refrigerantSeedFactors() []SeedFactor {
	return []SeedFactor{
		seedFactor(domain.Scope1, "refrigerant_leakage", "", "", "", string(vocab.RefrigerantR134a), string(vocab.UnitKg), "kgCO2e/kg", 1430, "Refrigerants", "HFCs", string(vocab.RefrigerantR134a), "", "Leakage"),
		seedFactor(domain.Scope1, "refrigerant_leakage", "", "", "", string(vocab.RefrigerantR410A), string(vocab.UnitKg), "kgCO2e/kg", 2088, "Refrigerants", "HFCs", string(vocab.RefrigerantR410A), "", "Leakage"),
		seedFactor(domain.Scope1, "refrigerant_leakage", "", "", "", string(vocab.RefrigerantR407C), string(vocab.UnitKg), "kgCO2e/kg", 1774, "Refrigerants", "HFCs", string(vocab.RefrigerantR407C), "", "Leakage"),
		seedFactor(domain.Scope1, "refrigerant_leakage", "", "", "", string(vocab.RefrigerantR404A), string(vocab.UnitKg), "kgCO2e/kg", 3922, "Refrigerants", "HFCs", string(vocab.RefrigerantR404A), "", "Leakage"),
		seedFactor(domain.Scope1, "refrigerant_leakage", "", "", "", string(vocab.RefrigerantR32), string(vocab.UnitKg), "kgCO2e/kg", 675, "Refrigerants", "HFCs", string(vocab.RefrigerantR32), "", "Leakage"),
		seedFactor(domain.Scope1, "refrigerant_leakage", "", "", "", string(vocab.RefrigerantR22), string(vocab.UnitKg), "kgCO2e/kg", 1810, "Refrigerants", "HCFCs", string(vocab.RefrigerantR22), "", "Leakage"),
	}
}

func seedFactor(scope domain.Scope, activityType string, fuelType string, vehicleType string, vehicleSizeClass string, substance string, inputUnit string, factorUnit string, value float64, level1 string, level2 string, level3 string, level4 string, columnText string) SeedFactor {
	return SeedFactor{
		Scope:            int(scope),
		Source:           defra2025Source,
		Level1:           level1,
		Level2:           level2,
		Level3:           level3,
		Level4:           level4,
		ColumnText:       columnText,
		ActivityType:     activityType,
		FuelType:         fuelType,
		VehicleType:      vehicleType,
		VehicleSizeClass: vehicleSizeClass,
		Substance:        substance,
		InputUnit:        inputUnit,
		FactorUnit:       factorUnit,
		GHG:              "kgCO2e",
		FactorValue:      value,
	}
}
