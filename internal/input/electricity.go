package input

import (
	"strconv"

	"ghgo/internal/vocab"
)

func parseElectricityRow(row *ParsedRow, state *parseState) {
	parseMonthlyAmountRow(row, state, vocab.UnitKWh)
}

func parseMonthlyAmountRow(row *ParsedRow, state *parseState, unit vocab.Unit) {
	if !requireColumnCount(row, 2) {
		return
	}

	monthText := row.RawFields[0]
	amountText := row.RawFields[1]

	if requireField(row, "month", monthText) {
		month, ok := vocab.ParseMonth(monthText)
		if !ok {
			row.Errors = append(row.Errors, issue(CodeInvalidMonth, "Month is not recognized."))
		} else {
			monthNumber := month.Number()
			if state.seenMonths[monthNumber] {
				row.Errors = append(row.Errors, issue(CodeDuplicateMonth, "Month appears more than once."))
			}
			state.seenMonths[monthNumber] = true

			row.Normalized["month_number"] = strconv.Itoa(monthNumber)
			row.Normalized["month_name"] = month.EnglishName()
		}
	}

	addAmount(row, amountText, unit)
}
