package factors

import "testing"

func TestMapDEFRA2025RowSkipsAmbiguousImportedRows(t *testing.T) {
	cases := []struct {
		name string
		row  defraRow
	}{
		{
			name: "bio diesel",
			row:  defraRow{Scope: 1, Level1: "Bioenergy", Level2: "Biofuel", Level3: "Biodiesel HVO", UOM: "litres", GHGUnit: "kg CO2e", ConversionFactor: 0.03},
		},
		{
			name: "mineral diesel",
			row:  defraRow{Scope: 1, Level1: "Fuels", Level2: "Liquid fuels", Level3: "Diesel (100% mineral diesel)", UOM: "litres", GHGUnit: "kg CO2e", ConversionFactor: 2.66},
		},
		{
			name: "petroleum gas",
			row:  defraRow{Scope: 1, Level1: "Fuels", Level2: "Gaseous fuels", Level3: "Other petroleum gas", UOM: "litres", GHGUnit: "kg CO2e", ConversionFactor: 0.94},
		},
		{
			name: "car market segment",
			row:  defraRow{Scope: 1, Level1: "Passenger vehicles", Level2: "Cars (by market segment)", Level3: "Upper medium", ColumnText: "Petrol", UOM: "km", GHGUnit: "kg CO2e", ConversionFactor: 0.18},
		},
		{
			name: "refrigerant partial total",
			row:  defraRow{Scope: 1, Level1: "Refrigerant & other", Level2: "Blends", Level3: "R410A", ColumnText: "Emissions including only Kyoto products", UOM: "kg", GHGUnit: "kg CO2e", ConversionFactor: 1924},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if candidate, ok := mapDEFRA2025Row(tc.row); ok {
				t.Fatalf("mapDEFRA2025Row() = %#v, want skipped", candidate)
			}
		})
	}
}

func TestMapDEFRA2025RowKeepsCanonicalRows(t *testing.T) {
	cases := []struct {
		name         string
		row          defraRow
		activityType string
	}{
		{
			name:         "diesel average blend",
			row:          defraRow{Scope: 1, Level1: "Fuels", Level2: "Liquid fuels", Level3: "Diesel (average biofuel blend)", UOM: "litres", GHGUnit: "kg CO2e", ConversionFactor: 2.57},
			activityType: "diesel_mobile",
		},
		{
			name:         "car by size",
			row:          defraRow{Scope: 1, Level1: "Passenger vehicles", Level2: "Cars (by size)", Level3: "Medium car", ColumnText: "Petrol", UOM: "km", GHGUnit: "kg CO2e", ConversionFactor: 0.17},
			activityType: "vehicle_distance",
		},
		{
			name:         "refrigerant total",
			row:          defraRow{Scope: 1, Level1: "Refrigerant & other", Level2: "Blends", Level3: "R410A", ColumnText: "Total emissions including non-Kyoto products", UOM: "kg", GHGUnit: "kg CO2e", ConversionFactor: 1924},
			activityType: "refrigerant_leakage",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate, ok := mapDEFRA2025Row(tc.row)
			if !ok {
				t.Fatalf("mapDEFRA2025Row() skipped canonical row")
			}
			if candidate.ActivityType != tc.activityType {
				t.Fatalf("activity type = %q, want %q", candidate.ActivityType, tc.activityType)
			}
		})
	}
}
