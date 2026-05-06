package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/vocab"
)

func (a *App) electricityScreen() fyne.CanvasObject {
	if a.state.ReportingPeriod == nil || len(a.state.Facilities) == 0 {
		return prerequisiteScreen("Electricity", "Select an organization, create a facility, and create a reporting period first.")
	}

	facilitySelect, facilityIDs := a.facilitySelector()
	fullGO := widget.NewCheck("Full consumption covered by Guarantees of Origin", nil)
	goReference := widget.NewEntry()
	goMarket := widget.NewEntry()
	goCancelledAt := widget.NewEntry()
	goCancelledAt.SetPlaceHolder("YYYY-MM-DD, optional")

	return a.pasteInputScreen(pasteScreenConfig{
		Title:       "Electricity",
		Unit:        "kWh",
		Expected:    "month | consumption",
		Example:     "January\t1000\nFebruary\t1100",
		InputKind:   vocab.InputElectricityMonthlyKWh,
		Facility:    facilitySelect,
		FacilityIDs: facilityIDs,
		ExtraContent: container.NewVBox(fullGO, widget.NewForm(
			widget.NewFormItem("GO reference", goReference),
			widget.NewFormItem("GO market", goMarket),
			widget.NewFormItem("GO cancellation date", goCancelledAt),
		)),
		BeforeCommit: func(facilityID string) error {
			cancelledAt, err := optionalDate(goCancelledAt.Text)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			settings := domain.ElectricitySettings{
				ID:                    "electricity_settings_" + a.state.ReportingPeriod.ID + "_" + facilityID,
				OrganizationID:        a.state.Organization.ID,
				ReportingPeriodID:     a.state.ReportingPeriod.ID,
				FacilityID:            facilityID,
				HasGuaranteesOfOrigin: fullGO.Checked,
				GOCoverage:            domain.GOCoverageNone,
				GOReference:           cleanText(goReference.Text),
				GOMarket:              cleanText(goMarket.Text),
				GOCancelledAt:         cancelledAt,
				CreatedAt:             now,
				UpdatedAt:             now,
			}
			if fullGO.Checked {
				settings.GOCoverage = domain.GOCoverageFull
			}
			return a.state.Store.UpsertElectricitySettings(settings)
		},
		CommitContext: func(facilityID string) input.CommitContext {
			hasGO := fullGO.Checked
			coverage := domain.GOCoverageNone
			if hasGO {
				coverage = domain.GOCoverageFull
			}
			return a.commitContext(vocab.InputElectricityMonthlyKWh, &facilityID, hasGO, coverage)
		},
		SavedData: a.savedElectricityData,
	})
}

func optionalDate(value string) (*time.Time, error) {
	if cleanText(value) == "" {
		return nil, nil
	}
	parsed, err := parseDate(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (a *App) facilitySelector() (*widget.Select, map[string]string) {
	options, ids := facilityOptions(a.state.Facilities)
	selectWidget := widget.NewSelect(options, nil)
	if len(options) > 0 {
		selectWidget.SetSelected(options[0])
	}
	return selectWidget, ids
}

type pasteScreenConfig struct {
	Title         string
	Unit          string
	Expected      string
	Example       string
	InputKind     vocab.InputKind
	Facility      *widget.Select
	FacilityIDs   map[string]string
	ExtraContent  fyne.CanvasObject
	BeforeCommit  func(facilityID string) error
	CommitContext func(facilityID string) input.CommitContext
	SavedData     func(facilityID string) (fyne.CanvasObject, error)
}

func (a *App) pasteInputScreen(config pasteScreenConfig) fyne.CanvasObject {
	paste := widget.NewMultiLineEntry()
	paste.SetPlaceHolder(config.Example)
	summary := widget.NewLabel("Paste data from a spreadsheet, then validate.")
	preview := container.NewStack(widget.NewLabel("Parsed preview will appear after validation."))
	saved := container.NewStack(widget.NewLabel("Loading saved data."))
	var parsed input.ParseResult
	parsedValid := false
	paste.OnChanged = func(string) { parsedValid = false }

	selectedFacilityID := func() string {
		if config.Facility == nil {
			return ""
		}
		return config.FacilityIDs[config.Facility.Selected]
	}
	loadSaved := func() {
		if config.SavedData == nil {
			saved.Objects = []fyne.CanvasObject{emptySavedData()}
			saved.Refresh()
			return
		}
		object, err := config.SavedData(selectedFacilityID())
		if err != nil {
			saved.Objects = []fyne.CanvasObject{widget.NewLabel("Saved data could not be loaded: " + err.Error())}
			saved.Refresh()
			return
		}
		saved.Objects = []fyne.CanvasObject{object}
		saved.Refresh()
	}
	if config.Facility != nil {
		config.Facility.OnChanged = func(string) { loadSaved() }
	}
	loadSaved()

	validate := widget.NewButton("Validate", func() {
		parsed = input.Parse(config.InputKind, paste.Text)
		parsedValid = true
		summary.SetText(parseSummary(parsed))
		preview.Objects = []fyne.CanvasObject{parsedRowsPreview(config.InputKind, parsed)}
		preview.Refresh()
	})

	commit := widget.NewButton("Save / replace", func() {
		if !parsedValid {
			a.showError(required("validated paste"))
			return
		}
		if parsed.RowsError > 0 {
			a.showError(fmt.Errorf("parsed rows contain blocking errors"))
			return
		}
		facilityID := ""
		if config.Facility != nil {
			facilityID = config.FacilityIDs[config.Facility.Selected]
			if facilityID == "" {
				a.showError(required("facility"))
				return
			}
		}
		if config.BeforeCommit != nil {
			if err := config.BeforeCommit(facilityID); err != nil {
				a.showError(err)
				return
			}
		}
		c := config.CommitContext(facilityID)
		result, err := input.CommitParsedInput(context.Background(), a.state.Store, c, parsed)
		if err != nil {
			a.showError(err)
			return
		}
		summary.SetText("Saved " + formatInt(len(result.ActivityRecordIDs)) + " record(s).")
		parsedValid = false
		loadSaved()
		a.showInfo("Data saved", "Saved data is now visible on this screen.")
	})

	clearPaste := widget.NewButton("Clear paste", func() {
		paste.SetText("")
		parsed = input.ParseResult{}
		parsedValid = false
		summary.SetText("Paste data from a spreadsheet, then validate.")
		preview.Objects = []fyne.CanvasObject{widget.NewLabel("Parsed preview will appear after validation.")}
		preview.Refresh()
	})

	formItems := []*widget.FormItem{}
	if config.Facility != nil {
		formItems = append(formItems, widget.NewFormItem("Facility", config.Facility))
	}
	if config.Unit != "" {
		formItems = append(formItems, widget.NewFormItem("Unit", widget.NewLabel(config.Unit)))
	}
	formItems = append(formItems, widget.NewFormItem("Expected format", widget.NewLabel(config.Expected)))

	objects := []fyne.CanvasObject{
		currentContextBar(a),
		widget.NewForm(formItems...),
	}
	if config.ExtraContent != nil {
		objects = append(objects, config.ExtraContent)
	}
	objects = append(objects,
		widget.NewLabel("Example"),
		widget.NewLabel(config.Example),
		widget.NewLabel("Paste from spreadsheet"),
		paste,
		actionRow(validate, commit, clearPaste),
		widget.NewLabel("Validation summary"),
		summary,
		widget.NewLabel("Parsed preview"),
		preview,
		saved,
	)
	return screen(config.Title, objects...)
}

func (a *App) commitContext(kind vocab.InputKind, facilityID *string, hasGO bool, coverage domain.GOCoverage) input.CommitContext {
	period := a.state.ReportingPeriod
	mobileMethod := domain.MobileMethod("")
	if a.state.Settings != nil {
		mobileMethod = a.state.Settings.MobileMethod
	}
	return input.CommitContext{
		OrganizationID:        a.state.Organization.ID,
		ReportingPeriodID:     period.ID,
		FacilityID:            facilityID,
		ReportingYear:         period.Year,
		PeriodStart:           period.StartsOn,
		PeriodEnd:             period.EndsOn,
		InputKind:             kind,
		MobileMethod:          mobileMethod,
		HasGuaranteesOfOrigin: hasGO,
		GOCoverage:            coverage,
	}
}

func formatInt(value int) string {
	return fmt.Sprintf("%d", value)
}
