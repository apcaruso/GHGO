package report

import (
	"encoding/json"
	"fmt"

	"ghgo/internal/domain"
	"ghgo/internal/ports"
	"ghgo/internal/vocab"
)

type MethodologyTable struct {
	Rows []MethodologyRow
}

type MethodologyRow struct {
	Area   string
	Method string
	Source string
	Notes  string
}

type ValidationNotesTable struct {
	Rows []ValidationNoteRow
}

type ValidationNoteRow struct {
	Severity string
	Code     string
	Message  string
}

func buildMethodology(data reportData) MethodologyTable {
	mobileMethod, _ := detectMobileMethod(data.rows)
	if mobileMethod == "" {
		mobileMethod = snapshotMobileMethod(data.run.SettingsSnapshotJSON)
	}

	electricityMethod := string(domain.ActivityMethodLocationBased)
	electricityNotes := "location-based primary"
	electricitySource := firstFactorSource(data.rows, domain.ActivityVectorElectricity)
	if data.hasLocationBasedComparison {
		electricityMethod = string(domain.ActivityMethodMarketBased)
		electricitySource = "Guarantees of Origin"
		electricityNotes = "market-based primary with full Guarantees of Origin; location-based shown for comparison"
	}

	rows := []MethodologyRow{
		{
			Area:   "Factor set",
			Method: "stored calculation_results",
			Source: data.factorSet.Source,
			Notes:  fmt.Sprintf("year %d; version %s", data.factorSet.Year, data.factorSet.Version),
		},
		{
			Area:   "Electricity",
			Method: electricityMethod,
			Source: electricitySource,
			Notes:  electricityNotes,
		},
		{
			Area:   "Natural gas",
			Method: string(domain.ActivityMethodFuelBased),
			Source: firstFactorSource(data.rows, domain.ActivityVectorNaturalGas),
			Notes:  "Smc",
		},
		{
			Area:   "Mobile combustion",
			Method: string(mobileMethod),
			Source: firstFactorSource(data.rows, domain.ActivityVectorMobileCombustion),
			Notes:  mobileUnitNote(mobileMethod),
		},
		{
			Area:   "Refrigerants",
			Method: string(domain.ActivityMethodDirectGWP),
			Source: firstFactorSource(data.rows, domain.ActivityVectorRefrigerants),
			Notes:  "kg",
		},
		{
			Area:   "Units",
			Method: "activity units",
			Source: "",
			Notes:  "electricity kWh; natural gas Smc; mobile L or km; refrigerants kg",
		},
	}
	return MethodologyTable{Rows: rows}
}

func buildValidationNotes(data reportData) ValidationNotesTable {
	rows := []ValidationNoteRow{}
	if data.hasLocationBasedComparison {
		rows = append(rows, ValidationNoteRow{
			Severity: "info",
			Code:     "location_based_comparison",
			Message:  "Location-based electricity emissions are shown for comparison.",
		})
	}
	if hasZeroMarketBasedElectricity(data.rows) {
		rows = append(rows, ValidationNoteRow{
			Severity: "info",
			Code:     "go_market_zero",
			Message:  "Market-based electricity emissions are zero because full Guarantees of Origin coverage was declared.",
		})
	}
	if data.primaryTotalKgCO2e == 0 {
		rows = append(rows, ValidationNoteRow{
			Severity: "warning",
			Code:     "zero_primary_total",
			Message:  "Primary total emissions are zero.",
		})
	}
	mobileMethod, _ := detectMobileMethod(data.rows)
	if mobileMethod == domain.ActivityMethodDistanceBased && hasBEVMobileDistance(data.rows) {
		rows = append(rows, ValidationNoteRow{
			Severity: "info",
			Code:     "bev_distance_no_charging_estimate",
			Message:  "Battery electric vehicle charging electricity is not estimated from km.",
		})
	}
	return ValidationNotesTable{Rows: rows}
}

func firstFactorSource(rows []ports.ReportResultRow, vector domain.ActivityVector) string {
	for _, row := range rows {
		result := row.CalculationResult
		if result.Vector == vector && result.FactorSource != "" {
			return result.FactorSource
		}
	}
	return ""
}

func mobileUnitNote(method domain.ActivityMethod) string {
	switch method {
	case domain.ActivityMethodFuelBased:
		return "L"
	case domain.ActivityMethodDistanceBased:
		return "km"
	}
	return ""
}

func snapshotMobileMethod(snapshotJSON string) domain.ActivityMethod {
	var snapshot struct {
		MobileMethod string `json:"mobile_method"`
	}
	if err := json.Unmarshal([]byte(snapshotJSON), &snapshot); err != nil {
		return ""
	}
	switch domain.MobileMethod(snapshot.MobileMethod) {
	case domain.MobileMethodFuelBased:
		return domain.ActivityMethodFuelBased
	case domain.MobileMethodDistanceBased:
		return domain.ActivityMethodDistanceBased
	}
	return ""
}

func hasZeroMarketBasedElectricity(rows []ports.ReportResultRow) bool {
	for _, row := range rows {
		result := row.CalculationResult
		if result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodMarketBased && result.EmissionsKgCO2e == 0 {
			return true
		}
	}
	return false
}

func hasBEVMobileDistance(rows []ports.ReportResultRow) bool {
	for _, row := range rows {
		record := row.ActivityRecord
		result := row.CalculationResult
		if result.Vector == domain.ActivityVectorMobileCombustion &&
			result.Method == domain.ActivityMethodDistanceBased &&
			record.FuelType == string(vocab.FuelBEV) {
			return true
		}
	}
	return false
}
