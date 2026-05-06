package factors

import "testing"

func TestIsKgCO2eRejectsGasComponentRows(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{value: "kg CO2e", want: true},
		{value: "kg CO2e per unit", want: true},
		{value: "kg CO2e of CO2 per unit", want: false},
		{value: "kg CO2e of CH4 per unit", want: false},
		{value: "kg CO2e of N2O per unit", want: false},
	}

	for _, tc := range cases {
		if got := isKgCO2e(tc.value); got != tc.want {
			t.Fatalf("isKgCO2e(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
