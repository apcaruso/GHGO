package report

import (
	"fmt"
	"sort"

	"ghgo/internal/domain"
	"ghgo/internal/store"
)

type ElectricityDetailTable struct {
	Rows []ElectricityDetailRow
}

type ElectricityDetailRow struct {
	FacilityID     string
	Month          int
	MonthName      string
	ConsumptionKWh float64

	PrimaryMethod string

	LocationBasedFactorValue *float64
	LocationBasedFactorUnit  string
	LocationBasedKgCO2e      *float64

	MarketBasedFactorValue *float64
	MarketBasedFactorUnit  string
	MarketBasedKgCO2e      *float64
	MarketBasedSource      string

	IsMarketBasedPrimary bool
}

type NaturalGasDetailTable struct {
	Rows []NaturalGasDetailRow
}

type NaturalGasDetailRow struct {
	FacilityID      string
	Month           int
	MonthName       string
	ConsumptionSmc  float64
	FactorValue     float64
	FactorUnit      string
	EmissionsKgCO2e float64
}

type MobileDetailTable struct {
	Method       string
	FuelRows     []MobileFuelRow
	DistanceRows []MobileDistanceRow
	TotalKgCO2e  float64
}

type MobileFuelRow struct {
	FuelType        string
	Litres          float64
	FactorValue     float64
	FactorUnit      string
	EmissionsKgCO2e float64
}

type MobileDistanceRow struct {
	VehicleName      string
	Plate            string
	VehicleType      string
	VehicleSizeClass string
	FuelType         string
	Km               float64
	FactorValue      float64
	FactorUnit       string
	EmissionsKgCO2e  float64
}

type RefrigerantsDetailTable struct {
	Rows []RefrigerantsDetailRow
}

type RefrigerantsDetailRow struct {
	FacilityID      string
	Substance       string
	QuantityKg      float64
	FactorValue     float64
	FactorUnit      string
	EmissionsKgCO2e float64
}

func buildElectricityDetail(data reportData) ElectricityDetailTable {
	groups := map[string]*ElectricityDetailRow{}
	for _, row := range data.rows {
		record := row.ActivityRecord
		result := row.CalculationResult
		if result.Vector != domain.ActivityVectorElectricity {
			continue
		}
		detail := groups[record.ID]
		if detail == nil {
			detail = &ElectricityDetailRow{
				FacilityID:     facilityID(record),
				Month:          monthNumber(record),
				MonthName:      monthName(monthNumber(record)),
				ConsumptionKWh: record.Amount,
				PrimaryMethod:  string(domain.ActivityMethodLocationBased),
			}
			groups[record.ID] = detail
		}
		if result.IsPrimary {
			detail.PrimaryMethod = string(result.Method)
			detail.IsMarketBasedPrimary = result.Method == domain.ActivityMethodMarketBased
		}
		switch result.Method {
		case domain.ActivityMethodLocationBased:
			detail.LocationBasedFactorValue = floatPtr(result.FactorValue)
			detail.LocationBasedFactorUnit = result.FactorUnit
			detail.LocationBasedKgCO2e = floatPtr(result.EmissionsKgCO2e)
		case domain.ActivityMethodMarketBased:
			detail.MarketBasedFactorValue = floatPtr(result.FactorValue)
			detail.MarketBasedFactorUnit = result.FactorUnit
			detail.MarketBasedKgCO2e = floatPtr(result.EmissionsKgCO2e)
			detail.MarketBasedSource = result.FactorSource
		}
	}

	rows := make([]ElectricityDetailRow, 0, len(groups))
	for _, row := range groups {
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].FacilityID != rows[j].FacilityID {
			return rows[i].FacilityID < rows[j].FacilityID
		}
		return rows[i].Month < rows[j].Month
	})
	return ElectricityDetailTable{Rows: rows}
}

func buildNaturalGasDetail(data reportData) NaturalGasDetailTable {
	rows := []NaturalGasDetailRow{}
	for _, row := range data.rows {
		record := row.ActivityRecord
		result := row.CalculationResult
		if !result.IsPrimary || result.Vector != domain.ActivityVectorNaturalGas {
			continue
		}
		rows = append(rows, NaturalGasDetailRow{
			FacilityID:      facilityID(record),
			Month:           monthNumber(record),
			MonthName:       monthName(monthNumber(record)),
			ConsumptionSmc:  record.Amount,
			FactorValue:     result.FactorValue,
			FactorUnit:      result.FactorUnit,
			EmissionsKgCO2e: result.EmissionsKgCO2e,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].FacilityID != rows[j].FacilityID {
			return rows[i].FacilityID < rows[j].FacilityID
		}
		return rows[i].Month < rows[j].Month
	})
	return NaturalGasDetailTable{Rows: rows}
}

func buildMobileDetail(data reportData) (MobileDetailTable, error) {
	method, err := detectMobileMethod(data.rows)
	if err != nil {
		return MobileDetailTable{}, err
	}
	table := MobileDetailTable{Method: string(method)}
	for _, row := range data.rows {
		record := row.ActivityRecord
		result := row.CalculationResult
		if !result.IsPrimary || result.Vector != domain.ActivityVectorMobileCombustion {
			continue
		}
		table.TotalKgCO2e += result.EmissionsKgCO2e
		switch result.Method {
		case domain.ActivityMethodFuelBased:
			table.FuelRows = append(table.FuelRows, MobileFuelRow{
				FuelType:        record.FuelType,
				Litres:          record.Amount,
				FactorValue:     result.FactorValue,
				FactorUnit:      result.FactorUnit,
				EmissionsKgCO2e: result.EmissionsKgCO2e,
			})
		case domain.ActivityMethodDistanceBased:
			table.DistanceRows = append(table.DistanceRows, MobileDistanceRow{
				VehicleName:      record.VehicleName,
				Plate:            record.Plate,
				VehicleType:      record.VehicleType,
				VehicleSizeClass: record.VehicleSizeClass,
				FuelType:         record.FuelType,
				Km:               record.Amount,
				FactorValue:      result.FactorValue,
				FactorUnit:       result.FactorUnit,
				EmissionsKgCO2e:  result.EmissionsKgCO2e,
			})
		}
	}
	sort.Slice(table.FuelRows, func(i, j int) bool {
		return table.FuelRows[i].FuelType < table.FuelRows[j].FuelType
	})
	sort.Slice(table.DistanceRows, func(i, j int) bool {
		if table.DistanceRows[i].VehicleName != table.DistanceRows[j].VehicleName {
			return table.DistanceRows[i].VehicleName < table.DistanceRows[j].VehicleName
		}
		return table.DistanceRows[i].Plate < table.DistanceRows[j].Plate
	})
	return table, nil
}

func detectMobileMethod(rows []store.ReportResultRow) (domain.ActivityMethod, error) {
	method := domain.ActivityMethod("")
	for _, row := range rows {
		result := row.CalculationResult
		if result.Vector != domain.ActivityVectorMobileCombustion {
			continue
		}
		if result.Method != domain.ActivityMethodFuelBased && result.Method != domain.ActivityMethodDistanceBased {
			continue
		}
		if method == "" {
			method = result.Method
			continue
		}
		if method != result.Method {
			return "", fmt.Errorf("mobile methods %q and %q appear in calculation run: %w", method, result.Method, ErrMixedMobileMethods)
		}
	}
	return method, nil
}

func buildRefrigerantsDetail(data reportData) RefrigerantsDetailTable {
	rows := []RefrigerantsDetailRow{}
	for _, row := range data.rows {
		record := row.ActivityRecord
		result := row.CalculationResult
		if !result.IsPrimary || result.Vector != domain.ActivityVectorRefrigerants {
			continue
		}
		rows = append(rows, RefrigerantsDetailRow{
			FacilityID:      facilityID(record),
			Substance:       record.Substance,
			QuantityKg:      record.Amount,
			FactorValue:     result.FactorValue,
			FactorUnit:      result.FactorUnit,
			EmissionsKgCO2e: result.EmissionsKgCO2e,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].FacilityID != rows[j].FacilityID {
			return rows[i].FacilityID < rows[j].FacilityID
		}
		return rows[i].Substance < rows[j].Substance
	})
	return RefrigerantsDetailTable{Rows: rows}
}
