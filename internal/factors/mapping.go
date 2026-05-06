package factors

import (
	"encoding/json"
	"strings"

	"ghgo/internal/domain"
	"ghgo/internal/vocab"
)

type factorCandidate struct {
	Scope  int
	Source string

	Level1     string
	Level2     string
	Level3     string
	Level4     string
	ColumnText string

	ActivityType     string
	FuelType         *string
	VehicleType      *string
	VehicleSizeClass *string
	Substance        *string

	InputUnit   string
	FactorUnit  string
	GHG         string
	FactorValue float64

	MetadataJSON string
}

func mapDEFRA2025Row(row defraRow) (*factorCandidate, bool) {
	if !supportedSourceRow(row) {
		return nil, false
	}

	mappers := []func(defraRow) (*factorCandidate, bool){
		mapElectricity,
		mapNaturalGas,
		mapMobileFuel,
		mapVehicleDistance,
		mapRefrigerant,
	}
	for _, mapper := range mappers {
		if candidate, ok := mapper(row); ok {
			return candidate, true
		}
	}
	return nil, false
}

func mapElectricity(row defraRow) (*factorCandidate, bool) {
	if !supportedSourceRow(row) {
		return nil, false
	}
	unit, ok := unitFromUOM(row.UOM)
	if !ok || row.Scope != 2 || unit != string(vocab.UnitKWh) {
		return nil, false
	}

	key := rowKey(row)
	if !strings.Contains(key, "electricity") {
		return nil, false
	}
	if !(strings.Contains(key, "uk electricity") || strings.Contains(key, "electricity uk") || strings.Contains(key, "electricity generated") || strings.Contains(key, "grid electricity")) {
		return nil, false
	}
	if strings.Contains(key, "transmission") || strings.Contains(key, "distribution") || strings.Contains(key, "market based") {
		return nil, false
	}

	candidate := baseCandidate(row, "purchased_electricity", string(vocab.UnitKWh), "kgCO2e/kWh")
	return &candidate, true
}

func mapNaturalGas(row defraRow) (*factorCandidate, bool) {
	if !supportedSourceRow(row) {
		return nil, false
	}
	unit, ok := unitFromUOM(row.UOM)
	if !ok || row.Scope != 1 || unit != string(vocab.UnitSmc) {
		return nil, false
	}
	if !strings.Contains(rowKey(row), "natural gas") {
		return nil, false
	}

	candidate := baseCandidate(row, "natural_gas", string(vocab.UnitSmc), "kgCO2e/Smc")
	return &candidate, true
}

func mapMobileFuel(row defraRow) (*factorCandidate, bool) {
	if !supportedSourceRow(row) {
		return nil, false
	}
	unit, ok := unitFromUOM(row.UOM)
	if !ok || row.Scope != 1 || unit != string(vocab.UnitLitre) {
		return nil, false
	}

	fuel, ok := mobileFuelFromRow(row)
	if !ok {
		return nil, false
	}

	fuelType := string(fuel)
	candidate := baseCandidate(row, fuelType+"_mobile", string(vocab.UnitLitre), "kgCO2e/L")
	candidate.FuelType = &fuelType
	return &candidate, true
}

func mapVehicleDistance(row defraRow) (*factorCandidate, bool) {
	if !supportedSourceRow(row) {
		return nil, false
	}
	unit, ok := unitFromUOM(row.UOM)
	if !ok || row.Scope != 1 || unit != string(vocab.UnitKm) {
		return nil, false
	}

	vehicleType, ok := vehicleTypeFromRow(row)
	if !ok {
		return nil, false
	}
	if vehicleType == vocab.VehicleCar && strings.Contains(vocab.Key(row.Level2), "market segment") {
		return nil, false
	}
	size, ok := vehicleSizeFromRow(row, vehicleType)
	if !ok {
		return nil, false
	}

	candidate := baseCandidate(row, "vehicle_distance", string(vocab.UnitKm), "kgCO2e/km")
	vehicleTypeValue := string(vehicleType)
	sizeValue := string(size)
	candidate.VehicleType = &vehicleTypeValue
	candidate.VehicleSizeClass = &sizeValue

	if fuel, ok := vehicleFuelFromRow(row); ok {
		fuelValue := string(fuel)
		candidate.FuelType = &fuelValue
	} else if vehicleType != vocab.VehicleMotorbike {
		return nil, false
	}

	return &candidate, true
}

func mapRefrigerant(row defraRow) (*factorCandidate, bool) {
	if !supportedSourceRow(row) {
		return nil, false
	}
	unit, ok := unitFromUOM(row.UOM)
	if !ok || row.Scope != 1 || unit != string(vocab.UnitKg) {
		return nil, false
	}

	refrigerant, ok := refrigerantFromRow(row)
	if !ok {
		return nil, false
	}
	columnKey := vocab.Key(row.ColumnText)
	if columnKey != "" && !strings.Contains(columnKey, "total emissions") {
		return nil, false
	}

	substance := string(refrigerant)
	candidate := baseCandidate(row, "refrigerant_leakage", string(vocab.UnitKg), "kgCO2e/kg")
	candidate.Substance = &substance
	return &candidate, true
}

func supportedSourceRow(row defraRow) bool {
	if row.Scope != 1 && row.Scope != 2 {
		return false
	}
	return isKgCO2e(row.GHGUnit) && !isWTT(row) && !isOutsideScopes(row)
}

func baseCandidate(row defraRow, activityType string, inputUnit string, factorUnit string) factorCandidate {
	return factorCandidate{
		Scope:        row.Scope,
		Source:       defra2025Source,
		Level1:       row.Level1,
		Level2:       row.Level2,
		Level3:       row.Level3,
		Level4:       row.Level4,
		ColumnText:   row.ColumnText,
		ActivityType: activityType,
		InputUnit:    inputUnit,
		FactorUnit:   factorUnit,
		GHG:          "kgCO2e",
		FactorValue:  row.ConversionFactor,
		MetadataJSON: metadataJSON(row),
	}
}

func (c factorCandidate) emissionFactor(id domain.ID, factorSetID domain.ID) domain.EmissionFactor {
	factor := domain.EmissionFactor{
		ID:          id,
		FactorSetID: factorSetID,
		Source:      c.Source,
		Scope:       domain.Scope(c.Scope),

		Level1:     c.Level1,
		Level2:     c.Level2,
		Level3:     c.Level3,
		Level4:     c.Level4,
		ColumnText: c.ColumnText,

		ActivityType: c.ActivityType,
		InputUnit:    c.InputUnit,
		FactorUnit:   c.FactorUnit,
		GHG:          c.GHG,
		FactorValue:  c.FactorValue,

		MetadataJSON: c.MetadataJSON,
	}
	if c.FuelType != nil {
		factor.FuelType = *c.FuelType
	}
	if c.VehicleType != nil {
		factor.VehicleType = *c.VehicleType
	}
	if c.VehicleSizeClass != nil {
		factor.VehicleSizeClass = *c.VehicleSizeClass
	}
	if c.Substance != nil {
		factor.Substance = *c.Substance
	}
	return factor
}

func metadataJSON(row defraRow) string {
	metadata := struct {
		OriginalScope string `json:"original_scope"`
		Level1        string `json:"level_1"`
		Level2        string `json:"level_2"`
		Level3        string `json:"level_3"`
		Level4        string `json:"level_4"`
		ColumnText    string `json:"column_text"`
		UOM           string `json:"uom"`
		GHGUnit       string `json:"ghg_unit"`
	}{
		OriginalScope: row.OriginalScope,
		Level1:        row.Level1,
		Level2:        row.Level2,
		Level3:        row.Level3,
		Level4:        row.Level4,
		ColumnText:    row.ColumnText,
		UOM:           row.UOM,
		GHGUnit:       row.GHGUnit,
	}
	data, _ := json.Marshal(metadata)
	return string(data)
}

func mobileFuelFromRow(row defraRow) (vocab.FuelType, bool) {
	fuel, ok := fuelFromRow(row)
	if !ok {
		return "", false
	}
	switch fuel {
	case vocab.FuelDiesel:
		if strings.Contains(rowKey(row), "diesel average biofuel blend") {
			return fuel, true
		}
	case vocab.FuelPetrol:
		if strings.Contains(rowKey(row), "petrol average biofuel blend") {
			return fuel, true
		}
	case vocab.FuelLPG, vocab.FuelCNG:
		return fuel, true
	}
	return "", false
}

func vehicleFuelFromRow(row defraRow) (vocab.FuelType, bool) {
	fuel, ok := fuelFromRow(row)
	if !ok {
		return "", false
	}
	switch fuel {
	case vocab.FuelDiesel, vocab.FuelPetrol, vocab.FuelHybrid, vocab.FuelCNG, vocab.FuelLPG, vocab.FuelUnknown, vocab.FuelPHEV, vocab.FuelBEV:
		return fuel, true
	}
	return "", false
}

func fuelFromRow(row defraRow) (vocab.FuelType, bool) {
	for _, value := range rowValues(row) {
		if fuel, ok := fuelFromText(value); ok {
			return fuel, true
		}
	}
	return "", false
}

func fuelFromText(s string) (vocab.FuelType, bool) {
	key := vocab.Key(s)
	switch {
	case strings.Contains(key, "plug in hybrid") || strings.Contains(key, "phev"):
		return vocab.FuelPHEV, true
	case strings.Contains(key, "battery electric") || strings.Contains(key, "bev"):
		return vocab.FuelBEV, true
	case containsWord(key, "hybrid"):
		return vocab.FuelHybrid, true
	case containsWord(key, "diesel"):
		return vocab.FuelDiesel, true
	case containsWord(key, "petrol") || containsWord(key, "gasoline"):
		return vocab.FuelPetrol, true
	case strings.Contains(key, "cng") || strings.Contains(key, "compressed natural gas"):
		return vocab.FuelCNG, true
	case containsWord(key, "lpg"):
		return vocab.FuelLPG, true
	case containsWord(key, "unknown"):
		return vocab.FuelUnknown, true
	}
	return "", false
}

func containsWord(key string, word string) bool {
	for _, field := range strings.Fields(key) {
		if field == word {
			return true
		}
	}
	return false
}

func vehicleTypeFromRow(row defraRow) (vocab.VehicleType, bool) {
	key := rowKey(row)
	switch {
	case strings.Contains(key, "motorbike") || strings.Contains(key, "motorcycle"):
		return vocab.VehicleMotorbike, true
	case strings.Contains(key, "van") || strings.Contains(key, "vans"):
		return vocab.VehicleVan, true
	case strings.Contains(key, "car") || strings.Contains(key, "cars"):
		return vocab.VehicleCar, true
	}
	return "", false
}

func vehicleSizeFromRow(row defraRow, vehicleType vocab.VehicleType) (vocab.VehicleSizeClass, bool) {
	for _, value := range rowValues(row) {
		key := vocab.Key(value)
		switch vehicleType {
		case vocab.VehicleCar, vocab.VehicleMotorbike:
			switch {
			case strings.Contains(key, "small"):
				return vocab.SizeSmall, true
			case strings.Contains(key, "medium"):
				return vocab.SizeMedium, true
			case strings.Contains(key, "large"):
				return vocab.SizeLarge, true
			case strings.Contains(key, "average"):
				return vocab.SizeAverage, true
			}
		case vocab.VehicleVan:
			switch {
			case strings.Contains(key, "class iii") || strings.Contains(key, "class 3"):
				return vocab.SizeClassIII, true
			case strings.Contains(key, "class ii") || strings.Contains(key, "class 2"):
				return vocab.SizeClassII, true
			case strings.Contains(key, "class i") || strings.Contains(key, "class 1"):
				return vocab.SizeClassI, true
			case strings.Contains(key, "average"):
				return vocab.SizeAverage, true
			}
		}
	}
	return "", false
}

func refrigerantFromRow(row defraRow) (vocab.Refrigerant, bool) {
	for _, value := range rowValues(row) {
		key := strings.ReplaceAll(vocab.Key(value), " ", "")
		switch {
		case strings.Contains(key, "r134a"):
			return vocab.RefrigerantR134a, true
		case strings.Contains(key, "r410a"):
			return vocab.RefrigerantR410A, true
		case strings.Contains(key, "r407c"):
			return vocab.RefrigerantR407C, true
		case strings.Contains(key, "r404a"):
			return vocab.RefrigerantR404A, true
		case strings.Contains(key, "r32"):
			return vocab.RefrigerantR32, true
		case strings.Contains(key, "r22"):
			return vocab.RefrigerantR22, true
		}
	}
	return "", false
}

func rowValues(row defraRow) []string {
	return []string{row.Level1, row.Level2, row.Level3, row.Level4, row.ColumnText}
}
