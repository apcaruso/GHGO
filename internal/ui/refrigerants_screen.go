package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/vocab"
)

func (a *App) refrigerantsScreen() fyne.CanvasObject {
	if a.state.ReportingPeriod == nil || len(a.state.Facilities) == 0 {
		return prerequisiteScreen("Refrigerants", "Select an organization, create a facility, and create a reporting period first.")
	}
	facilitySelect, facilityIDs := a.facilitySelector()
	return a.pasteInputScreen(pasteScreenConfig{
		Title:        "Refrigerants",
		Unit:         "kg",
		Expected:     "gas | kg",
		Example:      "R410A\t1.5\nR134a\t0.5",
		InputKind:    vocab.InputRefrigerantsAnnualKg,
		Facility:     facilitySelect,
		FacilityIDs:  facilityIDs,
		ExtraContent: widget.NewLabel("Supported gases: R134a, R410A, R407C, R404A, R32, R22. Unit: kg."),
		CommitContext: func(facilityID string) input.CommitContext {
			return a.commitContext(vocab.InputRefrigerantsAnnualKg, &facilityID, false, domain.GOCoverageNone)
		},
		SavedData: a.savedRefrigerantsData,
	})
}
