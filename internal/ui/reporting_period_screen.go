package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
)

func (a *App) reportingPeriodScreen() fyne.CanvasObject {
	if a.state.Organization == nil {
		return prerequisiteScreen("Reporting period", "Create or select an organization first.")
	}

	currentYear := time.Now().Year()
	year := widget.NewEntry()
	year.SetText(fmt.Sprintf("%d", currentYear))
	start := widget.NewEntry()
	start.SetText(fmt.Sprintf("%d-01-01", currentYear))
	end := widget.NewEntry()
	end.SetText(fmt.Sprintf("%d-12-31", currentYear))

	method := widget.NewSelect([]string{"Fuel consumed", "Distance travelled"}, nil)
	if a.state.Settings != nil {
		method.SetSelected(mobileMethodLabel(a.state.Settings.MobileMethod))
	}

	options, ids := reportingPeriodOptions(a.state.ReportingPeriods)
	selectPeriod := widget.NewSelect(options, nil)
	if a.state.ReportingPeriod != nil {
		for label, id := range ids {
			if id == a.state.ReportingPeriod.ID {
				selectPeriod.SetSelected(label)
				break
			}
		}
	}
	selectPeriod.OnChanged = func(label string) {
		a.state.SetReportingPeriod(ids[label])
		a.refreshAndRender()
	}

	create := widget.NewButton("Create reporting period", func() {
		y, err := parseYear(year.Text)
		if err != nil {
			a.showError(err)
			return
		}
		startsOn, err := parseDate(start.Text)
		if err != nil {
			a.showError(err)
			return
		}
		endsOn, err := parseDate(end.Text)
		if err != nil {
			a.showError(err)
			return
		}
		if startsOn.After(endsOn) {
			a.showError(fmt.Errorf("start date must be before or equal to end date"))
			return
		}
		id, err := newID("reporting_period")
		if err != nil {
			a.showError(err)
			return
		}
		now := time.Now().UTC()
		period := domain.ReportingPeriod{
			ID:             id,
			OrganizationID: a.state.Organization.ID,
			Year:           y,
			StartsOn:       startsOn,
			EndsOn:         endsOn,
			Status:         domain.ReportingPeriodStatusDraft,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := a.state.Store.CreateReportingPeriod(period); err != nil {
			a.showError(err)
			return
		}
		a.state.ReportingPeriod = &period
		a.showInfo("Reporting period created", "Reporting period created.")
		a.refreshAndRender()
	})

	saveSettings := widget.NewButton("Save", func() {
		if a.state.ReportingPeriod == nil {
			a.showError(prerequisite("select a reporting period first"))
			return
		}
		selectedMethod := mobileMethodFromLabel(method.Selected)
		if selectedMethod == "" {
			a.showError(required("mobile combustion method"))
			return
		}
		if a.state.Settings != nil && a.state.Settings.MobileMethod != "" && a.state.Settings.MobileMethod != selectedMethod {
			hasData, err := a.hasActiveMobileCombustionData()
			if err != nil {
				a.showError(err)
				return
			}
			if hasData {
				a.showError(fmt.Errorf("mobile combustion data already exists for this reporting period; keep the current method or start a new reporting period"))
				return
			}
		}
		id := domain.ID("reporting_period_settings_" + a.state.ReportingPeriod.ID)
		if a.state.Settings != nil {
			id = a.state.Settings.ID
		}
		now := time.Now().UTC()
		if err := a.state.Store.UpsertReportingPeriodSettings(domain.ReportingPeriodSettings{
			ID:                id,
			OrganizationID:    a.state.Organization.ID,
			ReportingPeriodID: a.state.ReportingPeriod.ID,
			MobileMethod:      selectedMethod,
			CreatedAt:         now,
			UpdatedAt:         now,
		}); err != nil {
			a.showError(err)
			return
		}
		a.showInfo("Settings saved", "Reporting period settings saved.")
		a.refreshAndRender()
	})

	rows := make([][]string, 0, len(a.state.ReportingPeriods))
	for _, period := range a.state.ReportingPeriods {
		rows = append(rows, []string{
			fmt.Sprintf("%d", period.Year),
			period.StartsOn.Format("2006-01-02"),
			period.EndsOn.Format("2006-01-02"),
			reportingPeriodStatusLabel(period.Status),
		})
	}

	return screen("Reporting period",
		currentContextBar(a),
		widget.NewForm(
			widget.NewFormItem("Existing periods", selectPeriod),
			widget.NewFormItem("Year", year),
			widget.NewFormItem("Start date", start),
			widget.NewFormItem("End date", end),
			widget.NewFormItem("Mobile combustion method", method),
		),
		actionRow(create, saveSettings, widget.NewButton("Refresh", a.refreshAndRender)),
		widget.NewLabel("Reporting periods"),
		simpleTable([]string{"Year", "Start", "End", "Status"}, rows),
	)
}

func (a *App) hasActiveMobileCombustionData() (bool, error) {
	if a.state.ReportingPeriod == nil {
		return false, nil
	}
	records, err := a.state.Store.ListActiveActivityRecordsByPeriod(a.state.ReportingPeriod.ID)
	if err != nil {
		return false, err
	}
	for _, record := range records {
		if record.SourceKind == domain.ActivitySourceKindMobileFuelLitres || record.SourceKind == domain.ActivitySourceKindVehicleDistanceKM {
			return true, nil
		}
	}
	return false, nil
}
