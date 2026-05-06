package report

import (
	"ghgo/internal/domain"
	"ghgo/internal/store"
)

type ExecutiveSummaryTable struct {
	ReportingPeriodID string
	FactorSetID       string

	PrimaryTotalKgCO2e       float64
	PrimaryTotalTCO2e        float64
	LocationBasedTotalKgCO2e *float64
	LocationBasedTotalTCO2e  *float64

	ElectricityPrimaryMethod   string
	HasLocationBasedComparison bool
	Scope1PrimaryKgCO2e        float64
	Scope2PrimaryKgCO2e        float64
	Scope2LocationBasedKgCO2e  *float64
}

type ScopeSummaryTable struct {
	Rows []ScopeSummaryRow
}

type ScopeSummaryRow struct {
	Scope         string
	PrimaryKgCO2e float64
	PrimaryTCO2e  float64
	PrimaryShare  float64

	LocationBasedKgCO2e *float64
	LocationBasedTCO2e  *float64
	LocationBasedShare  *float64
}

type VectorSummaryTable struct {
	Rows []VectorSummaryRow
}

type VectorSummaryRow struct {
	Vector          string
	ActivitySummary string
	PrimaryKgCO2e   float64
	PrimaryTCO2e    float64
	PrimaryShare    float64

	LocationBasedKgCO2e *float64
	LocationBasedTCO2e  *float64
	LocationBasedShare  *float64
}

func buildExecutiveSummary(data reportData) ExecutiveSummaryTable {
	table := ExecutiveSummaryTable{
		ReportingPeriodID:          data.run.ReportingPeriodID,
		FactorSetID:                data.run.FactorSetID,
		PrimaryTotalKgCO2e:         data.primaryTotalKgCO2e,
		PrimaryTotalTCO2e:          kgToT(data.primaryTotalKgCO2e),
		ElectricityPrimaryMethod:   string(data.electricityPrimaryMethod),
		HasLocationBasedComparison: data.hasLocationBasedComparison,
		Scope1PrimaryKgCO2e:        data.scope1PrimaryKgCO2e,
		Scope2PrimaryKgCO2e:        data.scope2PrimaryKgCO2e,
	}
	if data.hasLocationBasedComparison {
		table.LocationBasedTotalKgCO2e = floatPtr(data.locationBasedTotalKgCO2e)
		table.LocationBasedTotalTCO2e = floatPtr(kgToT(data.locationBasedTotalKgCO2e))
		table.Scope2LocationBasedKgCO2e = floatPtr(data.scope2LocationBasedKgCO2e)
	}
	return table
}

func buildScopeSummary(data reportData) ScopeSummaryTable {
	scope1 := scopeSummaryRow("Scope 1", data.scope1PrimaryKgCO2e, data.primaryTotalKgCO2e)
	scope2 := scopeSummaryRow("Scope 2", data.scope2PrimaryKgCO2e, data.primaryTotalKgCO2e)
	total := scopeSummaryRow("Total", data.primaryTotalKgCO2e, data.primaryTotalKgCO2e)

	if data.hasLocationBasedComparison {
		scope1.setLocationBased(data.scope1PrimaryKgCO2e, data.locationBasedTotalKgCO2e)
		scope2.setLocationBased(data.scope2LocationBasedKgCO2e, data.locationBasedTotalKgCO2e)
		total.setLocationBased(data.locationBasedTotalKgCO2e, data.locationBasedTotalKgCO2e)
	}

	return ScopeSummaryTable{Rows: []ScopeSummaryRow{scope1, scope2, total}}
}

func scopeSummaryRow(scope string, kg float64, total float64) ScopeSummaryRow {
	return ScopeSummaryRow{
		Scope:         scope,
		PrimaryKgCO2e: kg,
		PrimaryTCO2e:  kgToT(kg),
		PrimaryShare:  share(kg, total),
	}
}

func (r *ScopeSummaryRow) setLocationBased(kg float64, total float64) {
	r.LocationBasedKgCO2e = floatPtr(kg)
	r.LocationBasedTCO2e = floatPtr(kgToT(kg))
	r.LocationBasedShare = floatPtr(share(kg, total))
}

func buildVectorSummary(data reportData) VectorSummaryTable {
	primary := map[domain.ActivityVector]float64{}
	locationBased := map[domain.ActivityVector]float64{}
	amounts := vectorAmounts(data.rows)

	for _, row := range data.rows {
		result := row.CalculationResult
		if result.IsPrimary {
			primary[result.Vector] += result.EmissionsKgCO2e
		}
		if result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodLocationBased {
			locationBased[result.Vector] += result.EmissionsKgCO2e
		}
	}

	order := []domain.ActivityVector{
		domain.ActivityVectorElectricity,
		domain.ActivityVectorNaturalGas,
		domain.ActivityVectorMobileCombustion,
		domain.ActivityVectorRefrigerants,
	}
	names := map[domain.ActivityVector]string{
		domain.ActivityVectorElectricity:      "Electricity",
		domain.ActivityVectorNaturalGas:       "Natural gas",
		domain.ActivityVectorMobileCombustion: "Mobile combustion",
		domain.ActivityVectorRefrigerants:     "Refrigerants",
	}

	rows := make([]VectorSummaryRow, 0, len(order)+1)
	for _, vector := range order {
		row := VectorSummaryRow{
			Vector:          names[vector],
			ActivitySummary: amounts.summary(vector),
			PrimaryKgCO2e:   primary[vector],
			PrimaryTCO2e:    kgToT(primary[vector]),
			PrimaryShare:    share(primary[vector], data.primaryTotalKgCO2e),
		}
		if data.hasLocationBasedComparison {
			kg := primary[vector]
			if vector == domain.ActivityVectorElectricity {
				kg = locationBased[vector]
			}
			row.LocationBasedKgCO2e = floatPtr(kg)
			row.LocationBasedTCO2e = floatPtr(kgToT(kg))
			row.LocationBasedShare = floatPtr(share(kg, data.locationBasedTotalKgCO2e))
		}
		rows = append(rows, row)
	}

	total := VectorSummaryRow{
		Vector:        "Total",
		PrimaryKgCO2e: data.primaryTotalKgCO2e,
		PrimaryTCO2e:  kgToT(data.primaryTotalKgCO2e),
		PrimaryShare:  share(data.primaryTotalKgCO2e, data.primaryTotalKgCO2e),
	}
	if data.hasLocationBasedComparison {
		total.LocationBasedKgCO2e = floatPtr(data.locationBasedTotalKgCO2e)
		total.LocationBasedTCO2e = floatPtr(kgToT(data.locationBasedTotalKgCO2e))
		total.LocationBasedShare = floatPtr(share(data.locationBasedTotalKgCO2e, data.locationBasedTotalKgCO2e))
	}
	rows = append(rows, total)
	return VectorSummaryTable{Rows: rows}
}

type vectorActivityAmounts struct {
	seen        map[string]bool
	electricity float64
	naturalGas  float64
	mobile      float64
	refrigerant float64
	mobileUnit  string
}

func vectorAmounts(rows []store.ReportResultRow) vectorActivityAmounts {
	amounts := vectorActivityAmounts{seen: map[string]bool{}}
	for _, row := range rows {
		record := row.ActivityRecord
		if amounts.seen[record.ID] {
			continue
		}
		amounts.seen[record.ID] = true
		switch record.Vector {
		case domain.ActivityVectorElectricity:
			amounts.electricity += record.Amount
		case domain.ActivityVectorNaturalGas:
			amounts.naturalGas += record.Amount
		case domain.ActivityVectorMobileCombustion:
			amounts.mobile += record.Amount
			if amounts.mobileUnit == "" {
				amounts.mobileUnit = record.Unit
			}
		case domain.ActivityVectorRefrigerants:
			amounts.refrigerant += record.Amount
		}
	}
	return amounts
}

func (a vectorActivityAmounts) summary(vector domain.ActivityVector) string {
	switch vector {
	case domain.ActivityVectorElectricity:
		return activitySummary(a.electricity, "kWh")
	case domain.ActivityVectorNaturalGas:
		return activitySummary(a.naturalGas, "Smc")
	case domain.ActivityVectorMobileCombustion:
		unit := a.mobileUnit
		if unit == "" {
			unit = "L"
		}
		return activitySummary(a.mobile, unit)
	case domain.ActivityVectorRefrigerants:
		return activitySummary(a.refrigerant, "kg")
	}
	return ""
}
