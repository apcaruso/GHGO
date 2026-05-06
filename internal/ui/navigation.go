package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type screenID string

const (
	screenOrganization    screenID = "organization"
	screenFacilities      screenID = "facilities"
	screenReportingPeriod screenID = "reporting_period"
	screenElectricity     screenID = "electricity"
	screenNaturalGas      screenID = "natural_gas"
	screenMobile          screenID = "mobile"
	screenRefrigerants    screenID = "refrigerants"
	screenCalculations    screenID = "calculations"
	screenReports         screenID = "reports"
	screenSettings        screenID = "settings"
)

type navigationItem struct {
	screen screenID
	label  string
}

var navigationItems = []navigationItem{
	{screenOrganization, "Organization"},
	{screenFacilities, "Facilities"},
	{screenReportingPeriod, "Reporting period"},
	{screenElectricity, "Electricity"},
	{screenNaturalGas, "Natural gas"},
	{screenMobile, "Mobile combustion"},
	{screenRefrigerants, "Refrigerants"},
	{screenCalculations, "Calculations"},
	{screenReports, "Reports"},
	{screenSettings, "Settings"},
}

func (a *App) buildNavigation() fyne.CanvasObject {
	buttons := []fyne.CanvasObject{widget.NewLabel("ghgo")}
	for _, item := range navigationItems {
		item := item
		button := widget.NewButton(item.label, func() {
			a.setScreen(item.screen)
		})
		if item.screen == a.screen {
			button.Disable()
		} else if !a.navigationAvailable(item.screen) {
			button.Disable()
		}
		buttons = append(buttons, button)
	}
	return container.NewVBox(buttons...)
}

func (a *App) navigationAvailable(screen screenID) bool {
	switch screen {
	case screenOrganization, screenSettings:
		return true
	case screenFacilities, screenReportingPeriod:
		return a.state.Organization != nil
	case screenElectricity, screenNaturalGas, screenRefrigerants:
		return a.state.Organization != nil && a.state.ReportingPeriod != nil && len(a.state.Facilities) > 0
	case screenMobile:
		return a.state.Organization != nil && a.state.ReportingPeriod != nil
	case screenCalculations:
		return a.state.Organization != nil && a.state.ReportingPeriod != nil
	case screenReports:
		return a.state.Organization != nil && a.state.ReportingPeriod != nil
	}
	return false
}
