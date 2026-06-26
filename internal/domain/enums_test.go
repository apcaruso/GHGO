package domain

import "testing"

type validEnum interface {
	Valid() bool
}

func TestEnumValidation(t *testing.T) {
	tests := []struct {
		name  string
		value validEnum
		want  bool
	}{
		{"mobile method fuel based", MobileMethodFuelBased, true},
		{"mobile method distance based", MobileMethodDistanceBased, true},
		{"mobile method invalid", MobileMethod("spend_based"), false},

		{"go coverage none", GOCoverageNone, true},
		{"go coverage full", GOCoverageFull, true},
		{"go coverage invalid", GOCoverage("partial"), false},

		{"reporting period status draft", ReportingPeriodStatusDraft, true},
		{"reporting period status locked", ReportingPeriodStatusLocked, true},
		{"reporting period status archived", ReportingPeriodStatusArchived, true},
		{"reporting period status invalid", ReportingPeriodStatus("closed"), false},

		{"activity source electricity", ActivitySourceKindElectricityMonthlyKWh, true},
		{"activity source natural gas", ActivitySourceKindNaturalGasMonthlySMC, true},
		{"activity source mobile fuel", ActivitySourceKindMobileFuelLitres, true},
		{"activity source vehicle distance", ActivitySourceKindVehicleDistanceKM, true},
		{"activity source refrigerants", ActivitySourceKindRefrigerantsAnnualKG, true},
		{"activity source invalid", ActivitySourceKind("commuting"), false},

		{"scope 1", Scope1, true},
		{"scope 2", Scope2, true},
		{"scope invalid", Scope(3), false},

		{"activity vector electricity", ActivityVectorElectricity, true},
		{"activity vector natural gas", ActivityVectorNaturalGas, true},
		{"activity vector mobile combustion", ActivityVectorMobileCombustion, true},
		{"activity vector refrigerants", ActivityVectorRefrigerants, true},
		{"activity vector invalid", ActivityVector("waste"), false},

		{"activity method location based", ActivityMethodLocationBased, true},
		{"activity method market based", ActivityMethodMarketBased, true},
		{"activity method fuel based", ActivityMethodFuelBased, true},
		{"activity method distance based", ActivityMethodDistanceBased, true},
		{"activity method direct gwp", ActivityMethodDirectGWP, true},
		{"activity method invalid", ActivityMethod("average_data"), false},

		{"activity status active", ActivityRecordStatusActive, true},
		{"activity status superseded", ActivityRecordStatusSuperseded, true},
		{"activity status deleted", ActivityRecordStatusDeleted, true},
		{"activity status invalid", ActivityRecordStatus("draft"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
