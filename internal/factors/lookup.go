package factors

import (
	"context"
	"errors"
	"fmt"

	"ghgo/internal/domain"
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

var ErrFactorNotFound = errors.New("factor not found")
var ErrAmbiguousFactor = errors.New("ambiguous factor")
var ErrUnsupportedActivity = errors.New("unsupported activity")

const ghgKgCO2e = "kgCO2e"

type Lookup struct {
	Store       *store.Store
	FactorSetID string
}

func NewLookup(st *store.Store, factorSetID string) *Lookup {
	return &Lookup{Store: st, FactorSetID: factorSetID}
}

func (l *Lookup) FindForActivityRecord(ctx context.Context, r domain.ActivityRecord) (*domain.EmissionFactor, error) {
	switch r.SourceKind {
	case domain.ActivitySourceKindElectricityMonthlyKWh:
		if err := validateCommon(r, domain.ActivityVectorElectricity, domain.Scope2, domain.ActivityMethodLocationBased, string(vocab.UnitKWh)); err != nil {
			return nil, err
		}
		return l.FindElectricityLocationFactor(ctx)
	case domain.ActivitySourceKindNaturalGasMonthlySMC:
		if err := validateCommon(r, domain.ActivityVectorNaturalGas, domain.Scope1, domain.ActivityMethodFuelBased, string(vocab.UnitSmc)); err != nil {
			return nil, err
		}
		if r.ActivityType != "natural_gas" {
			return nil, unsupported("natural gas activity_type must be natural_gas")
		}
		return l.FindNaturalGasFactor(ctx, r.Unit)
	case domain.ActivitySourceKindMobileFuelLitres:
		if err := validateCommon(r, domain.ActivityVectorMobileCombustion, domain.Scope1, domain.ActivityMethodFuelBased, string(vocab.UnitLitre)); err != nil {
			return nil, err
		}
		if r.FuelType == "" {
			return nil, unsupported("mobile fuel fuel_type is required")
		}
		if !supportedMobileFuel(r.FuelType) {
			return nil, unsupported("unsupported mobile fuel_type %q", r.FuelType)
		}
		activityType := r.ActivityType
		if activityType == "" {
			activityType = r.FuelType + "_mobile"
		}
		return l.findMobileFuelFactor(ctx, r.FuelType, r.Unit, activityType)
	case domain.ActivitySourceKindVehicleDistanceKM:
		if err := validateCommon(r, domain.ActivityVectorMobileCombustion, domain.Scope1, domain.ActivityMethodDistanceBased, string(vocab.UnitKm)); err != nil {
			return nil, err
		}
		return l.FindVehicleDistanceFactor(ctx, r.VehicleType, r.VehicleSizeClass, r.FuelType, r.Unit)
	case domain.ActivitySourceKindRefrigerantsAnnualKG:
		if err := validateCommon(r, domain.ActivityVectorRefrigerants, domain.Scope1, domain.ActivityMethodDirectGWP, string(vocab.UnitKg)); err != nil {
			return nil, err
		}
		return l.FindRefrigerantFactor(ctx, r.Substance, r.Unit)
	}

	return nil, unsupported("unsupported source_kind %q", r.SourceKind)
}

func (l *Lookup) FindElectricityLocationFactor(ctx context.Context) (*domain.EmissionFactor, error) {
	scope := int(domain.Scope2)
	activityType := "purchased_electricity"
	unit := string(vocab.UnitKWh)
	factorUnit := "kgCO2e/kWh"
	ghg := ghgKgCO2e
	return l.findExactlyOne(ctx, store.EmissionFactorQuery{
		FactorSetID:  l.FactorSetID,
		Scope:        &scope,
		ActivityType: &activityType,
		InputUnit:    &unit,
		FactorUnit:   &factorUnit,
		GHG:          &ghg,
	})
}

func (l *Lookup) FindNaturalGasFactor(ctx context.Context, unit string) (*domain.EmissionFactor, error) {
	if unit != string(vocab.UnitSmc) {
		return nil, unsupported("natural gas unit must be Smc")
	}
	scope := int(domain.Scope1)
	activityType := "natural_gas"
	factorUnit := "kgCO2e/Smc"
	ghg := ghgKgCO2e
	return l.findExactlyOne(ctx, store.EmissionFactorQuery{
		FactorSetID:  l.FactorSetID,
		Scope:        &scope,
		ActivityType: &activityType,
		InputUnit:    &unit,
		FactorUnit:   &factorUnit,
		GHG:          &ghg,
	})
}

func (l *Lookup) FindMobileFuelFactor(ctx context.Context, fuelType string, unit string) (*domain.EmissionFactor, error) {
	return l.findMobileFuelFactor(ctx, fuelType, unit, fuelType+"_mobile")
}

func (l *Lookup) findMobileFuelFactor(ctx context.Context, fuelType string, unit string, activityType string) (*domain.EmissionFactor, error) {
	if unit != string(vocab.UnitLitre) {
		return nil, unsupported("mobile fuel unit must be L")
	}
	if fuelType == "" {
		return nil, unsupported("mobile fuel fuel_type is required")
	}
	if !supportedMobileFuel(fuelType) {
		return nil, unsupported("unsupported mobile fuel_type %q", fuelType)
	}
	scope := int(domain.Scope1)
	factorUnit := "kgCO2e/L"
	ghg := ghgKgCO2e
	return l.findExactlyOne(ctx, store.EmissionFactorQuery{
		FactorSetID:  l.FactorSetID,
		Scope:        &scope,
		ActivityType: &activityType,
		FuelType:     &fuelType,
		InputUnit:    &unit,
		FactorUnit:   &factorUnit,
		GHG:          &ghg,
	})
}

func (l *Lookup) FindVehicleDistanceFactor(ctx context.Context, vehicleType string, sizeClass string, fuelType string, unit string) (*domain.EmissionFactor, error) {
	if unit != string(vocab.UnitKm) {
		return nil, unsupported("vehicle distance unit must be km")
	}
	if vehicleType == "" {
		return nil, unsupported("vehicle_type is required")
	}
	if sizeClass == "" {
		return nil, unsupported("vehicle_size_class is required")
	}
	if !supportedVehicleSize(vehicleType, sizeClass) {
		return nil, unsupported("unsupported vehicle type/size %q/%q", vehicleType, sizeClass)
	}
	if vehicleType != string(vocab.VehicleMotorbike) {
		if fuelType == "" {
			return nil, unsupported("vehicle fuel_type is required")
		}
		if !supportedVehicleFuel(fuelType) {
			return nil, unsupported("unsupported vehicle fuel_type %q", fuelType)
		}
		return l.findVehicleDistanceFactor(ctx, vehicleType, sizeClass, &fuelType, unit)
	}
	if fuelType != "" && !supportedVehicleFuel(fuelType) {
		return nil, unsupported("unsupported vehicle fuel_type %q", fuelType)
	}
	return l.findMotorbikeFactor(ctx, vehicleType, sizeClass, fuelType, unit)
}

func (l *Lookup) findVehicleDistanceFactor(ctx context.Context, vehicleType string, sizeClass string, fuelType *string, unit string) (*domain.EmissionFactor, error) {
	scope := int(domain.Scope1)
	activityType := "vehicle_distance"
	factorUnit := "kgCO2e/km"
	ghg := ghgKgCO2e
	return l.findExactlyOne(ctx, store.EmissionFactorQuery{
		FactorSetID:      l.FactorSetID,
		Scope:            &scope,
		ActivityType:     &activityType,
		FuelType:         fuelType,
		VehicleType:      &vehicleType,
		VehicleSizeClass: &sizeClass,
		InputUnit:        &unit,
		FactorUnit:       &factorUnit,
		GHG:              &ghg,
	})
}

func (l *Lookup) findMotorbikeFactor(ctx context.Context, vehicleType string, sizeClass string, fuelType string, unit string) (*domain.EmissionFactor, error) {
	scope := int(domain.Scope1)
	activityType := "vehicle_distance"
	factorUnit := "kgCO2e/km"
	ghg := ghgKgCO2e
	factors, err := l.find(ctx, store.EmissionFactorQuery{
		FactorSetID:      l.FactorSetID,
		Scope:            &scope,
		ActivityType:     &activityType,
		VehicleType:      &vehicleType,
		VehicleSizeClass: &sizeClass,
		InputUnit:        &unit,
		FactorUnit:       &factorUnit,
		GHG:              &ghg,
	})
	if err != nil {
		return nil, err
	}

	var generic []domain.EmissionFactor
	var fuelSpecific []domain.EmissionFactor
	for _, factor := range factors {
		if factor.FuelType == "" {
			generic = append(generic, factor)
		} else {
			fuelSpecific = append(fuelSpecific, factor)
		}
	}
	if len(fuelSpecific) == 0 {
		return exactOne(generic)
	}
	if fuelType == "" {
		return nil, unsupported("motorbike fuel_type is required for fuel-specific factors")
	}

	matching := []domain.EmissionFactor{}
	for _, factor := range fuelSpecific {
		if factor.FuelType == fuelType {
			matching = append(matching, factor)
		}
	}
	return exactOne(matching)
}

func (l *Lookup) FindRefrigerantFactor(ctx context.Context, substance string, unit string) (*domain.EmissionFactor, error) {
	if unit != string(vocab.UnitKg) {
		return nil, unsupported("refrigerant unit must be kg")
	}
	if substance == "" {
		return nil, unsupported("substance is required")
	}
	if !supportedRefrigerant(substance) {
		return nil, unsupported("unsupported refrigerant %q", substance)
	}
	scope := int(domain.Scope1)
	activityType := "refrigerant_leakage"
	factorUnit := "kgCO2e/kg"
	ghg := ghgKgCO2e
	return l.findExactlyOne(ctx, store.EmissionFactorQuery{
		FactorSetID:  l.FactorSetID,
		Scope:        &scope,
		ActivityType: &activityType,
		Substance:    &substance,
		InputUnit:    &unit,
		FactorUnit:   &factorUnit,
		GHG:          &ghg,
	})
}

func (l *Lookup) findExactlyOne(ctx context.Context, q store.EmissionFactorQuery) (*domain.EmissionFactor, error) {
	factors, err := l.find(ctx, q)
	if err != nil {
		return nil, err
	}
	return exactOne(factors)
}

func (l *Lookup) find(ctx context.Context, q store.EmissionFactorQuery) ([]domain.EmissionFactor, error) {
	if l == nil || l.Store == nil {
		return nil, fmt.Errorf("store is required")
	}
	if l.FactorSetID == "" {
		return nil, fmt.Errorf("factor set id is required")
	}
	return l.Store.FindEmissionFactors(ctx, q)
}

func exactOne(factors []domain.EmissionFactor) (*domain.EmissionFactor, error) {
	if len(factors) == 0 {
		return nil, ErrFactorNotFound
	}
	if len(factors) > 1 {
		return nil, ErrAmbiguousFactor
	}
	return &factors[0], nil
}

func validateCommon(r domain.ActivityRecord, vector domain.ActivityVector, scope domain.Scope, method domain.ActivityMethod, unit string) error {
	if r.Vector != vector {
		return unsupported("unsupported vector %q", r.Vector)
	}
	if r.Scope != scope {
		return unsupported("unsupported scope %d", r.Scope)
	}
	if r.Method != method {
		return unsupported("unsupported method %q", r.Method)
	}
	if r.Unit != unit {
		return unsupported("unsupported unit %q", r.Unit)
	}
	return nil
}

func supportedMobileFuel(fuelType string) bool {
	switch fuelType {
	case string(vocab.FuelDiesel), string(vocab.FuelPetrol), string(vocab.FuelLPG), string(vocab.FuelCNG):
		return true
	}
	return false
}

func supportedVehicleFuel(fuelType string) bool {
	switch fuelType {
	case string(vocab.FuelDiesel), string(vocab.FuelPetrol), string(vocab.FuelHybrid), string(vocab.FuelCNG), string(vocab.FuelLPG), string(vocab.FuelUnknown), string(vocab.FuelPHEV), string(vocab.FuelBEV):
		return true
	}
	return false
}

func supportedVehicleSize(vehicleType string, sizeClass string) bool {
	switch vehicleType {
	case string(vocab.VehicleCar), string(vocab.VehicleMotorbike):
		switch sizeClass {
		case string(vocab.SizeSmall), string(vocab.SizeMedium), string(vocab.SizeLarge), string(vocab.SizeAverage):
			return true
		}
	case string(vocab.VehicleVan):
		switch sizeClass {
		case string(vocab.SizeClassI), string(vocab.SizeClassII), string(vocab.SizeClassIII), string(vocab.SizeAverage):
			return true
		}
	}
	return false
}

func supportedRefrigerant(substance string) bool {
	switch substance {
	case string(vocab.RefrigerantR134a), string(vocab.RefrigerantR410A), string(vocab.RefrigerantR407C), string(vocab.RefrigerantR404A), string(vocab.RefrigerantR32), string(vocab.RefrigerantR22):
		return true
	}
	return false
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrUnsupportedActivity)...)
}
