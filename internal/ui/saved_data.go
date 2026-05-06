package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

func (a *App) savedElectricityData(facilityID string) (fyne.CanvasObject, error) {
	records, err := a.listSavedRecords(facilityID, domain.ActivitySourceKindElectricityMonthlyKWh)
	if err != nil {
		return nil, err
	}
	settings, err := a.state.Store.GetElectricitySettings(a.state.ReportingPeriod.ID, domain.ID(facilityID))
	if errors.Is(err, store.ErrNotFound) {
		settings = nil
	} else if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return emptySavedData(), nil
	}

	coverage := "No GO"
	if settings != nil && settings.HasGuaranteesOfOrigin && settings.GOCoverage == domain.GOCoverageFull {
		coverage = "Full GO"
	}
	rows := make([][]string, 0, len(records))
	total := 0.0
	for _, record := range records {
		total += record.Amount
		rows = append(rows, []string{monthNameFromRecord(record), formatFloat(record.Amount), coverage})
	}

	return container.NewVBox(
		widget.NewLabel("Saved data"),
		simpleTable([]string{"Month", "Consumption kWh", "GO coverage"}, rows),
		simpleTable([]string{"Metric", "Value"}, [][]string{
			{"Total kWh", formatFloat(total)},
			{"Missing months", missingMonthsLabel(records)},
			{"Last saved", latestSavedLabel(records)},
		}),
	), nil
}

func (a *App) savedNaturalGasData(facilityID string) (fyne.CanvasObject, error) {
	records, err := a.listSavedRecords(facilityID, domain.ActivitySourceKindNaturalGasMonthlySMC)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return emptySavedData(), nil
	}

	rows := make([][]string, 0, len(records))
	total := 0.0
	for _, record := range records {
		total += record.Amount
		rows = append(rows, []string{monthNameFromRecord(record), formatFloat(record.Amount)})
	}

	return container.NewVBox(
		widget.NewLabel("Saved data"),
		simpleTable([]string{"Month", "Consumption Smc"}, rows),
		simpleTable([]string{"Metric", "Value"}, [][]string{
			{"Total Smc", formatFloat(total)},
			{"Missing months", missingMonthsLabel(records)},
		}),
	), nil
}

func (a *App) savedMobileFuelData() (fyne.CanvasObject, error) {
	records, err := a.listSavedRecords("", domain.ActivitySourceKindMobileFuelLitres)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return emptySavedData(), nil
	}

	rows := make([][]string, 0, len(records))
	total := 0.0
	for _, record := range records {
		total += record.Amount
		rows = append(rows, []string{fuelTypeLabel(record.FuelType), formatFloat(record.Amount)})
	}

	return container.NewVBox(
		widget.NewLabel("Saved data"),
		simpleTable([]string{"Fuel type", "Litres"}, rows),
		simpleTable([]string{"Metric", "Value"}, [][]string{{"Total litres", formatFloat(total)}}),
	), nil
}

func (a *App) savedMobileDistanceData() (fyne.CanvasObject, error) {
	records, err := a.listSavedRecords("", domain.ActivitySourceKindVehicleDistanceKM)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return emptySavedData(), nil
	}

	rows := make([][]string, 0, len(records))
	total := 0.0
	for _, record := range records {
		total += record.Amount
		rows = append(rows, []string{
			record.VehicleName,
			record.Plate,
			vehicleTypeLabel(record.VehicleType),
			vehicleSizeClassLabel(record.VehicleType, record.VehicleSizeClass),
			fuelTypeLabel(record.FuelType),
			formatFloat(record.Amount),
		})
	}

	return container.NewVBox(
		widget.NewLabel("Saved data"),
		simpleTable([]string{"Vehicle name", "Plate", "Vehicle type", "Size class", "Fuel type", "km"}, rows),
		simpleTable([]string{"Metric", "Value"}, [][]string{
			{"Total km", formatFloat(total)},
			{"Vehicles", formatInt(len(records))},
		}),
	), nil
}

func (a *App) savedRefrigerantsData(facilityID string) (fyne.CanvasObject, error) {
	records, err := a.listSavedRecords(facilityID, domain.ActivitySourceKindRefrigerantsAnnualKG)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return emptySavedData(), nil
	}

	rows := make([][]string, 0, len(records))
	total := 0.0
	for _, record := range records {
		total += record.Amount
		rows = append(rows, []string{record.Substance, formatFloat(record.Amount)})
	}

	return container.NewVBox(
		widget.NewLabel("Saved data"),
		simpleTable([]string{"Gas", "kg"}, rows),
		simpleTable([]string{"Metric", "Value"}, [][]string{{"Total kg", formatFloat(total)}}),
	), nil
}

func (a *App) listSavedRecords(facilityID string, sourceKind domain.ActivitySourceKind) ([]domain.ActivityRecord, error) {
	if a.state.ReportingPeriod == nil {
		return nil, required("reporting period")
	}
	var facilityIDPtr *domain.ID
	if strings.TrimSpace(facilityID) != "" {
		id := domain.ID(facilityID)
		facilityIDPtr = &id
	}
	return a.state.Store.ListActiveActivityRecordsByPeriodFacilitySource(a.state.ReportingPeriod.ID, facilityIDPtr, sourceKind)
}

func emptySavedData() fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel("Saved data"), widget.NewLabel("No saved data yet."))
}

func monthNameFromRecord(record domain.ActivityRecord) string {
	return vocab.Month(record.PeriodStart.Month()).EnglishName()
}

func missingMonthsLabel(records []domain.ActivityRecord) string {
	seen := map[int]bool{}
	for _, record := range records {
		seen[int(record.PeriodStart.Month())] = true
	}
	missing := []string{}
	for month := 1; month <= 12; month++ {
		if !seen[month] {
			missing = append(missing, vocab.Month(month).EnglishName())
		}
	}
	if len(missing) == 0 {
		return "None"
	}
	return strings.Join(missing, ", ")
}

func latestSavedLabel(records []domain.ActivityRecord) string {
	if len(records) == 0 {
		return ""
	}
	latest := records[0].UpdatedAt
	for _, record := range records[1:] {
		if record.UpdatedAt.After(latest) {
			latest = record.UpdatedAt
		}
	}
	return latest.Local().Format("2006-01-02 15:04")
}

func formatDate(value time.Time) string {
	return value.Format("2006-01-02")
}

func formatKgCO2e(value float64) string {
	return fmt.Sprintf("%s kgCO2e", formatFloat(value))
}
