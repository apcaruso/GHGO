//go:build ghgo_devtools

package factors

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type parsedDEFRAFactors struct {
	Candidates  []factorCandidate
	RowsRead    int
	RowsSkipped int
	Warnings    []string
}

func parseDEFRA2025File(path string) (parsedDEFRAFactors, error) {
	var parsed parsedDEFRAFactors

	workbook, err := excelize.OpenFile(path)
	if err != nil {
		return parsed, fmt.Errorf("open DEFRA 2025 workbook: %w", err)
	}
	defer workbook.Close()

	sheet, headerRow, columns, err := detectDEFRAFlatSheet(workbook)
	if err != nil {
		return parsed, err
	}

	rows, err := workbook.GetRows(sheet)
	if err != nil {
		return parsed, fmt.Errorf("read DEFRA sheet %q: %w", sheet, err)
	}

	for i := headerRow + 1; i < len(rows); i++ {
		cells := rows[i]
		if rowEmpty(cells) {
			continue
		}

		parsed.RowsRead++
		row := defraRowFromCells(cells, columns)

		scope, ok := parseScope(row.OriginalScope)
		if !ok {
			continue
		}
		row.Scope = scope

		factor, ok := parseConversionFactor(row.ConversionFactorText)
		if !ok {
			continue
		}
		row.ConversionFactor = factor

		candidate, ok := mapDEFRA2025Row(row)
		if !ok {
			continue
		}
		parsed.Candidates = append(parsed.Candidates, *candidate)
	}

	parsed.RowsSkipped = parsed.RowsRead - len(parsed.Candidates)
	if !hasActivity(parsed.Candidates, "natural_gas") {
		parsed.Warnings = append(parsed.Warnings, "no Smc-compatible natural gas factor found in DEFRA/DESNZ 2025")
	}

	return parsed, nil
}

func detectDEFRAFlatSheet(workbook *excelize.File) (string, int, map[string]int, error) {
	var bestMissing []string
	bestMatches := -1

	for _, sheet := range workbook.GetSheetList() {
		rows, err := workbook.GetRows(sheet)
		if err != nil {
			return "", 0, nil, fmt.Errorf("read DEFRA sheet %q: %w", sheet, err)
		}

		limit := len(rows)
		if limit > 25 {
			limit = 25
		}
		for rowIndex := 0; rowIndex < limit; rowIndex++ {
			columns, missing := mapHeaderColumns(rows[rowIndex])
			matches := len(requiredDEFRAColumns) - len(missing)
			if matches > bestMatches {
				bestMatches = matches
				bestMissing = missing
			}
			if len(missing) == 0 {
				return sheet, rowIndex, columns, nil
			}
		}
	}

	if bestMissing != nil {
		return "", 0, nil, fmt.Errorf("missing required DEFRA column(s): %s", strings.Join(bestMissing, ", "))
	}
	return "", 0, nil, fmt.Errorf("no sheets found in DEFRA workbook")
}

func defraRowFromCells(cells []string, columns map[string]int) defraRow {
	return defraRow{
		OriginalScope:        cell(cells, columns["scope"]),
		Level1:               cell(cells, columns["level_1"]),
		Level2:               cell(cells, columns["level_2"]),
		Level3:               cell(cells, columns["level_3"]),
		Level4:               cell(cells, columns["level_4"]),
		ColumnText:           cell(cells, columns["column_text"]),
		UOM:                  cell(cells, columns["uom"]),
		GHGUnit:              cell(cells, columns["ghg_unit"]),
		ConversionFactorText: cell(cells, columns["conversion_factor"]),
	}
}

func parseConversionFactor(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	s = strings.ReplaceAll(s, ",", "")
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

func cell(cells []string, index int) string {
	if index < 0 || index >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[index])
}

func rowEmpty(cells []string) bool {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func hasActivity(candidates []factorCandidate, activityType string) bool {
	for _, candidate := range candidates {
		if candidate.ActivityType == activityType {
			return true
		}
	}
	return false
}
