package input

import "ghgo/internal/vocab"

func parseMobileFuelRow(row *ParsedRow, state *parseState) {
	if !requireColumnCount(row, 2) {
		return
	}

	fuelText := row.RawFields[0]
	amountText := row.RawFields[1]

	if requireField(row, "fuel type", fuelText) {
		fuel, ok := vocab.NormalizeFuelType(fuelText)
		if !ok {
			row.Errors = append(row.Errors, issue(CodeInvalidFuelType, "Fuel type is not recognized."))
		} else {
			if state.seenFuelTypes[fuel] {
				row.Warnings = append(row.Warnings, issue(CodeDuplicateFuelType, "Fuel type appears more than once."))
			}
			state.seenFuelTypes[fuel] = true

			row.Normalized["fuel_type"] = string(fuel)
			if !fuelAllowedForMobileFuel(fuel) {
				row.Errors = append(row.Errors, issue(CodeUnsupportedFuelForMethod, "Fuel type is not supported for this input method."))
			}
		}
	}

	addAmount(row, amountText, vocab.UnitLitre)
}

func fuelAllowedForMobileFuel(fuel vocab.FuelType) bool {
	switch fuel {
	case vocab.FuelDiesel, vocab.FuelPetrol, vocab.FuelLPG, vocab.FuelCNG:
		return true
	}
	return false
}
