package report

import "ghgo/internal/domain"

type MonthlyEmissionsTable struct {
	Rows []MonthlyEmissionsRow
}

type MonthlyEmissionsRow struct {
	Month     int
	MonthName string

	ElectricityConsumptionKWh      float64
	ElectricityPrimaryKgCO2e       float64
	ElectricityLocationBasedKgCO2e *float64
	NaturalGasConsumptionSmc       float64
	NaturalGasKgCO2e               float64
	MonthlyPrimaryKgCO2e           float64
	MonthlyLocationBasedKgCO2e     *float64
}

func buildMonthlyEmissions(data reportData) MonthlyEmissionsTable {
	rowsByMonth := map[int]*MonthlyEmissionsRow{}
	seenConsumption := map[string]bool{}
	hasMonthlyData := false

	for _, row := range data.rows {
		record := row.ActivityRecord
		result := row.CalculationResult
		if record.SourceKind != domain.ActivitySourceKindElectricityMonthlyKWh && record.SourceKind != domain.ActivitySourceKindNaturalGasMonthlySMC {
			continue
		}

		month := monthNumber(record)
		monthly := monthlyRow(rowsByMonth, month)
		hasMonthlyData = true

		switch record.SourceKind {
		case domain.ActivitySourceKindElectricityMonthlyKWh:
			if !seenConsumption[record.ID] {
				monthly.ElectricityConsumptionKWh += record.Amount
				seenConsumption[record.ID] = true
			}
			if result.IsPrimary && result.Vector == domain.ActivityVectorElectricity {
				monthly.ElectricityPrimaryKgCO2e += result.EmissionsKgCO2e
			}
			if result.Vector == domain.ActivityVectorElectricity && result.Method == domain.ActivityMethodLocationBased && data.hasLocationBasedComparison {
				if monthly.ElectricityLocationBasedKgCO2e == nil {
					monthly.ElectricityLocationBasedKgCO2e = floatPtr(0)
				}
				*monthly.ElectricityLocationBasedKgCO2e += result.EmissionsKgCO2e
			}
		case domain.ActivitySourceKindNaturalGasMonthlySMC:
			if !seenConsumption[record.ID] {
				monthly.NaturalGasConsumptionSmc += record.Amount
				seenConsumption[record.ID] = true
			}
			if result.IsPrimary && result.Vector == domain.ActivityVectorNaturalGas {
				monthly.NaturalGasKgCO2e += result.EmissionsKgCO2e
			}
		}
	}

	if !hasMonthlyData {
		return MonthlyEmissionsTable{}
	}

	rows := make([]MonthlyEmissionsRow, 0, 12)
	for month := 1; month <= 12; month++ {
		monthly := monthlyRow(rowsByMonth, month)
		monthly.MonthlyPrimaryKgCO2e = monthly.ElectricityPrimaryKgCO2e + monthly.NaturalGasKgCO2e
		if data.hasLocationBasedComparison {
			electricity := 0.0
			if monthly.ElectricityLocationBasedKgCO2e != nil {
				electricity = *monthly.ElectricityLocationBasedKgCO2e
			} else {
				monthly.ElectricityLocationBasedKgCO2e = floatPtr(0)
			}
			monthly.MonthlyLocationBasedKgCO2e = floatPtr(electricity + monthly.NaturalGasKgCO2e)
		}
		rows = append(rows, *monthly)
	}
	return MonthlyEmissionsTable{Rows: rows}
}

func monthlyRow(rowsByMonth map[int]*MonthlyEmissionsRow, month int) *MonthlyEmissionsRow {
	row := rowsByMonth[month]
	if row != nil {
		return row
	}
	row = &MonthlyEmissionsRow{Month: month, MonthName: monthName(month)}
	rowsByMonth[month] = row
	return row
}
