package vocab

import "testing"

func TestParseMonth(t *testing.T) {
	tests := []struct {
		input string
		want  Month
		ok    bool
	}{
		{"January", MonthJanuary, true},
		{"jan", MonthJanuary, true},
		{"gennaio", MonthJanuary, true},
		{"gen", MonthJanuary, true},
		{"1", MonthJanuary, true},
		{"01", MonthJanuary, true},
		{"2025-01", MonthJanuary, true},
		{"2025/01", MonthJanuary, true},
		{"01/02/2025", 0, false},
		{"nonsense", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseMonth(tt.input)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("month = %v, want %v", got, tt.want)
			}
			if ok && got.EnglishName() == "" {
				t.Fatalf("EnglishName() is empty for %v", got)
			}
		})
	}
}

func TestNormalizeUnit(t *testing.T) {
	tests := []struct {
		input string
		want  Unit
		ok    bool
	}{
		{"kwh", UnitKWh, true},
		{"smc", UnitSmc, true},
		{"litre", UnitLitre, true},
		{"km", UnitKm, true},
		{"kg", UnitKg, true},
		{"m3", "", false},
		{"MWh", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := NormalizeUnit(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizeUnit(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNormalizeFuelType(t *testing.T) {
	tests := []struct {
		input string
		want  FuelType
		ok    bool
	}{
		{"gasolio", FuelDiesel, true},
		{"benzina", FuelPetrol, true},
		{"gpl", FuelLPG, true},
		{"metano", FuelCNG, true},
		{"plug-in", FuelPHEV, true},
		{"elettrica", FuelBEV, true},
		{"unknown", FuelUnknown, true},
		{"hydrogen", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := NormalizeFuelType(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizeFuelType(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNormalizeVehicleType(t *testing.T) {
	tests := []struct {
		input string
		want  VehicleType
		ok    bool
	}{
		{"auto", VehicleCar, true},
		{"furgone", VehicleVan, true},
		{"moto", VehicleMotorbike, true},
		{"lorry", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := NormalizeVehicleType(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizeVehicleType(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNormalizeVehicleSizeClass(t *testing.T) {
	tests := []struct {
		vehicleType VehicleType
		input       string
		want        VehicleSizeClass
		ok          bool
	}{
		{VehicleCar, "small car", SizeSmall, true},
		{VehicleVan, "classe 2", SizeClassII, true},
		{VehicleVan, "medium", "", false},
		{VehicleCar, "class_i", "", false},
		{VehicleMotorbike, "media", SizeMedium, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.vehicleType)+"/"+tt.input, func(t *testing.T) {
			got, ok := NormalizeVehicleSizeClass(tt.vehicleType, tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizeVehicleSizeClass(%q, %q) = %q, %v; want %q, %v", tt.vehicleType, tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestVehicleSizeClassCompatible(t *testing.T) {
	tests := []struct {
		vehicleType VehicleType
		size        VehicleSizeClass
		want        bool
	}{
		{VehicleCar, SizeSmall, true},
		{VehicleCar, SizeClassI, false},
		{VehicleVan, SizeClassIII, true},
		{VehicleVan, SizeMedium, false},
		{VehicleMotorbike, SizeLarge, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.vehicleType)+"/"+string(tt.size), func(t *testing.T) {
			if got := VehicleSizeClassCompatible(tt.vehicleType, tt.size); got != tt.want {
				t.Fatalf("VehicleSizeClassCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeRefrigerant(t *testing.T) {
	tests := []struct {
		input string
		want  Refrigerant
		ok    bool
	}{
		{"r-410a", RefrigerantR410A, true},
		{"410a", RefrigerantR410A, true},
		{"r134a", RefrigerantR134a, true},
		{"co2", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := NormalizeRefrigerant(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizeRefrigerant(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestInputKindValid(t *testing.T) {
	valid := []InputKind{
		InputElectricityMonthlyKWh,
		InputNaturalGasMonthlySmc,
		InputMobileFuelLitres,
		InputVehicleDistanceKm,
		InputRefrigerantsAnnualKg,
	}
	for _, input := range valid {
		t.Run(string(input), func(t *testing.T) {
			if !input.Valid() {
				t.Fatalf("%q should be valid", input)
			}
		})
	}
	if InputKind("invalid").Valid() {
		t.Fatal("invalid input kind should be invalid")
	}
}

func TestMethodValid(t *testing.T) {
	valid := []Method{
		MethodLocationBased,
		MethodMarketBased,
		MethodFuelBased,
		MethodDistanceBased,
		MethodDirectGWP,
	}
	for _, method := range valid {
		t.Run(string(method), func(t *testing.T) {
			if !method.Valid() {
				t.Fatalf("%q should be valid", method)
			}
			if method.EnglishLabel() == "" {
				t.Fatalf("EnglishLabel() is empty for %q", method)
			}
		})
	}
	if Method("invalid").Valid() {
		t.Fatal("invalid method should be invalid")
	}
}

func TestVectorValid(t *testing.T) {
	valid := []Vector{
		VectorElectricity,
		VectorNaturalGas,
		VectorMobileCombustion,
		VectorRefrigerants,
	}
	for _, vector := range valid {
		t.Run(string(vector), func(t *testing.T) {
			if !vector.Valid() {
				t.Fatalf("%q should be valid", vector)
			}
			if vector.EnglishLabel() == "" {
				t.Fatalf("EnglishLabel() is empty for %q", vector)
			}
		})
	}
	if Vector("invalid").Valid() {
		t.Fatal("invalid vector should be invalid")
	}
}
