package input

import "ghgo/internal/vocab"

func parseNaturalGasRow(row *ParsedRow, state *parseState) {
	parseMonthlyAmountRow(row, state, vocab.UnitSmc)
}
