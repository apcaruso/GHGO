package input

import "ghgo/internal/vocab"

func parseRefrigerantsRow(row *ParsedRow, state *parseState) {
	if !requireColumnCount(row, 2) {
		return
	}

	refrigerantText := row.RawFields[0]
	amountText := row.RawFields[1]

	if requireField(row, "refrigerant", refrigerantText) {
		refrigerant, ok := vocab.NormalizeRefrigerant(refrigerantText)
		if !ok {
			row.Errors = append(row.Errors, issue(CodeInvalidRefrigerant, "Refrigerant is not recognized."))
		} else {
			if state.seenRefrigerants[refrigerant] {
				row.Warnings = append(row.Warnings, issue(CodeDuplicateRefrigerant, "Refrigerant appears more than once."))
			}
			state.seenRefrigerants[refrigerant] = true

			row.Normalized["substance"] = string(refrigerant)
		}
	}

	addAmount(row, amountText, vocab.UnitKg)
}
