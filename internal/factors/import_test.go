//go:build ghgo_devtools

package factors

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"ghgo/internal/domain"
)

func TestImportDEFRA2025SyntheticWorkbook(t *testing.T) {
	st := newTestStore(t)
	path := writeSyntheticWorkbook(t, syntheticHeaders(), syntheticRows())

	summary, err := ImportDEFRA2025(context.Background(), st, path, ImportOptions{})
	if err != nil {
		t.Fatalf("import DEFRA 2025: %v", err)
	}
	if summary.FactorSetID == "" {
		t.Fatalf("factor set id is empty")
	}
	if summary.RowsRead != 11 {
		t.Fatalf("rows read = %d, want 11", summary.RowsRead)
	}
	if summary.RowsImported != 8 {
		t.Fatalf("rows imported = %d, want 8", summary.RowsImported)
	}
	if summary.RowsSkipped != 3 {
		t.Fatalf("rows skipped = %d, want 3", summary.RowsSkipped)
	}
	if !hasWarning(summary.Warnings, "no Smc-compatible natural gas factor") {
		t.Fatalf("warnings = %#v, want natural gas warning", summary.Warnings)
	}

	factorSet, err := st.GetFactorSet(domain.ID(summary.FactorSetID))
	if err != nil {
		t.Fatalf("get factor set: %v", err)
	}
	if factorSet.Name != "DEFRA/DESNZ 2025" || factorSet.Source != "DEFRA" || factorSet.Year != 2025 || factorSet.Version != "2025" {
		t.Fatalf("factor set = %#v, want DEFRA/DESNZ 2025", factorSet)
	}

	factors, err := st.ListEmissionFactorsBySet(factorSet.ID)
	if err != nil {
		t.Fatalf("list factors: %v", err)
	}
	if len(factors) != 8 {
		t.Fatalf("factor count = %d, want 8", len(factors))
	}
	if hasFactor(factors, func(f domain.EmissionFactor) bool { return f.InputUnit == "miles" || f.Scope == domain.Scope(3) }) {
		t.Fatalf("unsupported scope 3 or miles factor was imported: %#v", factors)
	}

	diesel := findFactor(t, factors, func(f domain.EmissionFactor) bool {
		return f.ActivityType == "diesel_mobile"
	})
	if diesel.FuelType != "diesel" || diesel.InputUnit != "L" || diesel.FactorUnit != "kgCO2e/L" {
		t.Fatalf("diesel factor = %#v, want diesel L kgCO2e/L", diesel)
	}

	carSmallPetrol := findFactor(t, factors, func(f domain.EmissionFactor) bool {
		return f.ActivityType == "vehicle_distance" && f.VehicleType == "car" && f.VehicleSizeClass == "small" && f.FuelType == "petrol"
	})
	if carSmallPetrol.InputUnit != "km" {
		t.Fatalf("car small petrol input unit = %q, want km", carSmallPetrol.InputUnit)
	}

	refrigerant := findFactor(t, factors, func(f domain.EmissionFactor) bool {
		return f.ActivityType == "refrigerant_leakage" && f.Substance == "R410A"
	})
	if refrigerant.InputUnit != "kg" {
		t.Fatalf("R410A input unit = %q, want kg", refrigerant.InputUnit)
	}

	secondSummary, err := ImportDEFRA2025(context.Background(), st, path, ImportOptions{})
	if err != nil {
		t.Fatalf("second import DEFRA 2025: %v", err)
	}
	if secondSummary.FactorSetID != summary.FactorSetID {
		t.Fatalf("second factor set id = %q, want %q", secondSummary.FactorSetID, summary.FactorSetID)
	}
	count, err := st.CountEmissionFactorsBySet(factorSet.ID)
	if err != nil {
		t.Fatalf("count factors after second import: %v", err)
	}
	if count != 8 {
		t.Fatalf("factor count after second import = %d, want 8", count)
	}

	forceSummary, err := ImportDEFRA2025(context.Background(), st, path, ImportOptions{Force: true})
	if err != nil {
		t.Fatalf("force import DEFRA 2025: %v", err)
	}
	if forceSummary.FactorSetID != summary.FactorSetID {
		t.Fatalf("force factor set id = %q, want %q", forceSummary.FactorSetID, summary.FactorSetID)
	}
	if forceSummary.RowsImported != 8 {
		t.Fatalf("force rows imported = %d, want 8", forceSummary.RowsImported)
	}
	count, err = st.CountEmissionFactorsBySet(factorSet.ID)
	if err != nil {
		t.Fatalf("count factors after force import: %v", err)
	}
	if count != 8 {
		t.Fatalf("factor count after force import = %d, want 8", count)
	}
	factorSets, err := st.ListFactorSets()
	if err != nil {
		t.Fatalf("list factor sets: %v", err)
	}
	if len(factorSets) != 1 {
		t.Fatalf("factor set count = %d, want 1", len(factorSets))
	}
}

func TestImportDEFRA2025MissingRequiredHeaders(t *testing.T) {
	st := newTestStore(t)
	headers := syntheticHeaders()
	path := writeSyntheticWorkbook(t, headers[:len(headers)-1], nil)

	_, err := ImportDEFRA2025(context.Background(), st, path, ImportOptions{})
	if err == nil {
		t.Fatalf("missing headers import succeeded")
	}
	if !strings.Contains(err.Error(), "conversion_factor") {
		t.Fatalf("error = %q, want missing conversion_factor", err.Error())
	}
}

func syntheticHeaders() []any {
	return []any{"Scope", "Level_1", "Level-2", "Level 3", "Level 4", "Column Text", "UOM", "GHG/Unit", "GHG Conversion Factor"}
}

func syntheticRows() [][]any {
	return [][]any{
		{"Scope 2", "UK electricity", "Electricity generated", "Electricity: UK", "", "Location-based", "kWh", "kg CO2e", 0.207},
		{"Scope 1", "Fuels", "Liquid fuels", "Diesel (average biofuel blend)", "", "Direct", "litres", "kg CO2e", 2.51},
		{"Scope 1", "Fuels", "Liquid fuels", "Petrol (average biofuel blend)", "", "Direct", "litres", "kg CO2e", 2.18},
		{"Scope 1", "Passenger vehicles", "Cars", "Small car", "Petrol", "", "km", "kg CO2e", 0.15},
		{"Scope 1", "Passenger vehicles", "Cars", "Medium car", "Diesel", "", "km", "kg CO2e", 0.17},
		{"Scope 1", "Delivery vehicles", "Vans", "Class II", "Diesel", "", "km", "kg CO2e", 0.25},
		{"Scope 1", "Passenger vehicles", "Motorbike", "Average", "", "", "km", "kg CO2e", 0.10},
		{"Scope 1", "Refrigerants", "HFCs", "R410A", "", "", "kg", "kg CO2e", 2088},
		{"Scope 3", "Fuels", "Liquid fuels", "Diesel", "", "", "litres", "kg CO2e", 2.50},
		{"Scope 1", "Passenger vehicles", "Cars", "Small car", "Petrol", "", "miles", "kg CO2e", 0.24},
		{"Scope 1", "Fuels", "Liquid fuels", "Diesel", "", "", "litres", "kg CO2e", ""},
	}
}

func writeSyntheticWorkbook(t *testing.T, headers []any, rows [][]any) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "defra2025-mini.xlsx")
	workbook := excelize.NewFile()
	defer workbook.Close()

	notes := workbook.GetSheetName(0)
	if err := workbook.SetCellValue(notes, "A1", "not the flat format"); err != nil {
		t.Fatalf("write notes sheet: %v", err)
	}
	factorsSheet := "Flat Format"
	index, err := workbook.NewSheet(factorsSheet)
	if err != nil {
		t.Fatalf("create factors sheet: %v", err)
	}
	workbook.SetActiveSheet(index)

	if err := workbook.SetSheetRow(factorsSheet, "A1", &headers); err != nil {
		t.Fatalf("write headers: %v", err)
	}
	for i, row := range rows {
		cell := fmt.Sprintf("A%d", i+2)
		if err := workbook.SetSheetRow(factorsSheet, cell, &row); err != nil {
			t.Fatalf("write row %d: %v", i+1, err)
		}
	}
	if err := workbook.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}

	return path
}

func findFactor(t *testing.T, factors []domain.EmissionFactor, match func(domain.EmissionFactor) bool) domain.EmissionFactor {
	t.Helper()
	for _, factor := range factors {
		if match(factor) {
			return factor
		}
	}
	t.Fatalf("factor not found in %#v", factors)
	return domain.EmissionFactor{}
}

func hasFactor(factors []domain.EmissionFactor, match func(domain.EmissionFactor) bool) bool {
	for _, factor := range factors {
		if match(factor) {
			return true
		}
	}
	return false
}

func hasWarning(warnings []string, want string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, want) {
			return true
		}
	}
	return false
}
