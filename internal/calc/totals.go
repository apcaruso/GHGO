package calc

import "ghgo/internal/domain"

func summarizeResults(runID domain.ID, results []domain.CalculationResult) RunResult {
	var primaryTotal float64
	hasMarketBasedPrimaryElectricity := false
	for _, result := range results {
		if result.IsPrimary {
			primaryTotal += result.EmissionsKgCO2e
		}
		if result.IsPrimary && result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodMarketBased {
			hasMarketBasedPrimaryElectricity = true
		}
	}

	var locationBasedTotal *float64
	if hasMarketBasedPrimaryElectricity {
		total := 0.0
		for _, result := range results {
			if result.Scope == domain.Scope1 && result.IsPrimary {
				total += result.EmissionsKgCO2e
			}
			if result.Scope == domain.Scope2 && result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodLocationBased {
				total += result.EmissionsKgCO2e
			}
		}
		locationBasedTotal = &total
	}

	return RunResult{
		CalculationRunID:         runID,
		ResultsCreated:           len(results),
		PrimaryTotalKgCO2e:       primaryTotal,
		LocationBasedTotalKgCO2e: locationBasedTotal,
	}
}
