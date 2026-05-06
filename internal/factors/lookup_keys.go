package factors

import (
	"strings"

	"ghgo/internal/vocab"
)

var requiredDEFRAColumns = []string{
	"scope",
	"level_1",
	"level_2",
	"level_3",
	"level_4",
	"column_text",
	"uom",
	"ghg_unit",
	"conversion_factor",
}

var defraHeaderAliases = map[string]string{
	"scope":                 "scope",
	"level 1":               "level_1",
	"level1":                "level_1",
	"level 2":               "level_2",
	"level2":                "level_2",
	"level 3":               "level_3",
	"level3":                "level_3",
	"level 4":               "level_4",
	"level4":                "level_4",
	"column text":           "column_text",
	"columntext":            "column_text",
	"uom":                   "uom",
	"unit":                  "uom",
	"units":                 "uom",
	"unit of measure":       "uom",
	"ghg unit":              "ghg_unit",
	"ghgunit":               "ghg_unit",
	"ghg per unit":          "ghg_unit",
	"ghg conversion factor": "conversion_factor",
	"conversion factor":     "conversion_factor",
	"conversionfactor":      "conversion_factor",
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '/', '-', '_', '\\', '.', ',', ';', ':', '(', ')', '[', ']', '{', '}', '\'', '"':
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func mapHeaderColumns(cells []string) (map[string]int, []string) {
	columns := make(map[string]int)
	for i, cell := range cells {
		canonical, ok := canonicalHeader(cell)
		if !ok {
			continue
		}
		if _, exists := columns[canonical]; !exists {
			columns[canonical] = i
		}
	}

	var missing []string
	for _, required := range requiredDEFRAColumns {
		if _, ok := columns[required]; !ok {
			missing = append(missing, required)
		}
	}
	return columns, missing
}

func canonicalHeader(s string) (string, bool) {
	normalized := normalizeHeader(s)
	if canonical, ok := defraHeaderAliases[normalized]; ok {
		return canonical, true
	}
	if strings.Contains(normalized, "ghg") && strings.Contains(normalized, "conversion factor") {
		return "conversion_factor", true
	}
	return "", false
}

func parseScope(s string) (int, bool) {
	key := vocab.Key(s)
	switch key {
	case "1", "scope 1":
		return 1, true
	case "2", "scope 2":
		return 2, true
	}
	return 0, false
}

func unitFromUOM(s string) (string, bool) {
	key := vocab.Key(s)
	switch {
	case strings.Contains(key, "mwh"):
		return "", false
	case key == "kwh" || strings.HasPrefix(key, "kwh "):
		return string(vocab.UnitKWh), true
	case key == "l" || key == "litre" || key == "liter" || key == "litres" || key == "liters" || key == "lt":
		return string(vocab.UnitLitre), true
	case key == "km" || key == "kilometre" || key == "kilometer" || key == "kilometres" || key == "kilometers":
		return string(vocab.UnitKm), true
	case key == "kg" || key == "kilogram" || key == "kilograms":
		return string(vocab.UnitKg), true
	case key == "smc" || key == "scm" || key == "sm3" || strings.Contains(key, "standard cubic metre") || strings.Contains(key, "standard cubic meter"):
		return string(vocab.UnitSmc), true
	}
	return "", false
}

func isKgCO2e(s string) bool {
	key := strings.Join(strings.Fields(strings.ToLower(s)), " ")
	if strings.Contains(key, " of ") {
		return false
	}
	for _, old := range []string{" ", "/", "-", "_", ".", "(", ")"} {
		key = strings.ReplaceAll(key, old, "")
	}
	if strings.Contains(key, "tonne") || strings.Contains(key, "tco2e") {
		return false
	}
	return strings.Contains(key, "kgco2e")
}

func rowKey(row defraRow) string {
	return vocab.Key(strings.Join([]string{row.Level1, row.Level2, row.Level3, row.Level4, row.ColumnText}, " "))
}

func isWTT(row defraRow) bool {
	key := rowKey(row)
	return strings.Contains(key, "wtt") || strings.Contains(key, "well to tank")
}

func isOutsideScopes(row defraRow) bool {
	key := rowKey(row)
	return strings.Contains(key, "outside of scopes") || strings.Contains(key, "outside scopes")
}
