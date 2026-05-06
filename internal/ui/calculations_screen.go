package ui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/calc"
	"ghgo/internal/domain"
	"ghgo/internal/factors"
)

func (a *App) calculationsScreen() fyne.CanvasObject {
	if a.state.Organization == nil || a.state.ReportingPeriod == nil {
		return prerequisiteScreen("Calculations", "Select an organization and reporting period first.")
	}
	factorSet := a.state.DefaultFactorSet()
	if factorSet == nil {
		return prerequisiteScreen("Calculations", "Emission factors are not available. Restart the app to run database initialization.")
	}

	runSummary := widget.NewLabel("")
	runsTable := container.NewStack(simpleTable([]string{"Created", "Status", "Primary total kgCO2e", "Primary total tCO2e", "Location-based total kgCO2e", "Location-based total tCO2e"}, a.calculationRunRows()))
	refreshRuns := func() {
		if err := a.refreshState(); err != nil {
			a.showError(err)
			return
		}
		runsTable.Objects = []fyne.CanvasObject{simpleTable([]string{"Created", "Status", "Primary total kgCO2e", "Primary total tCO2e", "Location-based total kgCO2e", "Location-based total tCO2e"}, a.calculationRunRows())}
		runsTable.Refresh()
	}

	runButton := widget.NewButton("Run calculation", func() {
		active, err := a.state.Store.ListActiveActivityRecordsByPeriod(a.state.ReportingPeriod.ID)
		if err != nil {
			a.showError(err)
			return
		}
		if len(active) == 0 {
			a.showError(prerequisite("at least one active activity record is required"))
			return
		}
		lookup := factors.NewLookup(a.state.Store, factorSet.ID)
		engine := calc.NewEngine(a.state.Store, lookup)
		result, err := engine.Run(context.Background(), calc.RunOptions{
			OrganizationID:    a.state.Organization.ID,
			ReportingPeriodID: a.state.ReportingPeriod.ID,
			FactorSetID:       factorSet.ID,
		})
		if err != nil {
			a.showError(err)
			return
		}
		runSummary.SetText(fmt.Sprintf("Calculation completed\nResults created: %d\nPrimary total kgCO2e: %s\nPrimary total tCO2e: %s\nLocation-based total kgCO2e: %s\nLocation-based total tCO2e: %s",
			result.ResultsCreated,
			formatFloat(result.PrimaryTotalKgCO2e),
			formatTCO2eFromKg(result.PrimaryTotalKgCO2e),
			formatFloatPtr(result.LocationBasedTotalKgCO2e),
			formatTCO2eFromKgPtr(result.LocationBasedTotalKgCO2e),
		))
		a.showInfo("Calculation completed", "Calculation run completed.")
		refreshRuns()
	})

	return screen("Calculations",
		currentContextBar(a),
		widget.NewLabel("Emission factor set: "+factorSetDisplayLabel(*factorSet)),
		actionRow(runButton, widget.NewButton("Refresh", refreshRuns)),
		runSummary,
		widget.NewLabel("Calculation runs"),
		runsTable,
	)
}

func (a *App) calculationRunRows() [][]string {
	rows := make([][]string, 0, len(a.state.CalculationRuns))
	for i := len(a.state.CalculationRuns) - 1; i >= 0; i-- {
		run := a.state.CalculationRuns[i]
		primary, location, err := a.calculationRunTotals(run)
		primaryText := ""
		primaryTonnesText := ""
		locationText := ""
		locationTonnesText := ""
		if err == nil && run.Status == domain.CalculationRunStatusCompleted {
			primaryText = formatFloat(primary)
			primaryTonnesText = formatTCO2eFromKg(primary)
			locationText = formatFloatPtr(location)
			locationTonnesText = formatTCO2eFromKgPtr(location)
		}
		rows = append(rows, []string{
			run.StartedAt.Local().Format("2006-01-02 15:04"),
			calculationRunStatusLabel(run.Status),
			primaryText,
			primaryTonnesText,
			locationText,
			locationTonnesText,
		})
	}
	return rows
}

func (a *App) calculationRunTotals(run domain.CalculationRun) (float64, *float64, error) {
	results, err := a.state.Store.ListCalculationResultsByRun(run.ID)
	if err != nil {
		return 0, nil, err
	}
	primaryTotal := 0.0
	locationTotal := 0.0
	hasLocationComparison := false
	for _, result := range results {
		if result.IsPrimary {
			primaryTotal += result.EmissionsKgCO2e
		}
		if result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodLocationBased && !result.IsPrimary {
			hasLocationComparison = true
		}
	}
	if !hasLocationComparison {
		return primaryTotal, nil, nil
	}
	for _, result := range results {
		switch {
		case result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodLocationBased:
			locationTotal += result.EmissionsKgCO2e
		case result.Vector != domain.ActivityVectorElectricity && result.IsPrimary:
			locationTotal += result.EmissionsKgCO2e
		}
	}
	return primaryTotal, &locationTotal, nil
}

func completedRunIDs(runs []domain.CalculationRun) []string {
	ids := []string{}
	for _, run := range runs {
		if run.Status == domain.CalculationRunStatusCompleted {
			ids = append(ids, run.ID)
		}
	}
	return ids
}
