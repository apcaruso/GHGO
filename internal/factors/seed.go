package factors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/store"
)

type SeedFactor struct {
	Scope int

	Source string

	Level1     string
	Level2     string
	Level3     string
	Level4     string
	ColumnText string

	ActivityType     string
	FuelType         string
	VehicleType      string
	VehicleSizeClass string
	Substance        string

	InputUnit   string
	FactorUnit  string
	GHG         string
	FactorValue float64
}

func EnsureDefaultFactors(ctx context.Context, st *store.Store) (*domain.FactorSet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if st == nil {
		return nil, fmt.Errorf("store is required")
	}

	var factorSet *domain.FactorSet
	if err := st.WithTx(ctx, func(tx *store.Store) error {
		set, err := tx.FindFactorSetBySourceYearVersion(defra2025Source, defra2025Year, defra2025Version)
		if errors.Is(err, store.ErrNotFound) {
			set = &domain.FactorSet{
				ID:           defra2025FactorSetID,
				Name:         defra2025Name,
				Source:       defra2025Source,
				Year:         defra2025Year,
				Version:      defra2025Version,
				ImportedAt:   time.Now().UTC(),
				MetadataJSON: `{"seeded":true}`,
			}
			if err := tx.CreateFactorSet(*set); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		for _, seed := range defra2025SeedFactors() {
			matches, err := tx.FindEmissionFactors(ctx, seed.query(set.ID))
			if err != nil {
				return err
			}
			if len(matches) > 1 {
				return fmt.Errorf("default factor %q has %d matching rows", seed.key(), len(matches))
			}
			if len(matches) == 1 {
				continue
			}

			if err := tx.CreateEmissionFactor(seed.emissionFactor(set.ID)); err != nil {
				return err
			}
		}

		factorSet = set
		return nil
	}); err != nil {
		return nil, err
	}

	return factorSet, nil
}

func (f SeedFactor) query(factorSetID domain.ID) store.EmissionFactorQuery {
	scope := f.Scope
	activityType := f.ActivityType
	inputUnit := f.InputUnit
	factorUnit := f.FactorUnit
	ghg := f.GHG
	q := store.EmissionFactorQuery{
		FactorSetID:  factorSetID,
		Scope:        &scope,
		ActivityType: &activityType,
		InputUnit:    &inputUnit,
		FactorUnit:   &factorUnit,
		GHG:          &ghg,
	}
	if f.FuelType != "" {
		q.FuelType = &f.FuelType
	}
	if f.VehicleType != "" {
		q.VehicleType = &f.VehicleType
	}
	if f.VehicleSizeClass != "" {
		q.VehicleSizeClass = &f.VehicleSizeClass
	}
	if f.Substance != "" {
		q.Substance = &f.Substance
	}
	return q
}

func (f SeedFactor) emissionFactor(factorSetID domain.ID) domain.EmissionFactor {
	return domain.EmissionFactor{
		ID:               f.id(),
		FactorSetID:      factorSetID,
		Source:           f.Source,
		Scope:            domain.Scope(f.Scope),
		Level1:           f.Level1,
		Level2:           f.Level2,
		Level3:           f.Level3,
		Level4:           f.Level4,
		ColumnText:       f.ColumnText,
		ActivityType:     f.ActivityType,
		FuelType:         f.FuelType,
		VehicleType:      f.VehicleType,
		VehicleSizeClass: f.VehicleSizeClass,
		Substance:        f.Substance,
		InputUnit:        f.InputUnit,
		FactorUnit:       f.FactorUnit,
		GHG:              f.GHG,
		FactorValue:      f.FactorValue,
		MetadataJSON:     `{"seeded":true}`,
	}
}

func (f SeedFactor) id() domain.ID {
	return domain.ID("emission_factor_defra_2025_" + f.key())
}

func (f SeedFactor) key() string {
	parts := []string{f.ActivityType, f.FuelType, f.VehicleType, f.VehicleSizeClass, f.Substance, f.InputUnit}
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		part = sanitizeSeedIDPart(part)
		if part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, "_")
}

func sanitizeSeedIDPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("/", "_", "-", "_", " ", "_", ".", "_")
	value = replacer.Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}
