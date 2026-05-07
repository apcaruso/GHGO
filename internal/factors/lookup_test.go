package factors

import (
	"context"
	"errors"
	"testing"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/store"
)

func TestLookupElectricityLocationFactor(t *testing.T) {
	st := newTestStore(t)
	factorSetID := createLookupFactorSet(t, st)
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:           "electricity-factor",
		FactorSetID:  factorSetID,
		Scope:        domain.Scope2,
		ActivityType: "purchased_electricity",
		InputUnit:    "kWh",
		FactorUnit:   "kgCO2e/kWh",
		GHG:          "kgCO2e",
		FactorValue:  0.2,
	})

	got, err := NewLookup(store.NewRepository(st), factorSetID).FindForActivityRecord(context.Background(), electricityRecord())
	if err != nil {
		t.Fatalf("lookup electricity factor: %v", err)
	}
	if got.ID != "electricity-factor" {
		t.Fatalf("factor id = %q, want electricity-factor", got.ID)
	}
}

func TestLookupNaturalGasDirectSmcOnly(t *testing.T) {
	t.Run("Smc factor succeeds", func(t *testing.T) {
		st := newTestStore(t)
		factorSetID := createLookupFactorSet(t, st)
		insertLookupFactor(t, st, domain.EmissionFactor{
			ID:           "natural-gas-smc",
			FactorSetID:  factorSetID,
			Scope:        domain.Scope1,
			ActivityType: "natural_gas",
			InputUnit:    "Smc",
			FactorUnit:   "kgCO2e/Smc",
			GHG:          "kgCO2e",
			FactorValue:  2.0,
		})

		got, err := NewLookup(store.NewRepository(st), factorSetID).FindForActivityRecord(context.Background(), naturalGasRecord())
		if err != nil {
			t.Fatalf("lookup natural gas factor: %v", err)
		}
		if got.ID != "natural-gas-smc" {
			t.Fatalf("factor id = %q, want natural-gas-smc", got.ID)
		}
	})

	t.Run("kWh factor does not satisfy Smc lookup", func(t *testing.T) {
		st := newTestStore(t)
		factorSetID := createLookupFactorSet(t, st)
		insertLookupFactor(t, st, domain.EmissionFactor{
			ID:           "natural-gas-kwh",
			FactorSetID:  factorSetID,
			Scope:        domain.Scope1,
			ActivityType: "natural_gas",
			InputUnit:    "kWh",
			FactorUnit:   "kgCO2e/kWh",
			GHG:          "kgCO2e",
			FactorValue:  0.18,
		})

		_, err := NewLookup(store.NewRepository(st), factorSetID).FindForActivityRecord(context.Background(), naturalGasRecord())
		expectErrorIs(t, err, ErrFactorNotFound)
	})
}

func TestLookupMobileFuelFactor(t *testing.T) {
	st := newTestStore(t)
	factorSetID := createLookupFactorSet(t, st)
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:           "diesel-litres",
		FactorSetID:  factorSetID,
		Scope:        domain.Scope1,
		ActivityType: "diesel_mobile",
		FuelType:     "diesel",
		InputUnit:    "L",
		FactorUnit:   "kgCO2e/L",
		GHG:          "kgCO2e",
		FactorValue:  2.51,
	})

	lookup := NewLookup(store.NewRepository(st), factorSetID)
	got, err := lookup.FindForActivityRecord(context.Background(), mobileFuelRecord("diesel"))
	if err != nil {
		t.Fatalf("lookup diesel factor: %v", err)
	}
	if got.ID != "diesel-litres" {
		t.Fatalf("factor id = %q, want diesel-litres", got.ID)
	}

	_, err = lookup.FindForActivityRecord(context.Background(), mobileFuelRecord(""))
	expectErrorIs(t, err, ErrUnsupportedActivity)

	_, err = lookup.FindForActivityRecord(context.Background(), mobileFuelRecord("bev"))
	expectErrorIs(t, err, ErrUnsupportedActivity)
}

func TestLookupVehicleDistanceFactor(t *testing.T) {
	st := newTestStore(t)
	factorSetID := createLookupFactorSet(t, st)
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:               "car-small-petrol",
		FactorSetID:      factorSetID,
		Scope:            domain.Scope1,
		ActivityType:     "vehicle_distance",
		FuelType:         "petrol",
		VehicleType:      "car",
		VehicleSizeClass: "small",
		InputUnit:        "km",
		FactorUnit:       "kgCO2e/km",
		GHG:              "kgCO2e",
		FactorValue:      0.15,
	})
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:               "van-class-ii-diesel",
		FactorSetID:      factorSetID,
		Scope:            domain.Scope1,
		ActivityType:     "vehicle_distance",
		FuelType:         "diesel",
		VehicleType:      "van",
		VehicleSizeClass: "class_ii",
		InputUnit:        "km",
		FactorUnit:       "kgCO2e/km",
		GHG:              "kgCO2e",
		FactorValue:      0.25,
	})

	lookup := NewLookup(store.NewRepository(st), factorSetID)
	got, err := lookup.FindForActivityRecord(context.Background(), vehicleDistanceRecord("car", "small", "petrol"))
	if err != nil {
		t.Fatalf("lookup car factor: %v", err)
	}
	if got.ID != "car-small-petrol" {
		t.Fatalf("factor id = %q, want car-small-petrol", got.ID)
	}

	_, err = lookup.FindForActivityRecord(context.Background(), vehicleDistanceRecord("car", "small", "diesel"))
	expectErrorIs(t, err, ErrFactorNotFound)

	got, err = lookup.FindForActivityRecord(context.Background(), vehicleDistanceRecord("van", "class_ii", "diesel"))
	if err != nil {
		t.Fatalf("lookup van factor: %v", err)
	}
	if got.ID != "van-class-ii-diesel" {
		t.Fatalf("factor id = %q, want van-class-ii-diesel", got.ID)
	}

	_, err = lookup.FindForActivityRecord(context.Background(), vehicleDistanceRecord("car", "class_i", "petrol"))
	expectErrorIs(t, err, ErrUnsupportedActivity)

	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:               "car-small-petrol-duplicate",
		FactorSetID:      factorSetID,
		Scope:            domain.Scope1,
		ActivityType:     "vehicle_distance",
		FuelType:         "petrol",
		VehicleType:      "car",
		VehicleSizeClass: "small",
		InputUnit:        "km",
		FactorUnit:       "kgCO2e/km",
		GHG:              "kgCO2e",
		FactorValue:      0.151,
	})
	_, err = lookup.FindForActivityRecord(context.Background(), vehicleDistanceRecord("car", "small", "petrol"))
	expectErrorIs(t, err, ErrAmbiguousFactor)
}

func TestLookupMotorbikeGenericFactor(t *testing.T) {
	st := newTestStore(t)
	factorSetID := createLookupFactorSet(t, st)
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:               "motorbike-average",
		FactorSetID:      factorSetID,
		Scope:            domain.Scope1,
		ActivityType:     "vehicle_distance",
		VehicleType:      "motorbike",
		VehicleSizeClass: "average",
		InputUnit:        "km",
		FactorUnit:       "kgCO2e/km",
		GHG:              "kgCO2e",
		FactorValue:      0.10,
	})

	got, err := NewLookup(store.NewRepository(st), factorSetID).FindForActivityRecord(context.Background(), vehicleDistanceRecord("motorbike", "average", ""))
	if err != nil {
		t.Fatalf("lookup motorbike factor: %v", err)
	}
	if got.ID != "motorbike-average" {
		t.Fatalf("factor id = %q, want motorbike-average", got.ID)
	}
}

func TestLookupRefrigerantFactor(t *testing.T) {
	st := newTestStore(t)
	factorSetID := createLookupFactorSet(t, st)
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:           "r410a-factor",
		FactorSetID:  factorSetID,
		Scope:        domain.Scope1,
		ActivityType: "refrigerant_leakage",
		Substance:    "R410A",
		InputUnit:    "kg",
		FactorUnit:   "kgCO2e/kg",
		GHG:          "kgCO2e",
		FactorValue:  2088,
	})

	lookup := NewLookup(store.NewRepository(st), factorSetID)
	got, err := lookup.FindForActivityRecord(context.Background(), refrigerantRecord("R410A"))
	if err != nil {
		t.Fatalf("lookup R410A factor: %v", err)
	}
	if got.ID != "r410a-factor" {
		t.Fatalf("factor id = %q, want r410a-factor", got.ID)
	}

	_, err = lookup.FindForActivityRecord(context.Background(), refrigerantRecord("R134a"))
	expectErrorIs(t, err, ErrFactorNotFound)
}

func TestLookupUnsupportedActivity(t *testing.T) {
	st := newTestStore(t)
	factorSetID := createLookupFactorSet(t, st)
	insertLookupFactor(t, st, domain.EmissionFactor{
		ID:           "electricity-factor",
		FactorSetID:  factorSetID,
		Scope:        domain.Scope2,
		ActivityType: "purchased_electricity",
		InputUnit:    "kWh",
		FactorUnit:   "kgCO2e/kWh",
		GHG:          "kgCO2e",
		FactorValue:  0.2,
	})
	lookup := NewLookup(store.NewRepository(st), factorSetID)

	unknown := electricityRecord()
	unknown.SourceKind = domain.ActivitySourceKind("unknown")
	_, err := lookup.FindForActivityRecord(context.Background(), unknown)
	expectErrorIs(t, err, ErrUnsupportedActivity)

	wrongUnit := electricityRecord()
	wrongUnit.Unit = "MWh"
	_, err = lookup.FindForActivityRecord(context.Background(), wrongUnit)
	expectErrorIs(t, err, ErrUnsupportedActivity)
}

func createLookupFactorSet(t *testing.T, st *store.Store) domain.ID {
	t.Helper()
	factorSet := domain.FactorSet{
		ID:           "lookup-factor-set",
		Name:         "Lookup Factors",
		Source:       "test",
		Year:         2025,
		Version:      "lookup",
		ImportedAt:   time.Now().UTC(),
		MetadataJSON: `{}`,
	}
	if err := st.CreateFactorSet(factorSet); err != nil {
		t.Fatalf("create factor set: %v", err)
	}
	return factorSet.ID
}

func insertLookupFactor(t *testing.T, st *store.Store, factor domain.EmissionFactor) {
	t.Helper()
	if factor.Source == "" {
		factor.Source = "test"
	}
	if factor.MetadataJSON == "" {
		factor.MetadataJSON = `{}`
	}
	if err := st.CreateEmissionFactor(factor); err != nil {
		t.Fatalf("create emission factor %q: %v", factor.ID, err)
	}
}

func electricityRecord() domain.ActivityRecord {
	return domain.ActivityRecord{
		SourceKind:   domain.ActivitySourceKindElectricityMonthlyKWh,
		Scope:        domain.Scope2,
		Vector:       domain.ActivityVectorElectricity,
		Method:       domain.ActivityMethodLocationBased,
		ActivityType: "purchased_electricity",
		Unit:         "kWh",
	}
}

func naturalGasRecord() domain.ActivityRecord {
	return domain.ActivityRecord{
		SourceKind:   domain.ActivitySourceKindNaturalGasMonthlySMC,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorNaturalGas,
		Method:       domain.ActivityMethodFuelBased,
		ActivityType: "natural_gas",
		Unit:         "Smc",
	}
}

func mobileFuelRecord(fuelType string) domain.ActivityRecord {
	activityType := ""
	if fuelType != "" {
		activityType = fuelType + "_mobile"
	}
	return domain.ActivityRecord{
		SourceKind:   domain.ActivitySourceKindMobileFuelLitres,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorMobileCombustion,
		Method:       domain.ActivityMethodFuelBased,
		ActivityType: activityType,
		FuelType:     fuelType,
		Unit:         "L",
	}
}

func vehicleDistanceRecord(vehicleType string, sizeClass string, fuelType string) domain.ActivityRecord {
	return domain.ActivityRecord{
		SourceKind:       domain.ActivitySourceKindVehicleDistanceKM,
		Scope:            domain.Scope1,
		Vector:           domain.ActivityVectorMobileCombustion,
		Method:           domain.ActivityMethodDistanceBased,
		ActivityType:     "vehicle_distance",
		VehicleType:      vehicleType,
		VehicleSizeClass: sizeClass,
		FuelType:         fuelType,
		Unit:             "km",
	}
}

func refrigerantRecord(substance string) domain.ActivityRecord {
	return domain.ActivityRecord{
		SourceKind:   domain.ActivitySourceKindRefrigerantsAnnualKG,
		Scope:        domain.Scope1,
		Vector:       domain.ActivityVectorRefrigerants,
		Method:       domain.ActivityMethodDirectGWP,
		ActivityType: "refrigerant_leakage",
		Substance:    substance,
		Unit:         "kg",
	}
}

func expectErrorIs(t *testing.T, err error, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
