package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
)

func (a *App) organizationScreen() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("Organization name")

	options, ids := organizationOptions(a.state.Organizations)
	selectOrganization := widget.NewSelect(options, nil)
	if a.state.Organization != nil {
		for label, id := range ids {
			if id == a.state.Organization.ID {
				selectOrganization.SetSelected(label)
				break
			}
		}
	}
	selectOrganization.OnChanged = func(label string) {
		a.state.SetOrganization(ids[label])
		a.refreshAndRender()
	}

	rows := make([][]string, 0, len(a.state.Organizations))
	for _, organization := range a.state.Organizations {
		rows = append(rows, []string{organizationDisplayLabel(organization)})
	}

	save := widget.NewButton("Save", func() {
		if cleanText(name.Text) == "" {
			a.showError(required("organization name"))
			return
		}
		id, err := newID("organization")
		if err != nil {
			a.showError(err)
			return
		}
		now := time.Now().UTC()
		organization := domain.Organization{
			ID:        id,
			Name:      cleanText(name.Text),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := a.state.Store.CreateOrganization(organization); err != nil {
			a.showError(err)
			return
		}
		a.state.Organization = &organization
		a.showInfo("Organization saved", "Organization created.")
		a.refreshAndRender()
	})

	return screen("Organization",
		widget.NewForm(
			widget.NewFormItem("Existing organizations", selectOrganization),
			widget.NewFormItem("Organization name", name),
		),
		actionRow(save, widget.NewButton("Refresh", a.refreshAndRender)),
		widget.NewLabel("Organizations"),
		simpleTable([]string{"Name"}, rows),
	)
}

func currentContextBar(a *App) fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Organization: "+currentOrganizationLabel(a.state.Organization)),
		widget.NewLabel("Reporting period: "+currentReportingPeriodLabel(a.state.ReportingPeriod)),
	)
}
