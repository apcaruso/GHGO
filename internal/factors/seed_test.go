package factors

import (
	"context"
	"encoding/json"
	"testing"

	"ghgo/factorpacks"
	"ghgo/internal/vocab"
)

func TestDefaultFactorPackIsVersionedData(t *testing.T) {
	pack, err := LoadFactorPack(factorpacks.FS, defaultFactorPackPath)
	if err != nil {
		t.Fatalf("load default factor pack: %v", err)
	}
	if pack.ID != defra2025FactorSetID || pack.Name != defra2025Name || pack.Source != defra2025Source || pack.Year != defra2025Year || pack.Version != defra2025Version {
		t.Fatalf("pack identity = %#v, want DEFRA/DESNZ 2025 identity", pack)
	}
	if len(pack.Rows) != 80 {
		t.Fatalf("pack row count = %d, want 80", len(pack.Rows))
	}

	var metadata map[string]bool
	if err := json.Unmarshal(pack.Metadata, &metadata); err != nil {
		t.Fatalf("pack metadata: %v", err)
	}
	if !metadata["normalized"] || !metadata["seeded"] {
		t.Fatalf("pack metadata = %#v, want normalized and seeded flags", metadata)
	}
}

func TestEnsureDefaultFactorsFreshDatabaseIsIdempotent(t *testing.T) {
	st := newTestStore(t)

	factorSet, err := EnsureDefaultFactors(context.Background(), st)
	if err != nil {
		t.Fatalf("ensure default factors: %v", err)
	}
	if factorSet == nil || factorSet.Name != defra2025Name {
		t.Fatalf("factor set = %#v, want %s", factorSet, defra2025Name)
	}
	count, err := st.CountEmissionFactorsBySet(factorSet.ID)
	if err != nil {
		t.Fatalf("count factors: %v", err)
	}
	if count == 0 {
		t.Fatalf("factor count = 0, want seeded factors")
	}

	second, err := EnsureDefaultFactors(context.Background(), st)
	if err != nil {
		t.Fatalf("ensure default factors second call: %v", err)
	}
	if second.ID != factorSet.ID {
		t.Fatalf("second factor set id = %q, want %q", second.ID, factorSet.ID)
	}
	secondCount, err := st.CountEmissionFactorsBySet(factorSet.ID)
	if err != nil {
		t.Fatalf("count factors after second call: %v", err)
	}
	if secondCount != count {
		t.Fatalf("factor count after second call = %d, want %d", secondCount, count)
	}
}

func TestEnsureDefaultFactorsSupportsLookups(t *testing.T) {
	st := newTestStore(t)
	factorSet, err := EnsureDefaultFactors(context.Background(), st)
	if err != nil {
		t.Fatalf("ensure default factors: %v", err)
	}
	lookup := NewLookup(st, factorSet.ID)

	if _, err := lookup.FindElectricityLocationFactor(context.Background()); err != nil {
		t.Fatalf("lookup electricity: %v", err)
	}
	if _, err := lookup.FindNaturalGasFactor(context.Background(), string(vocab.UnitSmc)); err != nil {
		t.Fatalf("lookup natural gas: %v", err)
	}
	if _, err := lookup.FindMobileFuelFactor(context.Background(), string(vocab.FuelDiesel), string(vocab.UnitLitre)); err != nil {
		t.Fatalf("lookup mobile fuel: %v", err)
	}
	if _, err := lookup.FindVehicleDistanceFactor(context.Background(), string(vocab.VehicleCar), string(vocab.SizeSmall), string(vocab.FuelPetrol), string(vocab.UnitKm)); err != nil {
		t.Fatalf("lookup vehicle distance: %v", err)
	}
	if _, err := lookup.FindRefrigerantFactor(context.Background(), string(vocab.RefrigerantR410A), string(vocab.UnitKg)); err != nil {
		t.Fatalf("lookup refrigerant: %v", err)
	}
}
