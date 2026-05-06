package ui

import (
	"fyne.io/fyne/v2"
)

func (a *App) settingsScreen() fyne.CanvasObject {
	organization := ""
	if a.state.Organization != nil {
		organization = organizationDisplayLabel(*a.state.Organization)
	}
	period := ""
	if a.state.ReportingPeriod != nil {
		period = reportingPeriodDisplayLabel(*a.state.ReportingPeriod)
	}
	factorSet := ""
	if a.state.FactorSet != nil {
		factorSet = factorSetDisplayLabel(*a.state.FactorSet)
	}
	return screen("Settings",
		simpleTable([]string{"Setting", "Value"}, [][]string{
			{"Database path", a.state.DBPath},
			{"Current organization", organization},
			{"Current reporting period", period},
			{"Selected factor set", factorSet},
			{"Project", "ghgo"},
			{"Version", "development"},
		}),
	)
}
