package input

import (
	"strings"

	"ghgo/internal/vocab"
)

func parseVehicleDistanceRow(row *ParsedRow, state *parseState) {
	if !requireColumnCount(row, 6) {
		return
	}

	vehicleName := row.RawFields[0]
	plate := row.RawFields[1]
	vehicleTypeText := row.RawFields[2]
	sizeText := row.RawFields[3]
	fuelText := row.RawFields[4]
	amountText := row.RawFields[5]

	if requireField(row, "vehicle name", vehicleName) {
		row.Normalized["vehicle_name"] = vehicleName
	}

	if requireField(row, "plate", plate) {
		row.Normalized["plate"] = plate
		plateKey := strings.ToUpper(strings.ReplaceAll(plate, " ", ""))
		if state.seenPlates[plateKey] {
			row.Warnings = append(row.Warnings, issue(CodeDuplicatePlate, "Plate appears more than once."))
		}
		state.seenPlates[plateKey] = true
	}

	var vehicleType vocab.VehicleType
	vehicleTypeOK := false
	if requireField(row, "vehicle type", vehicleTypeText) {
		var ok bool
		vehicleType, ok = vocab.NormalizeVehicleType(vehicleTypeText)
		if !ok {
			row.Errors = append(row.Errors, issue(CodeInvalidVehicleType, "Vehicle type is not recognized."))
		} else {
			vehicleTypeOK = true
			row.Normalized["vehicle_type"] = string(vehicleType)
		}
	}

	if requireField(row, "vehicle size class", sizeText) && vehicleTypeOK {
		size, ok := vocab.NormalizeVehicleSizeClass(vehicleType, sizeText)
		if !ok {
			if _, any := normalizeAnyVehicleSizeClass(sizeText); any {
				row.Errors = append(row.Errors, issue(CodeIncompatibleVehicleSizeClass, "Vehicle size class is not compatible with vehicle type."))
			} else {
				row.Errors = append(row.Errors, issue(CodeInvalidVehicleSizeClass, "Vehicle size class is not recognized."))
			}
		} else {
			row.Normalized["vehicle_size_class"] = string(size)
		}
	}

	if requireField(row, "fuel type", fuelText) {
		fuel, ok := vocab.NormalizeFuelType(fuelText)
		if !ok {
			row.Errors = append(row.Errors, issue(CodeInvalidFuelType, "Fuel type is not recognized."))
		} else {
			row.Normalized["fuel_type"] = string(fuel)
			addVehicleFuelWarnings(row, fuel)
		}
	}

	addAmount(row, amountText, vocab.UnitKm)
}

func normalizeAnyVehicleSizeClass(s string) (vocab.VehicleSizeClass, bool) {
	for _, vehicleType := range []vocab.VehicleType{vocab.VehicleCar, vocab.VehicleVan, vocab.VehicleMotorbike} {
		if size, ok := vocab.NormalizeVehicleSizeClass(vehicleType, s); ok {
			return size, true
		}
	}
	return "", false
}

func addVehicleFuelWarnings(row *ParsedRow, fuel vocab.FuelType) {
	switch fuel {
	case vocab.FuelUnknown:
		row.Warnings = append(row.Warnings, issue(CodeUnknownFuelType, "Fuel type is unknown."))
	case vocab.FuelBEV:
		row.Warnings = append(row.Warnings, issue(CodeBEVScope2NotEstimated, "Battery electric vehicle has no direct Scope 1 tailpipe emissions. Charging electricity is not estimated from km."))
	case vocab.FuelPHEV:
		row.Warnings = append(row.Warnings, issue(CodePHEVAverageFactor, "Plug-in hybrid vehicle distance uses an average distance-based factor."))
	}
}
