package ui

import (
	"fmt"
	"strings"

	"ghgo/internal/domain"
	"ghgo/internal/vocab"
)

func organizationDisplayLabel(organization domain.Organization) string {
	if strings.TrimSpace(organization.Name) != "" {
		return organization.Name
	}
	return "Unnamed organization"
}

func currentOrganizationLabel(organization *domain.Organization) string {
	if organization == nil {
		return "No organization selected"
	}
	return organizationDisplayLabel(*organization)
}

func facilityDisplayLabel(facility domain.Facility) string {
	name := strings.TrimSpace(facility.Name)
	if name == "" {
		name = "Unnamed facility"
	}
	if strings.TrimSpace(facility.CountryCode) == "" {
		return name
	}
	return name + ", " + strings.ToUpper(facility.CountryCode)
}

func reportingPeriodDisplayLabel(period domain.ReportingPeriod) string {
	if period.Year > 0 && period.StartsOn.Month() == 1 && period.StartsOn.Day() == 1 && period.EndsOn.Month() == 12 && period.EndsOn.Day() == 31 {
		return fmt.Sprintf("%d", period.Year)
	}
	if period.Year > 0 {
		return fmt.Sprintf("%d - %s-%s", period.Year, period.StartsOn.Format("Jan 2"), period.EndsOn.Format("Jan 2"))
	}
	return "Reporting period"
}

func currentReportingPeriodLabel(period *domain.ReportingPeriod) string {
	if period == nil {
		return "No reporting period selected"
	}
	return reportingPeriodDisplayLabel(*period)
}

func factorSetDisplayLabel(factorSet domain.FactorSet) string {
	if strings.TrimSpace(factorSet.Name) != "" {
		return factorSet.Name
	}
	if factorSet.Source != "" && factorSet.Year > 0 {
		return fmt.Sprintf("%s %d", factorSet.Source, factorSet.Year)
	}
	return "Emission factor set"
}

func calculationRunDisplayLabel(run domain.CalculationRun) string {
	timestamp := run.StartedAt
	if run.CompletedAt != nil {
		timestamp = *run.CompletedAt
	}
	return calculationRunStatusLabel(run.Status) + " - " + timestamp.Local().Format("2006-01-02 15:04")
}

func calculationRunStatusLabel(status domain.CalculationRunStatus) string {
	switch status {
	case domain.CalculationRunStatusCompleted:
		return "Completed"
	case domain.CalculationRunStatusRunning:
		return "Running"
	case domain.CalculationRunStatusFailed:
		return "Failed"
	}
	return "Unknown"
}

func reportingPeriodStatusLabel(status domain.ReportingPeriodStatus) string {
	switch status {
	case domain.ReportingPeriodStatusDraft:
		return "Draft"
	case domain.ReportingPeriodStatusLocked:
		return "Locked"
	case domain.ReportingPeriodStatusArchived:
		return "Archived"
	}
	return "Unknown"
}

func activityVectorLabel(vector domain.ActivityVector) string {
	switch vector {
	case domain.ActivityVectorElectricity:
		return "Electricity"
	case domain.ActivityVectorNaturalGas:
		return "Natural gas"
	case domain.ActivityVectorMobileCombustion:
		return "Mobile combustion"
	case domain.ActivityVectorRefrigerants:
		return "Refrigerants"
	}
	return string(vector)
}

func activityMethodLabel(method domain.ActivityMethod) string {
	switch method {
	case domain.ActivityMethodLocationBased:
		return "Location-based"
	case domain.ActivityMethodMarketBased:
		return "Market-based"
	case domain.ActivityMethodFuelBased:
		return "Fuel consumed"
	case domain.ActivityMethodDistanceBased:
		return "Distance travelled"
	case domain.ActivityMethodDirectGWP:
		return "Direct GWP"
	}
	return string(method)
}

func fuelTypeLabel(value string) string {
	label := vocab.FuelType(value).EnglishLabel()
	if label != "" {
		return label
	}
	return value
}

func vehicleTypeLabel(value string) string {
	label := vocab.VehicleType(value).EnglishLabel()
	if label != "" {
		return label
	}
	return value
}

func vehicleSizeClassLabel(vehicleType string, value string) string {
	label := vocab.VehicleSizeClass(value).EnglishLabel(vocab.VehicleType(vehicleType))
	if label != "" {
		return label
	}
	return value
}

func uniqueOptionLabel(base string, used map[string]int) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "Unnamed"
	}
	if used[base] == 0 {
		used[base] = 1
		return base
	}
	for index := 2; ; index++ {
		label := fmt.Sprintf("%s %d", base, index)
		if used[label] == 0 {
			used[base]++
			used[label] = 1
			return label
		}
	}
}
