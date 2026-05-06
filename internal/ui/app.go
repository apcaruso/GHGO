package ui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"ghgo/internal/store"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window
	state   *State
	content *fyne.Container
	screen  screenID
}

func Run(dbPath string, st *store.Store) error {
	fyneApp := app.New()
	window := fyneApp.NewWindow("ghgo")
	window.Resize(fyne.NewSize(1100, 750))

	ui := &App{
		fyneApp: fyneApp,
		window:  window,
		state:   NewState(st, dbPath),
		content: container.NewStack(),
		screen:  screenOrganization,
	}
	if err := ui.refreshState(); err != nil {
		return err
	}
	if ui.state.Organization != nil {
		ui.screen = screenFacilities
	}
	ui.render()
	window.ShowAndRun()
	return nil
}

func (a *App) refreshState() error {
	return a.state.Refresh(context.Background())
}

func (a *App) render() {
	a.content.Objects = []fyne.CanvasObject{a.buildScreen(a.screen)}
	a.content.Refresh()
	a.window.SetContent(container.NewBorder(nil, nil, a.buildNavigation(), nil, a.content))
}

func (a *App) setScreen(screen screenID) {
	a.screen = screen
	a.render()
}

func (a *App) refreshAndRender() {
	if err := a.refreshState(); err != nil {
		a.showError(err)
		return
	}
	a.render()
}

func (a *App) showError(err error) {
	if err == nil {
		return
	}
	dialog.ShowError(err, a.window)
}

func (a *App) showInfo(title, message string) {
	dialog.ShowInformation(title, message, a.window)
}

func (a *App) buildScreen(screen screenID) fyne.CanvasObject {
	switch screen {
	case screenOrganization:
		return a.organizationScreen()
	case screenFacilities:
		return a.facilitiesScreen()
	case screenReportingPeriod:
		return a.reportingPeriodScreen()
	case screenElectricity:
		return a.electricityScreen()
	case screenNaturalGas:
		return a.naturalGasScreen()
	case screenMobile:
		return a.mobileScreen()
	case screenRefrigerants:
		return a.refrigerantsScreen()
	case screenCalculations:
		return a.calculationsScreen()
	case screenReports:
		return a.reportsScreen()
	case screenSettings:
		return a.settingsScreen()
	}
	return messageScreen("Unknown screen")
}
