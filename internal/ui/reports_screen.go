package ui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
	"ghgo/internal/report"
)

func (a *App) reportsScreen() fyne.CanvasObject {
	if a.state.ReportingPeriod == nil {
		return prerequisiteScreen("Reports", "Select a reporting period first.")
	}
	options, ids := calculationRunOptions(a.state.CalculationRuns)
	if len(options) == 0 {
		return prerequisiteScreen("Reports", "No completed calculation run found. Run a calculation first.")
	}

	runSelect := widget.NewSelect(options, nil)
	latest := a.state.LatestCompletedRun()
	if latest != nil {
		for label, id := range ids {
			if id == latest.ID {
				runSelect.SetSelected(label)
				break
			}
		}
	}
	if runSelect.Selected == "" {
		runSelect.SetSelected(options[len(options)-1])
	}

	output := container.NewStack(widget.NewLabel("Select a completed calculation run, then load the report."))
	build := widget.NewButton("Load latest report", func() {
		runID := ids[runSelect.Selected]
		if runID == "" {
			a.showError(required("calculation run"))
			return
		}
		tables, err := report.NewBuilder(a.state.Store).BuildTables(context.Background(), report.BuildOptions{CalculationRunID: runID})
		if err != nil {
			a.showError(err)
			return
		}
		output.Objects = []fyne.CanvasObject{a.renderReportTables(tables)}
		output.Refresh()
	})

	return screen("Reports",
		currentContextBar(a),
		widget.NewForm(widget.NewFormItem("Calculation run", runSelect)),
		actionRow(build, widget.NewButton("Refresh", a.refreshAndRender)),
		output,
	)
}

func (a *App) renderReportTables(tables *report.ReportTables) fyne.CanvasObject {
	objects := []fyne.CanvasObject{
		widget.NewLabel("Executive summary"),
		simpleTable([]string{"Metric", "Value"}, [][]string{
			{"Primary total kgCO2e", formatFloat(tables.ExecutiveSummary.PrimaryTotalKgCO2e)},
			{"Primary total tCO2e", formatFloat(tables.ExecutiveSummary.PrimaryTotalTCO2e)},
			{"Location-based total kgCO2e", formatFloatPtr(tables.ExecutiveSummary.LocationBasedTotalKgCO2e)},
			{"Location-based total tCO2e", formatFloatPtr(tables.ExecutiveSummary.LocationBasedTotalTCO2e)},
			{"Electricity primary method", methodLabel(tables.ExecutiveSummary.ElectricityPrimaryMethod)},
			{"Scope 1 primary kgCO2e", formatFloat(tables.ExecutiveSummary.Scope1PrimaryKgCO2e)},
			{"Scope 1 primary tCO2e", formatTCO2eFromKg(tables.ExecutiveSummary.Scope1PrimaryKgCO2e)},
			{"Scope 2 primary kgCO2e", formatFloat(tables.ExecutiveSummary.Scope2PrimaryKgCO2e)},
			{"Scope 2 primary tCO2e", formatTCO2eFromKg(tables.ExecutiveSummary.Scope2PrimaryKgCO2e)},
		}),
		widget.NewLabel("Scope summary"),
		scopeSummaryTable(tables.ScopeSummary),
		widget.NewLabel("Vector summary"),
		vectorSummaryTable(tables.VectorSummary),
		widget.NewLabel("Monthly emissions"),
		monthlyTable(tables.MonthlyEmissions),
		widget.NewLabel("Electricity detail"),
		a.electricityDetailTable(tables.ElectricityDetail),
		widget.NewLabel("Natural gas detail"),
		a.naturalGasDetailTable(tables.NaturalGasDetail),
		widget.NewLabel("Mobile combustion detail"),
		mobileDetailTable(tables.MobileDetail),
		widget.NewLabel("Refrigerants detail"),
		a.refrigerantsDetailTable(tables.RefrigerantsDetail),
		widget.NewLabel("Chart datasets"),
		chartDatasetTables(tables),
	}
	return container.NewVBox(objects...)
}

func scopeSummaryTable(table report.ScopeSummaryTable) fyne.CanvasObject {
	rows := [][]string{}
	for _, row := range table.Rows {
		rows = append(rows, []string{scopeLabel(row.Scope), formatFloat(row.PrimaryKgCO2e), formatFloat(row.PrimaryTCO2e), formatFloat(row.PrimaryShare), formatFloatPtr(row.LocationBasedKgCO2e), formatFloatPtr(row.LocationBasedTCO2e), formatFloatPtr(row.LocationBasedShare)})
	}
	return simpleTable([]string{"Scope", "Primary kgCO2e", "Primary tCO2e", "Primary share", "Location kgCO2e", "Location tCO2e", "Location share"}, rows)
}

func vectorSummaryTable(table report.VectorSummaryTable) fyne.CanvasObject {
	rows := [][]string{}
	for _, row := range table.Rows {
		rows = append(rows, []string{vectorLabel(row.Vector), row.ActivitySummary, formatFloat(row.PrimaryKgCO2e), formatFloat(row.PrimaryTCO2e), formatFloat(row.PrimaryShare), formatFloatPtr(row.LocationBasedKgCO2e), formatFloatPtr(row.LocationBasedTCO2e)})
	}
	return simpleTable([]string{"Vector", "Activity", "Primary kgCO2e", "Primary tCO2e", "Primary share", "Location kgCO2e", "Location tCO2e"}, rows)
}

func monthlyTable(table report.MonthlyEmissionsTable) fyne.CanvasObject {
	rows := [][]string{}
	for _, row := range table.Rows {
		rows = append(rows, []string{formatInt(row.Month), row.MonthName, formatFloat(row.ElectricityConsumptionKWh), formatFloat(row.ElectricityPrimaryKgCO2e), formatTCO2eFromKg(row.ElectricityPrimaryKgCO2e), formatFloatPtr(row.ElectricityLocationBasedKgCO2e), formatTCO2eFromKgPtr(row.ElectricityLocationBasedKgCO2e), formatFloat(row.NaturalGasConsumptionSmc), formatFloat(row.NaturalGasKgCO2e), formatTCO2eFromKg(row.NaturalGasKgCO2e), formatFloat(row.MonthlyPrimaryKgCO2e), formatTCO2eFromKg(row.MonthlyPrimaryKgCO2e), formatFloatPtr(row.MonthlyLocationBasedKgCO2e), formatTCO2eFromKgPtr(row.MonthlyLocationBasedKgCO2e)})
	}
	return simpleTable([]string{"Month", "Name", "kWh", "Elec primary kgCO2e", "Elec primary tCO2e", "Elec location kgCO2e", "Elec location tCO2e", "Smc", "Gas kgCO2e", "Gas tCO2e", "Monthly primary kgCO2e", "Monthly primary tCO2e", "Monthly location kgCO2e", "Monthly location tCO2e"}, rows)
}

func (a *App) electricityDetailTable(table report.ElectricityDetailTable) fyne.CanvasObject {
	rows := [][]string{}
	for _, row := range table.Rows {
		rows = append(rows, []string{a.facilityLabelByID(row.FacilityID), row.MonthName, formatFloat(row.ConsumptionKWh), methodLabel(row.PrimaryMethod), formatFloatPtr(row.LocationBasedKgCO2e), formatTCO2eFromKgPtr(row.LocationBasedKgCO2e), formatFloatPtr(row.MarketBasedKgCO2e), formatTCO2eFromKgPtr(row.MarketBasedKgCO2e), row.MarketBasedSource})
	}
	return simpleTable([]string{"Facility", "Month", "kWh", "Primary method", "Location kgCO2e", "Location tCO2e", "Market kgCO2e", "Market tCO2e", "Market source"}, rows)
}

func (a *App) naturalGasDetailTable(table report.NaturalGasDetailTable) fyne.CanvasObject {
	rows := [][]string{}
	for _, row := range table.Rows {
		rows = append(rows, []string{a.facilityLabelByID(row.FacilityID), row.MonthName, formatFloat(row.ConsumptionSmc), formatFloat(row.FactorValue), row.FactorUnit, formatFloat(row.EmissionsKgCO2e), formatTCO2eFromKg(row.EmissionsKgCO2e)})
	}
	return simpleTable([]string{"Facility", "Month", "Smc", "Factor", "Unit", "kgCO2e", "tCO2e"}, rows)
}

func mobileDetailTable(table report.MobileDetailTable) fyne.CanvasObject {
	objects := []fyne.CanvasObject{simpleTable([]string{"Metric", "Value"}, [][]string{{"Method", methodLabel(table.Method)}, {"Total kgCO2e", formatFloat(table.TotalKgCO2e)}, {"Total tCO2e", formatTCO2eFromKg(table.TotalKgCO2e)}})}
	fuelRows := [][]string{}
	for _, row := range table.FuelRows {
		fuelRows = append(fuelRows, []string{fuelTypeLabel(row.FuelType), formatFloat(row.Litres), formatFloat(row.FactorValue), row.FactorUnit, formatFloat(row.EmissionsKgCO2e), formatTCO2eFromKg(row.EmissionsKgCO2e)})
	}
	if len(fuelRows) > 0 {
		objects = append(objects, simpleTable([]string{"Fuel type", "Litres", "Factor", "Unit", "kgCO2e", "tCO2e"}, fuelRows))
	}
	distanceRows := [][]string{}
	for _, row := range table.DistanceRows {
		distanceRows = append(distanceRows, []string{row.VehicleName, row.Plate, vehicleTypeLabel(row.VehicleType), vehicleSizeClassLabel(row.VehicleType, row.VehicleSizeClass), fuelTypeLabel(row.FuelType), formatFloat(row.Km), formatFloat(row.FactorValue), row.FactorUnit, formatFloat(row.EmissionsKgCO2e), formatTCO2eFromKg(row.EmissionsKgCO2e)})
	}
	if len(distanceRows) > 0 {
		objects = append(objects, simpleTable([]string{"Vehicle name", "Plate", "Vehicle type", "Size class", "Fuel type", "km", "Factor", "Unit", "kgCO2e", "tCO2e"}, distanceRows))
	}
	return container.NewVBox(objects...)
}

func (a *App) refrigerantsDetailTable(table report.RefrigerantsDetailTable) fyne.CanvasObject {
	rows := [][]string{}
	for _, row := range table.Rows {
		rows = append(rows, []string{a.facilityLabelByID(row.FacilityID), row.Substance, formatFloat(row.QuantityKg), formatFloat(row.FactorValue), row.FactorUnit, formatFloat(row.EmissionsKgCO2e), formatTCO2eFromKg(row.EmissionsKgCO2e)})
	}
	return simpleTable([]string{"Facility", "Substance", "kg", "Factor", "Unit", "kgCO2e", "tCO2e"}, rows)
}

func chartDatasetTables(tables *report.ReportTables) fyne.CanvasObject {
	vectorRows := [][]string{}
	for _, row := range tables.VectorSummary.Rows {
		if row.Vector == "Total" {
			continue
		}
		vectorRows = append(vectorRows, []string{vectorLabel(row.Vector), formatFloat(row.PrimaryKgCO2e), formatFloat(row.PrimaryTCO2e)})
	}
	monthlyRows := [][]string{}
	for _, row := range tables.MonthlyEmissions.Rows {
		monthlyRows = append(monthlyRows, []string{row.MonthName, formatFloat(row.MonthlyPrimaryKgCO2e), formatTCO2eFromKg(row.MonthlyPrimaryKgCO2e)})
	}
	return container.NewVBox(
		widget.NewLabel("Primary emissions by vector"),
		simpleTable([]string{"Label", "Primary kgCO2e", "Primary tCO2e"}, vectorRows),
		widget.NewLabel("Monthly primary emissions"),
		simpleTable([]string{"Month", "Primary kgCO2e", "Primary tCO2e"}, monthlyRows),
	)
}

func (a *App) facilityLabelByID(id string) string {
	if id == "" {
		return ""
	}
	for _, facility := range a.state.Facilities {
		if facility.ID == id {
			return facilityDisplayLabel(facility)
		}
	}
	return "Facility"
}

func scopeLabel(value string) string {
	switch value {
	case "Scope 1", "Scope 2", "Total":
		return value
	case "1":
		return "Scope 1"
	case "2":
		return "Scope 2"
	}
	return value
}

func vectorLabel(value string) string {
	switch value {
	case string(domain.ActivityVectorElectricity):
		return "Electricity"
	case string(domain.ActivityVectorNaturalGas):
		return "Natural gas"
	case string(domain.ActivityVectorMobileCombustion):
		return "Mobile combustion"
	case string(domain.ActivityVectorRefrigerants):
		return "Refrigerants"
	case "Total":
		return "Total"
	}
	return value
}

func methodLabel(value string) string {
	return activityMethodLabel(domain.ActivityMethod(value))
}
