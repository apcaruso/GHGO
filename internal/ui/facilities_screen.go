package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
)

func (a *App) facilitiesScreen() fyne.CanvasObject {
	if a.state.Organization == nil {
		return prerequisiteScreen("Facilities", "Create or select an organization first.")
	}

	name := widget.NewEntry()
	country := widget.NewEntry()
	country.SetText("IT")

	rows := make([][]string, 0, len(a.state.Facilities))
	for _, facility := range a.state.Facilities {
		rows = append(rows, []string{facility.Name, facility.CountryCode})
	}

	add := widget.NewButton("Save", func() {
		if cleanText(name.Text) == "" {
			a.showError(required("facility name"))
			return
		}
		countryCode, err := validateCountryCode(country.Text)
		if err != nil {
			a.showError(err)
			return
		}
		id, err := newID("facility")
		if err != nil {
			a.showError(err)
			return
		}
		now := time.Now().UTC()
		if err := a.state.Store.CreateFacility(domain.Facility{
			ID:             id,
			OrganizationID: a.state.Organization.ID,
			Name:           cleanText(name.Text),
			CountryCode:    countryCode,
			CreatedAt:      now,
			UpdatedAt:      now,
		}); err != nil {
			a.showError(err)
			return
		}
		a.showInfo("Facility added", "Facility created.")
		a.refreshAndRender()
	})

	return screen("Facilities",
		currentContextBar(a),
		widget.NewForm(
			widget.NewFormItem("Facility name", name),
			widget.NewFormItem("Country code", country),
		),
		actionRow(add, widget.NewButton("Refresh", a.refreshAndRender)),
		widget.NewLabel("Facilities"),
		simpleTable([]string{"Name", "Country"}, rows),
	)
}
