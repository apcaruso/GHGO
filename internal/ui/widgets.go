package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/vocab"
)

func simpleTable(headers []string, rows [][]string) fyne.CanvasObject {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var b strings.Builder
	writeRow := func(row []string) {
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			b.WriteString(cell)
			if pad := widths[i] - len(cell) + 2; pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		}
		b.WriteString("\n")
	}
	writeRow(headers)
	separator := make([]string, len(headers))
	for i := range headers {
		separator[i] = strings.Repeat("-", widths[i])
	}
	writeRow(separator)
	for _, row := range rows {
		writeRow(row)
	}
	label := widget.NewLabel(b.String())
	label.TextStyle.Monospace = true
	return container.NewHScroll(label)
}

func parseSummary(parsed input.ParseResult) string {
	warnings := 0
	for _, row := range parsed.Rows {
		warnings += len(row.Warnings)
	}
	return fmt.Sprintf("Rows: %d\nValid: %d\nErrors: %d\nWarnings: %d", parsed.RowsTotal, parsed.RowsValid, parsed.RowsError, warnings)
}

func parsedRowsPreview(inputKind vocab.InputKind, parsed input.ParseResult) fyne.CanvasObject {
	switch inputKind {
	case vocab.InputElectricityMonthlyKWh:
		return parsedMonthlyRowsPreview(parsed, "Consumption kWh")
	case vocab.InputNaturalGasMonthlySmc:
		return parsedMonthlyRowsPreview(parsed, "Consumption Smc")
	case vocab.InputMobileFuelLitres:
		return parsedMobileFuelRowsPreview(parsed)
	case vocab.InputVehicleDistanceKm:
		return parsedVehicleDistanceRowsPreview(parsed)
	case vocab.InputRefrigerantsAnnualKg:
		return parsedRefrigerantRowsPreview(parsed)
	}
	return parsedRawRowsPreview(parsed)
}

func parsedRawRowsPreview(parsed input.ParseResult) fyne.CanvasObject {
	rows := make([][]string, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		status := parsedRowStatus(row)
		rows = append(rows, []string{
			fmt.Sprintf("%d", row.RowNumber),
			status,
			strings.Join(row.RawFields, " | "),
			formatParsedRowIssues(row),
		})
	}
	return simpleTable([]string{"Row", "Status", "Raw fields", "Issues"}, rows)
}

func parsedMonthlyRowsPreview(parsed input.ParseResult, amountHeader string) fyne.CanvasObject {
	rows := make([][]string, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		rows = append(rows, []string{displayField(row, "month_name", 0), displayField(row, "amount", 1), parsedRowStatus(row), formatParsedRowIssues(row)})
	}
	return simpleTable([]string{"Month", amountHeader, "Status", "Issues"}, rows)
}

func parsedMobileFuelRowsPreview(parsed input.ParseResult) fyne.CanvasObject {
	rows := make([][]string, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		rows = append(rows, []string{fuelTypeLabel(displayField(row, "fuel_type", 0)), displayField(row, "amount", 1), parsedRowStatus(row), formatParsedRowIssues(row)})
	}
	return simpleTable([]string{"Fuel type", "Litres", "Status", "Issues"}, rows)
}

func parsedVehicleDistanceRowsPreview(parsed input.ParseResult) fyne.CanvasObject {
	rows := make([][]string, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		vehicleType := displayField(row, "vehicle_type", 2)
		rows = append(rows, []string{
			displayField(row, "vehicle_name", 0),
			displayField(row, "plate", 1),
			vehicleTypeLabel(vehicleType),
			vehicleSizeClassLabel(vehicleType, displayField(row, "vehicle_size_class", 3)),
			fuelTypeLabel(displayField(row, "fuel_type", 4)),
			displayField(row, "amount", 5),
			parsedRowStatus(row),
			formatParsedRowIssues(row),
		})
	}
	return simpleTable([]string{"Vehicle name", "Plate", "Vehicle type", "Size class", "Fuel type", "km", "Status", "Issues"}, rows)
}

func parsedRefrigerantRowsPreview(parsed input.ParseResult) fyne.CanvasObject {
	rows := make([][]string, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		rows = append(rows, []string{displayField(row, "substance", 0), displayField(row, "amount", 1), parsedRowStatus(row), formatParsedRowIssues(row)})
	}
	return simpleTable([]string{"Gas", "kg", "Status", "Issues"}, rows)
}

func displayField(row input.ParsedRow, normalizedKey string, rawIndex int) string {
	if value := row.Normalized[normalizedKey]; value != "" {
		return value
	}
	if rawIndex >= 0 && rawIndex < len(row.RawFields) {
		return row.RawFields[rawIndex]
	}
	return ""
}

func parsedRowStatus(row input.ParsedRow) string {
	if len(row.Errors) > 0 {
		return "Error"
	}
	if len(row.Warnings) > 0 {
		return "Warning"
	}
	return "Valid"
}

func formatParsedRowIssues(row input.ParsedRow) string {
	issues := []string{}
	if len(row.Errors) > 0 {
		issues = append(issues, formatIssues(row.Errors))
	}
	if len(row.Warnings) > 0 {
		issues = append(issues, formatIssues(row.Warnings))
	}
	return strings.Join(issues, "; ")
}

func formatIssues(issues []input.ParseIssue) string {
	if len(issues) == 0 {
		return ""
	}
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		parts = append(parts, issue.Message)
	}
	return strings.Join(parts, "; ")
}

func organizationOptions(organizations []domain.Organization) ([]string, map[string]string) {
	options := make([]string, 0, len(organizations))
	ids := map[string]string{}
	used := map[string]int{}
	for _, organization := range organizations {
		label := uniqueOptionLabel(organizationDisplayLabel(organization), used)
		options = append(options, label)
		ids[label] = organization.ID
	}
	return options, ids
}

func facilityOptions(facilities []domain.Facility) ([]string, map[string]string) {
	options := make([]string, 0, len(facilities))
	ids := map[string]string{}
	used := map[string]int{}
	for _, facility := range facilities {
		label := uniqueOptionLabel(facilityDisplayLabel(facility), used)
		options = append(options, label)
		ids[label] = facility.ID
	}
	return options, ids
}

func reportingPeriodOptions(periods []domain.ReportingPeriod) ([]string, map[string]string) {
	options := make([]string, 0, len(periods))
	ids := map[string]string{}
	used := map[string]int{}
	for _, period := range periods {
		label := uniqueOptionLabel(reportingPeriodDisplayLabel(period), used)
		options = append(options, label)
		ids[label] = period.ID
	}
	return options, ids
}

func factorSetOptions(factorSets []domain.FactorSet) ([]string, map[string]string) {
	options := make([]string, 0, len(factorSets))
	ids := map[string]string{}
	used := map[string]int{}
	for _, factorSet := range factorSets {
		label := uniqueOptionLabel(factorSetDisplayLabel(factorSet), used)
		options = append(options, label)
		ids[label] = factorSet.ID
	}
	return options, ids
}

func calculationRunOptions(runs []domain.CalculationRun) ([]string, map[string]string) {
	options := make([]string, 0, len(runs))
	ids := map[string]string{}
	used := map[string]int{}
	for _, run := range runs {
		if run.Status != domain.CalculationRunStatusCompleted {
			continue
		}
		label := uniqueOptionLabel(calculationRunDisplayLabel(run), used)
		options = append(options, label)
		ids[label] = run.ID
	}
	return options, ids
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.6g", value)
}

func formatFloatPtr(value *float64) string {
	if value == nil {
		return ""
	}
	return formatFloat(*value)
}

func formatTCO2eFromKg(value float64) string {
	return formatFloat(value / 1000)
}

func formatTCO2eFromKgPtr(value *float64) string {
	if value == nil {
		return ""
	}
	return formatTCO2eFromKg(*value)
}
