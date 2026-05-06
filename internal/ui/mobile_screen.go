package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/vocab"
)

func (a *App) mobileScreen() fyne.CanvasObject {
	if a.state.ReportingPeriod == nil {
		return prerequisiteScreen("Mobile combustion", "Create or select a reporting period first.")
	}
	if a.state.Settings == nil || a.state.Settings.MobileMethod == "" {
		return prerequisiteScreen("Mobile combustion", "Select a mobile combustion method in Reporting period first.")
	}

	switch a.state.Settings.MobileMethod {
	case domain.MobileMethodFuelBased:
		return a.mobileFuelScreen()
	case domain.MobileMethodDistanceBased:
		return a.mobileDistanceScreen()
	}
	return prerequisiteScreen("Mobile combustion", "The selected mobile combustion method is invalid.")
}

func (a *App) mobileFuelScreen() fyne.CanvasObject {
	kind := vocab.InputMobileFuelLitres
	return a.pasteInputScreen(pasteScreenConfig{
		Title:     "Mobile combustion - Fuel consumed",
		Unit:      "litres",
		Expected:  "fuel type | litres",
		Example:   "Diesel\t200\nPetrol\t50",
		InputKind: kind,
		ExtraContent: container.NewVBox(
			widget.NewLabel("Allowed fuel types"),
			widget.NewLabel("Diesel, Petrol, LPG, CNG"),
		),
		CommitContext: func(_ string) input.CommitContext {
			return a.commitContext(kind, nil, false, domain.GOCoverageNone)
		},
		SavedData: func(_ string) (fyne.CanvasObject, error) { return a.savedMobileFuelData() },
	})
}

func (a *App) mobileDistanceScreen() fyne.CanvasObject {
	kind := vocab.InputVehicleDistanceKm
	return a.pasteInputScreen(pasteScreenConfig{
		Title:     "Mobile combustion - Distance travelled",
		Unit:      "km",
		Expected:  "vehicle name | plate | vehicle type | size class | fuel type | km",
		Example:   "Fiat Panda\tAB123CD\tCar\tSmall\tPetrol\t1000\nFiat Doblo\tCD456EF\tVan\tClass II\tDiesel\t2000",
		InputKind: kind,
		ExtraContent: container.NewVBox(
			widget.NewLabel("Allowed vehicle types: Car, Van, Motorbike"),
			widget.NewLabel("Car and Motorbike size classes: Small, Medium, Large, Average"),
			widget.NewLabel("Van size classes: Class I, Class II, Class III, Average"),
			widget.NewLabel("Allowed fuel types: Diesel, Petrol, LPG, CNG, Hybrid, Plug-in Hybrid Electric Vehicle, Battery Electric Vehicle, Unknown"),
		),
		CommitContext: func(_ string) input.CommitContext {
			return a.commitContext(kind, nil, false, domain.GOCoverageNone)
		},
		SavedData: func(_ string) (fyne.CanvasObject, error) { return a.savedMobileDistanceData() },
	})
}
