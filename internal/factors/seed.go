package factors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"ghgo/factorpacks"
	"ghgo/internal/domain"
	"ghgo/internal/store"
)

const defaultFactorPackPath = "defra-2025/factor-pack.json"

type FactorPack struct {
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Source   string          `json:"source"`
	Year     int             `json:"year"`
	Version  string          `json:"version"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Rows     []FactorPackRow `json:"normalized_rows"`
}

type FactorPackRow struct {
	ID               string          `json:"id,omitempty"`
	Source           string          `json:"source,omitempty"`
	Scope            int             `json:"scope"`
	Level1           string          `json:"level_1,omitempty"`
	Level2           string          `json:"level_2,omitempty"`
	Level3           string          `json:"level_3,omitempty"`
	Level4           string          `json:"level_4,omitempty"`
	ColumnText       string          `json:"column_text,omitempty"`
	ActivityType     string          `json:"activity_type"`
	FuelType         string          `json:"fuel_type,omitempty"`
	VehicleType      string          `json:"vehicle_type,omitempty"`
	VehicleSizeClass string          `json:"vehicle_size_class,omitempty"`
	Substance        string          `json:"substance,omitempty"`
	InputUnit        string          `json:"input_unit"`
	FactorUnit       string          `json:"factor_unit"`
	GHG              string          `json:"ghg"`
	FactorValue      float64         `json:"factor_value"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
}

func LoadFactorPack(fsys fs.FS, path string) (*FactorPack, error) {
	if fsys == nil {
		return nil, fmt.Errorf("factor pack filesystem is required")
	}
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("factor pack path is required")
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read factor pack %q: %w", path, err)
	}

	var pack FactorPack
	if err := json.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("decode factor pack %q: %w", path, err)
	}
	pack = pack.normalized()
	if err := pack.validate(); err != nil {
		return nil, fmt.Errorf("validate factor pack %q: %w", path, err)
	}
	return &pack, nil
}

func EnsureDefaultFactors(ctx context.Context, st *store.Store) (*domain.FactorSet, error) {
	pack, err := LoadFactorPack(factorpacks.FS, defaultFactorPackPath)
	if err != nil {
		return nil, err
	}
	return EnsureFactorPack(ctx, st, *pack)
}

func EnsureFactorPack(ctx context.Context, st *store.Store, pack FactorPack) (*domain.FactorSet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if st == nil {
		return nil, fmt.Errorf("store is required")
	}
	pack = pack.normalized()
	if err := pack.validate(); err != nil {
		return nil, err
	}

	var factorSet *domain.FactorSet
	if err := st.WithTx(ctx, func(tx *store.Store) error {
		set, err := tx.FindFactorSetBySourceYearVersion(pack.Source, pack.Year, pack.Version)
		if errors.Is(err, store.ErrNotFound) {
			metadataJSON, err := compactMetadata(pack.Metadata)
			if err != nil {
				return err
			}
			set = &domain.FactorSet{
				ID:           pack.factorSetID(),
				Name:         pack.factorSetName(),
				Source:       pack.Source,
				Year:         pack.Year,
				Version:      pack.Version,
				ImportedAt:   time.Now().UTC(),
				MetadataJSON: metadataJSON,
			}
			if err := tx.CreateFactorSet(*set); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		for _, row := range pack.Rows {
			matches, err := tx.FindEmissionFactors(ctx, row.query(set.ID))
			if err != nil {
				return err
			}
			if len(matches) > 1 {
				return fmt.Errorf("factor pack row %q has %d matching rows", row.key(), len(matches))
			}
			if len(matches) == 1 {
				continue
			}

			factor, err := row.emissionFactor(pack, set.ID)
			if err != nil {
				return err
			}
			if err := tx.CreateEmissionFactor(factor); err != nil {
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

func (p FactorPack) normalized() FactorPack {
	p.ID = strings.TrimSpace(p.ID)
	p.Name = strings.TrimSpace(p.Name)
	p.Source = strings.TrimSpace(p.Source)
	p.Version = strings.TrimSpace(p.Version)
	for i := range p.Rows {
		p.Rows[i] = p.Rows[i].normalized()
	}
	return p
}

func (p FactorPack) validate() error {
	if p.Source == "" {
		return fmt.Errorf("source is required")
	}
	if p.Year <= 0 {
		return fmt.Errorf("year is required")
	}
	if p.Version == "" {
		return fmt.Errorf("version is required")
	}
	if _, err := compactMetadata(p.Metadata); err != nil {
		return fmt.Errorf("metadata: %w", err)
	}
	if len(p.Rows) == 0 {
		return fmt.Errorf("rows are required")
	}

	ids := map[string]struct{}{}
	keys := map[string]struct{}{}
	for i, row := range p.Rows {
		if err := row.validate(i); err != nil {
			return err
		}
		id := string(row.id(p))
		if _, ok := ids[id]; ok {
			return fmt.Errorf("row %d duplicates id %q", i+1, id)
		}
		ids[id] = struct{}{}

		key := row.key()
		if _, ok := keys[key]; ok {
			return fmt.Errorf("row %d duplicates normalized key %q", i+1, key)
		}
		keys[key] = struct{}{}
	}

	return nil
}

func (r FactorPackRow) normalized() FactorPackRow {
	r.ID = strings.TrimSpace(r.ID)
	r.Source = strings.TrimSpace(r.Source)
	r.Level1 = strings.TrimSpace(r.Level1)
	r.Level2 = strings.TrimSpace(r.Level2)
	r.Level3 = strings.TrimSpace(r.Level3)
	r.Level4 = strings.TrimSpace(r.Level4)
	r.ColumnText = strings.TrimSpace(r.ColumnText)
	r.ActivityType = strings.TrimSpace(r.ActivityType)
	r.FuelType = strings.TrimSpace(r.FuelType)
	r.VehicleType = strings.TrimSpace(r.VehicleType)
	r.VehicleSizeClass = strings.TrimSpace(r.VehicleSizeClass)
	r.Substance = strings.TrimSpace(r.Substance)
	r.InputUnit = strings.TrimSpace(r.InputUnit)
	r.FactorUnit = strings.TrimSpace(r.FactorUnit)
	r.GHG = strings.TrimSpace(r.GHG)
	return r
}

func (r FactorPackRow) validate(index int) error {
	row := index + 1
	if r.Scope <= 0 {
		return fmt.Errorf("row %d scope is required", row)
	}
	if r.ActivityType == "" {
		return fmt.Errorf("row %d activity_type is required", row)
	}
	if r.InputUnit == "" {
		return fmt.Errorf("row %d input_unit is required", row)
	}
	if r.FactorUnit == "" {
		return fmt.Errorf("row %d factor_unit is required", row)
	}
	if r.GHG == "" {
		return fmt.Errorf("row %d ghg is required", row)
	}
	if _, err := compactMetadata(r.Metadata); err != nil {
		return fmt.Errorf("row %d metadata: %w", row, err)
	}
	return nil
}

func (p FactorPack) factorSetID() domain.ID {
	if strings.TrimSpace(p.ID) != "" {
		return domain.ID(p.ID)
	}
	return domain.ID("factor_set_" + sanitizeIDPart(p.Source) + "_" + strconv.Itoa(p.Year) + "_" + sanitizeIDPart(p.Version))
}

func (p FactorPack) factorSetName() string {
	if strings.TrimSpace(p.Name) != "" {
		return p.Name
	}
	return strings.TrimSpace(fmt.Sprintf("%s %d", p.Source, p.Year))
}

func (r FactorPackRow) query(factorSetID domain.ID) store.EmissionFactorQuery {
	scope := r.Scope
	activityType := r.ActivityType
	fuelType := r.FuelType
	vehicleType := r.VehicleType
	vehicleSizeClass := r.VehicleSizeClass
	substance := r.Substance
	inputUnit := r.InputUnit
	factorUnit := r.FactorUnit
	ghg := r.GHG
	return store.EmissionFactorQuery{
		FactorSetID:      factorSetID,
		Scope:            &scope,
		ActivityType:     &activityType,
		FuelType:         &fuelType,
		VehicleType:      &vehicleType,
		VehicleSizeClass: &vehicleSizeClass,
		Substance:        &substance,
		InputUnit:        &inputUnit,
		FactorUnit:       &factorUnit,
		GHG:              &ghg,
	}
}

func (r FactorPackRow) emissionFactor(pack FactorPack, factorSetID domain.ID) (domain.EmissionFactor, error) {
	metadataJSON, err := compactMetadata(r.Metadata)
	if err != nil {
		return domain.EmissionFactor{}, err
	}
	source := strings.TrimSpace(r.Source)
	if source == "" {
		source = pack.Source
	}
	return domain.EmissionFactor{
		ID:               r.id(pack),
		FactorSetID:      factorSetID,
		Source:           source,
		Scope:            domain.Scope(r.Scope),
		Level1:           r.Level1,
		Level2:           r.Level2,
		Level3:           r.Level3,
		Level4:           r.Level4,
		ColumnText:       r.ColumnText,
		ActivityType:     r.ActivityType,
		FuelType:         r.FuelType,
		VehicleType:      r.VehicleType,
		VehicleSizeClass: r.VehicleSizeClass,
		Substance:        r.Substance,
		InputUnit:        r.InputUnit,
		FactorUnit:       r.FactorUnit,
		GHG:              r.GHG,
		FactorValue:      r.FactorValue,
		MetadataJSON:     metadataJSON,
	}, nil
}

func (r FactorPackRow) id(pack FactorPack) domain.ID {
	if strings.TrimSpace(r.ID) != "" {
		return domain.ID(r.ID)
	}
	return domain.ID("emission_factor_" + sanitizeIDPart(pack.Source) + "_" + strconv.Itoa(pack.Year) + "_" + sanitizeIDPart(pack.Version) + "_" + r.key())
}

func (r FactorPackRow) key() string {
	parts := []string{
		r.ActivityType,
		r.FuelType,
		r.VehicleType,
		r.VehicleSizeClass,
		r.Substance,
		r.InputUnit,
		r.FactorUnit,
		r.GHG,
	}
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		part = sanitizeIDPart(part)
		if part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, "_")
}

func compactMetadata(raw json.RawMessage) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "{}", nil
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", err
	}
	if obj == nil {
		return "{}", nil
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", err
	}
	return compact.String(), nil
}

func sanitizeIDPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("/", "_", "-", "_", " ", "_", ".", "_")
	value = replacer.Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}
