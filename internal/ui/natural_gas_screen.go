package ui

import (
	"fyne.io/fyne/v2"

	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/vocab"
)

func (a *App) naturalGasScreen() fyne.CanvasObject {
	if a.state.ReportingPeriod == nil || len(a.state.Facilities) == 0 {
		return prerequisiteScreen("Natural gas", "Select an organization, create a facility, and create a reporting period first.")
	}
	facilitySelect, facilityIDs := a.facilitySelector()
	return a.pasteInputScreen(pasteScreenConfig{
		Title:       "Natural gas",
		Unit:        "Smc",
		Expected:    "month | consumption",
		Example:     "January\t100\nFebruary\t120",
		InputKind:   vocab.InputNaturalGasMonthlySmc,
		Facility:    facilitySelect,
		FacilityIDs: facilityIDs,
		CommitContext: func(facilityID string) input.CommitContext {
			return a.commitContext(vocab.InputNaturalGasMonthlySmc, &facilityID, false, domain.GOCoverageNone)
		},
		SavedData: a.savedNaturalGasData,
	})
}
