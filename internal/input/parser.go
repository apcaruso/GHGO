package input

import (
	"strconv"
	"strings"

	"ghgo/internal/vocab"
)

type ParseResult struct {
	InputKind string
	RawText   string
	Rows      []ParsedRow
	RowsTotal int
	RowsValid int
	RowsError int
}

type ParsedRow struct {
	RowNumber  int
	RawFields  []string
	Normalized map[string]string
	Errors     []ParseIssue
	Warnings   []ParseIssue
}

type parseState struct {
	seenMonths       map[int]bool
	seenFuelTypes    map[vocab.FuelType]bool
	seenPlates       map[string]bool
	seenRefrigerants map[vocab.Refrigerant]bool
}

func Parse(inputKind vocab.InputKind, raw string) ParseResult {
	result := ParseResult{InputKind: string(inputKind), RawText: raw}
	state := parseState{
		seenMonths:       make(map[int]bool),
		seenFuelTypes:    make(map[vocab.FuelType]bool),
		seenPlates:       make(map[string]bool),
		seenRefrigerants: make(map[vocab.Refrigerant]bool),
	}

	headerSkipped := false
	seenBody := false
	for _, rawRow := range splitRows(raw) {
		if isHeader(inputKind, rawRow.fields) {
			if !headerSkipped && !seenBody {
				headerSkipped = true
				continue
			}

			row := newParsedRow(rawRow)
			row.Errors = append(row.Errors, issue(CodeHeaderInBody, "Header row is only allowed as the first non-empty row."))
			result.addRow(row)
			seenBody = true
			continue
		}

		row := newParsedRow(rawRow)
		switch inputKind {
		case vocab.InputElectricityMonthlyKWh:
			parseElectricityRow(&row, &state)
		case vocab.InputNaturalGasMonthlySmc:
			parseNaturalGasRow(&row, &state)
		case vocab.InputMobileFuelLitres:
			parseMobileFuelRow(&row, &state)
		case vocab.InputVehicleDistanceKm:
			parseVehicleDistanceRow(&row, &state)
		case vocab.InputRefrigerantsAnnualKg:
			parseRefrigerantsRow(&row, &state)
		default:
			return result
		}

		result.addRow(row)
		seenBody = true
	}

	return result
}

func newParsedRow(raw rawRow) ParsedRow {
	return ParsedRow{
		RowNumber:  raw.number,
		RawFields:  raw.fields,
		Normalized: make(map[string]string),
	}
}

func (r *ParseResult) addRow(row ParsedRow) {
	r.Rows = append(r.Rows, row)
	r.RowsTotal++
	if len(row.Errors) > 0 {
		r.RowsError++
		return
	}
	r.RowsValid++
}

func requireField(row *ParsedRow, fieldName, value string) bool {
	if strings.TrimSpace(value) != "" {
		return true
	}
	row.Errors = append(row.Errors, issue(CodeEmptyField, "Required field "+fieldName+" is empty."))
	return false
}

func requireColumnCount(row *ParsedRow, want int) bool {
	if len(row.RawFields) == want {
		return true
	}
	row.Errors = append(row.Errors, issue(CodeWrongColumnCount, "Wrong number of columns."))
	return false
}

func addAmount(row *ParsedRow, value string, unit vocab.Unit) bool {
	if !requireField(row, "amount", value) {
		return false
	}

	amount, issues, ok := ParsePositiveNumber(value)
	if !ok {
		row.Errors = append(row.Errors, issues...)
		return false
	}

	row.Warnings = append(row.Warnings, issues...)
	row.Normalized["amount"] = formatAmount(amount)
	row.Normalized["unit"] = string(unit)
	return true
}

func formatAmount(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
