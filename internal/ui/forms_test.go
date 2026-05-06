package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/container"

	"ghgo/internal/domain"
)

func TestValidateCountryCode(t *testing.T) {
	got, err := validateCountryCode("it")
	if err != nil {
		t.Fatalf("validate country code: %v", err)
	}
	if got != "IT" {
		t.Fatalf("country code = %q, want IT", got)
	}
	if _, err := validateCountryCode("ITA"); err == nil {
		t.Fatalf("validate country code with three letters returned nil error")
	}
}

func TestMobileMethodLabels(t *testing.T) {
	if mobileMethodFromLabel("Fuel consumed") != domain.MobileMethodFuelBased {
		t.Fatalf("Fuel consumed did not map to fuel_based")
	}
	if mobileMethodFromLabel("Distance travelled") != domain.MobileMethodDistanceBased {
		t.Fatalf("Distance travelled did not map to distance_based")
	}
	if mobileMethodLabel(domain.MobileMethodFuelBased) != "Fuel consumed" {
		t.Fatalf("fuel_based label mismatch")
	}
}

func TestParseDate(t *testing.T) {
	date, err := parseDate("2026-01-02")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	if date.Year() != 2026 || date.Month() != 1 || date.Day() != 2 {
		t.Fatalf("date = %v, want 2026-01-02", date)
	}
	if _, err := parseDate("02/01/2026"); err == nil {
		t.Fatalf("parse invalid date returned nil error")
	}
}

func TestDisplayLabelsHideInternalIDs(t *testing.T) {
	organization := domain.Organization{ID: "organization_123", Name: "prova"}
	if got := organizationDisplayLabel(organization); got != "prova" {
		t.Fatalf("organization label = %q, want prova", got)
	}

	facility := domain.Facility{ID: "facility_123", Name: "Alessandria", CountryCode: "IT"}
	if got := facilityDisplayLabel(facility); got != "Alessandria, IT" {
		t.Fatalf("facility label = %q, want Alessandria, IT", got)
	}

	period := domain.ReportingPeriod{
		ID:       "reporting_period_123",
		Year:     2025,
		StartsOn: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndsOn:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	if got := reportingPeriodDisplayLabel(period); got != "2025" {
		t.Fatalf("reporting period label = %q, want 2025", got)
	}

	factorSet := domain.FactorSet{ID: "factor_set_defra_2025", Name: "DEFRA/DESNZ 2025", Source: "DEFRA", Year: 2025, Version: "2025"}
	if got := factorSetDisplayLabel(factorSet); got != "DEFRA/DESNZ 2025" {
		t.Fatalf("factor set label = %q, want DEFRA/DESNZ 2025", got)
	}
}

func TestSimpleTableKeepsVerticalMinSize(t *testing.T) {
	table := simpleTable([]string{"Month", "Consumption kWh"}, [][]string{{"January", "100"}, {"February", "200"}})
	scroll, ok := table.(*container.Scroll)
	if !ok {
		t.Fatalf("simpleTable returned %T, want scroll", table)
	}
	if scroll.Direction != container.ScrollHorizontalOnly {
		t.Fatalf("scroll direction = %v, want horizontal only", scroll.Direction)
	}
	if scroll.MinSize().Height != scroll.Content.MinSize().Height {
		t.Fatalf("scroll min height = %v, content min height = %v", scroll.MinSize().Height, scroll.Content.MinSize().Height)
	}
}
