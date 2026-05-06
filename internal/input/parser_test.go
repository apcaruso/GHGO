package input

import (
	"testing"

	"ghgo/internal/vocab"
)

func TestParseElectricity(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		rowsTotal  int
		rowsValid  int
		rowsError  int
		rowIndex   int
		normalized map[string]string
		errorCode  string
		warnCode   string
	}{
		{
			name:      "tab delimited",
			raw:       "January\t12000\nFebruary\t11800",
			rowsTotal: 2,
			rowsValid: 2,
			normalized: map[string]string{
				"month_name": "January",
				"amount":     "12000",
				"unit":       "kWh",
			},
		},
		{
			name:      "semicolon delimited",
			raw:       "January;12000",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"month_number": "1",
				"unit":         "kWh",
			},
		},
		{
			name:      "pipe delimited",
			raw:       "January | 12000",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"month_number": "1",
				"amount":       "12000",
				"unit":         "kWh",
			},
		},
		{
			name:      "italian month",
			raw:       "gennaio\t12000",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"month_name": "January",
			},
		},
		{
			name:      "italian pipe header ignored",
			raw:       "mese | consumo\ngennaio | 12000",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"month_name": "January",
				"amount":     "12000",
				"unit":       "kWh",
			},
		},
		{
			name:      "first header ignored",
			raw:       "month\tconsumption\nJanuary\t12000",
			rowsTotal: 1,
			rowsValid: 1,
			rowIndex:  0,
			normalized: map[string]string{
				"month_name": "January",
			},
		},
		{
			name:      "header in body rejected",
			raw:       "January\t12000\nmonth\tconsumption",
			rowsTotal: 2,
			rowsValid: 1,
			rowsError: 1,
			rowIndex:  1,
			errorCode: CodeHeaderInBody,
		},
		{
			name:      "duplicate month rejected",
			raw:       "Jan\t12000\nJanuary\t11800",
			rowsTotal: 2,
			rowsValid: 1,
			rowsError: 1,
			rowIndex:  1,
			errorCode: CodeDuplicateMonth,
		},
		{
			name:      "unit inside number rejected",
			raw:       "Jan\t12000 kWh",
			rowsTotal: 1,
			rowsError: 1,
			errorCode: CodeUnitInNumber,
		},
		{
			name:      "empty amount rejected",
			raw:       "Jan\t",
			rowsTotal: 1,
			rowsError: 1,
			errorCode: CodeEmptyField,
		},
		{
			name:      "zero warns",
			raw:       "Jan\t0",
			rowsTotal: 1,
			rowsValid: 1,
			warnCode:  CodeZeroValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(vocab.InputElectricityMonthlyKWh, tt.raw)
			requireCounts(t, result, tt.rowsTotal, tt.rowsValid, tt.rowsError)
			row := result.Rows[tt.rowIndex]
			for key, want := range tt.normalized {
				requireNormalized(t, row, key, want)
			}
			if tt.errorCode != "" {
				requireError(t, row, tt.errorCode)
			}
			if tt.warnCode != "" {
				requireWarning(t, row, tt.warnCode)
			}
		})
	}
}

func TestParseNaturalGas(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		rowsValid int
		rowsError int
		rowIndex  int
		errorCode string
		warnCode  string
	}{
		{name: "tab delimited", raw: "January\t830", rowsValid: 1},
		{name: "semicolon delimited", raw: "January;830", rowsValid: 1},
		{name: "pipe delimited with italian header", raw: "mese | consumo\ngennaio | 830", rowsValid: 1},
		{name: "italian month", raw: "gennaio\t830", rowsValid: 1},
		{name: "first header ignored", raw: "month\tsmc\nJanuary\t830", rowsValid: 1},
		{name: "header in body rejected", raw: "January\t830\nmonth\tsmc", rowsValid: 1, rowsError: 1, rowIndex: 1, errorCode: CodeHeaderInBody},
		{name: "duplicate month rejected", raw: "Jan\t830\nJanuary\t760", rowsValid: 1, rowsError: 1, rowIndex: 1, errorCode: CodeDuplicateMonth},
		{name: "unit inside number rejected", raw: "Jan\t830 Smc", rowsError: 1, errorCode: CodeUnitInNumber},
		{name: "zero warns", raw: "Jan\t0", rowsValid: 1, warnCode: CodeZeroValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(vocab.InputNaturalGasMonthlySmc, tt.raw)
			requireCounts(t, result, tt.rowsValid+tt.rowsError, tt.rowsValid, tt.rowsError)
			row := result.Rows[tt.rowIndex]
			if tt.rowsError == 0 {
				requireNormalized(t, row, "unit", "Smc")
			}
			if tt.errorCode != "" {
				requireError(t, row, tt.errorCode)
			}
			if tt.warnCode != "" {
				requireWarning(t, row, tt.warnCode)
			}
		})
	}
}

func TestParseMobileFuel(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		rowsTotal int
		rowsValid int
		rowsError int
		rowIndex  int
		fuelType  string
		errorCode string
		warnCode  string
	}{
		{name: "diesel", raw: "Diesel\t4200", rowsTotal: 1, rowsValid: 1, fuelType: "diesel"},
		{name: "gasolio", raw: "gasolio\t4200", rowsTotal: 1, rowsValid: 1, fuelType: "diesel"},
		{name: "benzina", raw: "benzina\t1300", rowsTotal: 1, rowsValid: 1, fuelType: "petrol"},
		{name: "reject bev", raw: "BEV\t100", rowsTotal: 1, rowsError: 1, fuelType: "bev", errorCode: CodeUnsupportedFuelForMethod},
		{name: "reject phev", raw: "PHEV\t100", rowsTotal: 1, rowsError: 1, fuelType: "phev", errorCode: CodeUnsupportedFuelForMethod},
		{name: "reject hybrid", raw: "Hybrid\t100", rowsTotal: 1, rowsError: 1, fuelType: "hybrid", errorCode: CodeUnsupportedFuelForMethod},
		{name: "reject unknown", raw: "Unknown\t100", rowsTotal: 1, rowsError: 1, fuelType: "unknown", errorCode: CodeUnsupportedFuelForMethod},
		{name: "duplicate fuel type warns", raw: "Diesel\t4200\ngasolio\t200", rowsTotal: 2, rowsValid: 2, rowIndex: 1, fuelType: "diesel", warnCode: CodeDuplicateFuelType},
		{name: "negative rejected", raw: "Diesel\t-1", rowsTotal: 1, rowsError: 1, errorCode: CodeNegativeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(vocab.InputMobileFuelLitres, tt.raw)
			requireCounts(t, result, tt.rowsTotal, tt.rowsValid, tt.rowsError)
			row := result.Rows[tt.rowIndex]
			if tt.fuelType != "" {
				requireNormalized(t, row, "fuel_type", tt.fuelType)
			}
			if tt.rowsValid > 0 && tt.rowsError == 0 {
				requireNormalized(t, row, "unit", "L")
			}
			if tt.errorCode != "" {
				requireError(t, row, tt.errorCode)
			}
			if tt.warnCode != "" {
				requireWarning(t, row, tt.warnCode)
			}
		})
	}
}

func TestParseVehicleDistance(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		rowsTotal  int
		rowsValid  int
		rowsError  int
		rowIndex   int
		normalized map[string]string
		errorCode  string
		warnCode   string
	}{
		{
			name:      "full format",
			raw:       "Fiat Panda\tAB123CD\tCar\tSmall\tPetrol\t18400",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"vehicle_name":       "Fiat Panda",
				"plate":              "AB123CD",
				"vehicle_type":       "car",
				"vehicle_size_class": "small",
				"fuel_type":          "petrol",
				"unit":               "km",
			},
		},
		{
			name:      "italian car aliases",
			raw:       "Fiat Panda\tAB123CD\tauto\tpiccola\tbenzina\t18400",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"vehicle_type":       "car",
				"vehicle_size_class": "small",
				"fuel_type":          "petrol",
			},
		},
		{
			name:      "italian van aliases",
			raw:       "Fiat Doblo\tEF789GH\tfurgone\tclasse 2\tgasolio\t19800",
			rowsTotal: 1,
			rowsValid: 1,
			normalized: map[string]string{
				"vehicle_type":       "van",
				"vehicle_size_class": "class_ii",
				"fuel_type":          "diesel",
			},
		},
		{name: "wrong column count", raw: "BMW 3 Series AB123CD Car Medium Diesel 22100", rowsTotal: 1, rowsError: 1, errorCode: CodeWrongColumnCount},
		{name: "van medium rejected", raw: "Van 1\tAA111AA\tVan\tMedium\tDiesel\t100", rowsTotal: 1, rowsError: 1, errorCode: CodeIncompatibleVehicleSizeClass},
		{name: "car class i rejected", raw: "Car 1\tAA111AA\tCar\tclass_i\tDiesel\t100", rowsTotal: 1, rowsError: 1, errorCode: CodeIncompatibleVehicleSizeClass},
		{name: "bev warns", raw: "EV\tAA111AA\tCar\tSmall\tBEV\t100", rowsTotal: 1, rowsValid: 1, warnCode: CodeBEVScope2NotEstimated},
		{name: "phev warns", raw: "PHEV\tAA111AA\tCar\tSmall\tPHEV\t100", rowsTotal: 1, rowsValid: 1, warnCode: CodePHEVAverageFactor},
		{name: "duplicate plate warns", raw: "One\tAA111AA\tCar\tSmall\tPetrol\t100\nTwo\tAA111AA\tCar\tMedium\tDiesel\t200", rowsTotal: 2, rowsValid: 2, rowIndex: 1, warnCode: CodeDuplicatePlate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(vocab.InputVehicleDistanceKm, tt.raw)
			requireCounts(t, result, tt.rowsTotal, tt.rowsValid, tt.rowsError)
			row := result.Rows[tt.rowIndex]
			for key, want := range tt.normalized {
				requireNormalized(t, row, key, want)
			}
			if tt.errorCode != "" {
				requireError(t, row, tt.errorCode)
			}
			if tt.warnCode != "" {
				requireWarning(t, row, tt.warnCode)
			}
		})
	}
}

func TestParseRefrigerants(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		rowsTotal int
		rowsValid int
		rowsError int
		rowIndex  int
		substance string
		amount    string
		errorCode string
		warnCode  string
	}{
		{name: "comma decimal", raw: "R410A\t3,2", rowsTotal: 1, rowsValid: 1, substance: "R410A", amount: "3.2"},
		{name: "normalize hyphen", raw: "r-410a\t3.2", rowsTotal: 1, rowsValid: 1, substance: "R410A"},
		{name: "reject co2", raw: "CO2\t3.2", rowsTotal: 1, rowsError: 1, errorCode: CodeInvalidRefrigerant},
		{name: "duplicate refrigerant warns", raw: "R410A\t3.2\nr-410a\t1.1", rowsTotal: 2, rowsValid: 2, rowIndex: 1, substance: "R410A", warnCode: CodeDuplicateRefrigerant},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(vocab.InputRefrigerantsAnnualKg, tt.raw)
			requireCounts(t, result, tt.rowsTotal, tt.rowsValid, tt.rowsError)
			row := result.Rows[tt.rowIndex]
			if tt.substance != "" {
				requireNormalized(t, row, "substance", tt.substance)
			}
			if tt.amount != "" {
				requireNormalized(t, row, "amount", tt.amount)
			}
			if tt.rowsValid > 0 && tt.rowsError == 0 {
				requireNormalized(t, row, "unit", "kg")
			}
			if tt.errorCode != "" {
				requireError(t, row, tt.errorCode)
			}
			if tt.warnCode != "" {
				requireWarning(t, row, tt.warnCode)
			}
		})
	}
}

func TestDelimiterBehavior(t *testing.T) {
	tests := []struct {
		name      string
		kind      vocab.InputKind
		raw       string
		rowsValid int
		rowsError int
		code      string
	}{
		{name: "tab", kind: vocab.InputRefrigerantsAnnualKg, raw: "R410A\t3.2", rowsValid: 1},
		{name: "semicolon", kind: vocab.InputRefrigerantsAnnualKg, raw: "R410A;3,2", rowsValid: 1},
		{name: "two or more spaces", kind: vocab.InputElectricityMonthlyKWh, raw: "January     12000", rowsValid: 1},
		{name: "single spaces do not split vehicle names", kind: vocab.InputVehicleDistanceKm, raw: "BMW 3 Series AB123CD Car Medium Diesel 22100", rowsError: 1, code: CodeWrongColumnCount},
		{name: "comma rows rejected", kind: vocab.InputRefrigerantsAnnualKg, raw: "R410A,3,2", rowsError: 1, code: CodeWrongColumnCount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.kind, tt.raw)
			requireCounts(t, result, tt.rowsValid+tt.rowsError, tt.rowsValid, tt.rowsError)
			if tt.code != "" {
				requireError(t, result.Rows[0], tt.code)
			}
		})
	}
}

func TestParsePositiveNumber(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
		code  string
	}{
		{input: "3.2", ok: true},
		{input: "3,2", ok: true},
		{input: "3.2 kg", code: CodeUnitInNumber},
		{input: "-", code: CodeInvalidNumber},
		{input: "n/a", code: CodeInvalidNumber},
		{input: "-3.2", code: CodeNegativeValue},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, issues, ok := ParsePositiveNumber(tt.input)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if tt.code != "" && !hasIssue(issues, tt.code) {
				t.Fatalf("issues = %#v, want code %q", issues, tt.code)
			}
		})
	}
}

func requireCounts(t *testing.T, result ParseResult, rowsTotal, rowsValid, rowsError int) {
	t.Helper()
	if result.RowsTotal != rowsTotal || result.RowsValid != rowsValid || result.RowsError != rowsError {
		t.Fatalf("counts = total %d valid %d error %d, want total %d valid %d error %d", result.RowsTotal, result.RowsValid, result.RowsError, rowsTotal, rowsValid, rowsError)
	}
}

func requireNormalized(t *testing.T, row ParsedRow, key, want string) {
	t.Helper()
	got := row.Normalized[key]
	if got != want {
		t.Fatalf("normalized[%q] = %q, want %q; row = %#v", key, got, want, row)
	}
}

func requireError(t *testing.T, row ParsedRow, code string) {
	t.Helper()
	if !hasIssue(row.Errors, code) {
		t.Fatalf("errors = %#v, want code %q", row.Errors, code)
	}
}

func requireWarning(t *testing.T, row ParsedRow, code string) {
	t.Helper()
	if !hasIssue(row.Warnings, code) {
		t.Fatalf("warnings = %#v, want code %q", row.Warnings, code)
	}
}

func hasIssue(issues []ParseIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
