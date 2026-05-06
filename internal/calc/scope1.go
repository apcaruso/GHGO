package calc

import (
	"context"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/vocab"
)

func calculateScope1(ctx context.Context, runID domain.ID, factorSet *domain.FactorSet, record domain.ActivityRecord, periodSettings *domain.ReportingPeriodSettings, lookup *factors.Lookup) ([]domain.CalculationResult, error) {
	if record.Scope != domain.Scope1 {
		return nil, unsupportedRecord("activity record %q has scope %d, expected scope 1", record.ID, record.Scope)
	}
	switch record.Vector {
	case domain.ActivityVectorNaturalGas, domain.ActivityVectorMobileCombustion, domain.ActivityVectorRefrigerants:
	default:
		return nil, unsupportedRecord("scope 1 activity record %q has unsupported vector %q", record.ID, record.Vector)
	}

	if isMobileActivityRecord(record) {
		if err := validateMobileRecordMethod(record, periodSettings); err != nil {
			return nil, err
		}
	}

	factor, err := lookup.FindForActivityRecord(ctx, record)
	if err != nil {
		return nil, wrapFactorError(record, err)
	}
	result, err := calculationResultFromFactor(runID, record, record.Method, factorSet, factor, true, `{}`)
	if err != nil {
		return nil, err
	}
	return []domain.CalculationResult{result}, nil
}

func validateMobileRecordMethod(record domain.ActivityRecord, periodSettings *domain.ReportingPeriodSettings) error {
	if periodSettings == nil {
		return invalidSettings("reporting period settings are required for mobile activity record %q", record.ID)
	}
	switch periodSettings.MobileMethod {
	case domain.MobileMethodFuelBased:
		if record.SourceKind != domain.ActivitySourceKindMobileFuelLitres || record.Method != domain.ActivityMethodFuelBased || record.Unit != string(vocab.UnitLitre) {
			return invalidSettings("mobile method %q does not allow activity record %q", periodSettings.MobileMethod, record.ID)
		}
	case domain.MobileMethodDistanceBased:
		if record.SourceKind != domain.ActivitySourceKindVehicleDistanceKM || record.Method != domain.ActivityMethodDistanceBased || record.Unit != string(vocab.UnitKm) {
			return invalidSettings("mobile method %q does not allow activity record %q", periodSettings.MobileMethod, record.ID)
		}
	default:
		return invalidSettings("mobile method %q is not valid", periodSettings.MobileMethod)
	}
	return nil
}
