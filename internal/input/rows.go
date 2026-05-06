package input

import (
	"regexp"
	"strings"

	"ghgo/internal/vocab"
)

type rawRow struct {
	number int
	fields []string
}

var multiSpaceDelimiter = regexp.MustCompile(` {2,}`)

func splitRows(raw string) []rawRow {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	lines := strings.Split(raw, "\n")
	rows := make([]rawRow, 0, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		rows = append(rows, rawRow{number: i + 1, fields: splitFields(line)})
	}

	return rows
}

func splitFields(line string) []string {
	var parts []string
	switch {
	case strings.Contains(line, "\t"):
		parts = strings.Split(line, "\t")
	case strings.Contains(line, "|"):
		parts = strings.Split(line, "|")
	case strings.Contains(line, ";"):
		parts = strings.Split(line, ";")
	default:
		line = strings.TrimSpace(line)
		if multiSpaceDelimiter.MatchString(line) {
			parts = multiSpaceDelimiter.Split(line, -1)
		} else {
			parts = []string{line}
		}
	}

	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		fields = append(fields, vocab.NormalizeSpace(part))
	}
	return fields
}

func isHeader(inputKind vocab.InputKind, fields []string) bool {
	keys := fieldKeys(fields)
	switch inputKind {
	case vocab.InputElectricityMonthlyKWh, vocab.InputNaturalGasMonthlySmc:
		return hasHeader(keys, [][]string{
			{"month", "consumption"},
			{"month", "kwh"},
			{"month", "smc"},
			{"mese", "consumo"},
			{"mese", "kwh"},
			{"mese", "smc"},
		})
	case vocab.InputMobileFuelLitres:
		return hasHeader(keys, [][]string{
			{"fuel type", "litres"},
			{"fuel", "litres"},
			{"carburante", "litri"},
		})
	case vocab.InputVehicleDistanceKm:
		return hasHeader(keys, [][]string{
			{"vehicle name", "plate", "vehicle type", "size class", "fuel type", "km"},
			{"vehicle", "plate", "vehicle type", "size class", "fuel type", "km"},
			{"mezzo", "targa", "classe", "dimensione", "carburante", "km"},
		})
	case vocab.InputRefrigerantsAnnualKg:
		return hasHeader(keys, [][]string{
			{"gas", "kg"},
			{"refrigerant", "kg"},
		})
	}
	return false
}

func fieldKeys(fields []string) []string {
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		keys = append(keys, vocab.Key(field))
	}
	return keys
}

func hasHeader(got []string, allowed [][]string) bool {
	for _, want := range allowed {
		if sameStrings(got, want) {
			return true
		}
	}
	return false
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
