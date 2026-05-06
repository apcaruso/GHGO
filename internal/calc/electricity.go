package calc

import (
	"context"
	"encoding/json"
	"fmt"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/vocab"
)

const guaranteesOfOriginSource = "Guarantees of Origin"

func calculateElectricity(ctx context.Context, runID domain.ID, factorSet *domain.FactorSet, record domain.ActivityRecord, settings *domain.ElectricitySettings, lookup *factors.Lookup) ([]domain.CalculationResult, error) {
	if err := validateElectricityRecord(record); err != nil {
		return nil, err
	}
	hasFullGO, err := electricityHasFullGO(record, settings)
	if err != nil {
		return nil, err
	}

	factor, err := lookup.FindForActivityRecord(ctx, record)
	if err != nil {
		return nil, wrapFactorError(record, err)
	}

	locationBased, err := calculationResultFromFactor(runID, record, domain.ActivityMethodLocationBased, factorSet, factor, !hasFullGO, `{}`)
	if err != nil {
		return nil, err
	}
	if !hasFullGO {
		return []domain.CalculationResult{locationBased}, nil
	}

	marketBased, err := marketBasedGOResult(runID, record)
	if err != nil {
		return nil, err
	}
	return []domain.CalculationResult{locationBased, marketBased}, nil
}

func validateElectricityRecord(record domain.ActivityRecord) error {
	if record.Scope != domain.Scope2 {
		return unsupportedRecord("electricity activity record %q has scope %d, expected scope 2", record.ID, record.Scope)
	}
	if record.Vector != domain.ActivityVectorElectricity {
		return unsupportedRecord("electricity activity record %q has unsupported vector %q", record.ID, record.Vector)
	}
	if record.Method != domain.ActivityMethodLocationBased {
		return unsupportedRecord("electricity activity record %q has unsupported method %q", record.ID, record.Method)
	}
	if record.Unit != string(vocab.UnitKWh) {
		return unsupportedRecord("electricity activity record %q has unsupported unit %q", record.ID, record.Unit)
	}
	if record.FacilityID == nil || *record.FacilityID == "" {
		return unsupportedRecord("electricity activity record %q is missing facility id", record.ID)
	}
	return nil
}

func electricityHasFullGO(record domain.ActivityRecord, settings *domain.ElectricitySettings) (bool, error) {
	if settings == nil {
		return false, nil
	}
	if settings.HasGuaranteesOfOrigin && settings.GOCoverage == domain.GOCoverageFull {
		return true, nil
	}
	if !settings.HasGuaranteesOfOrigin && settings.GOCoverage == domain.GOCoverageNone {
		return false, nil
	}
	return false, invalidSettings("electricity settings for activity record %q are inconsistent: has_go=%t coverage=%q", record.ID, settings.HasGuaranteesOfOrigin, settings.GOCoverage)
}

func marketBasedGOResult(runID domain.ID, record domain.ActivityRecord) (domain.CalculationResult, error) {
	id, err := newID("calculation_result")
	if err != nil {
		return domain.CalculationResult{}, err
	}
	notes, err := json.Marshal(map[string]string{
		"note": "full GO coverage; market-based emissions set to zero",
	})
	if err != nil {
		return domain.CalculationResult{}, fmt.Errorf("marshal market-based electricity notes: %w", err)
	}

	return domain.CalculationResult{
		ID:               id,
		CalculationRunID: runID,
		ActivityRecordID: record.ID,
		Scope:            record.Scope,
		Vector:           record.Vector,
		Method:           domain.ActivityMethodMarketBased,
		ActivityAmount:   record.Amount,
		ActivityUnit:     record.Unit,
		FactorValue:      0,
		FactorUnit:       "kgCO2e/kWh",
		FactorSource:     guaranteesOfOriginSource,
		EmissionsKgCO2e:  0,
		IsPrimary:        true,
		NotesJSON:        string(notes),
	}, nil
}
