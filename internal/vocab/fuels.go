package vocab

type FuelType string

const (
	FuelDiesel  FuelType = "diesel"
	FuelPetrol  FuelType = "petrol"
	FuelLPG     FuelType = "lpg"
	FuelCNG     FuelType = "cng"
	FuelHybrid  FuelType = "hybrid"
	FuelPHEV    FuelType = "phev"
	FuelBEV     FuelType = "bev"
	FuelUnknown FuelType = "unknown"
)

var fuelAliases = map[string]FuelType{
	"diesel":                          FuelDiesel,
	"gasolio":                         FuelDiesel,
	"petrol":                          FuelPetrol,
	"benzina":                         FuelPetrol,
	"gasoline":                        FuelPetrol,
	"lpg":                             FuelLPG,
	"gpl":                             FuelLPG,
	"cng":                             FuelCNG,
	"metano":                          FuelCNG,
	"methane":                         FuelCNG,
	"hybrid":                          FuelHybrid,
	"ibrida":                          FuelHybrid,
	"phev":                            FuelPHEV,
	"plug in":                         FuelPHEV,
	"plugin":                          FuelPHEV,
	"plug in hybrid":                  FuelPHEV,
	"plug in hybrid electric vehicle": FuelPHEV,
	"bev":                             FuelBEV,
	"electric":                        FuelBEV,
	"elettrica":                       FuelBEV,
	"battery electric vehicle":        FuelBEV,
	"unknown":                         FuelUnknown,
	"unknown fuel":                    FuelUnknown,
}

func NormalizeFuelType(s string) (FuelType, bool) {
	f, ok := fuelAliases[Key(s)]
	return f, ok
}

func (f FuelType) Valid() bool {
	switch f {
	case FuelDiesel, FuelPetrol, FuelLPG, FuelCNG, FuelHybrid, FuelPHEV, FuelBEV, FuelUnknown:
		return true
	}
	return false
}

func (f FuelType) EnglishLabel() string {
	switch f {
	case FuelDiesel:
		return "Diesel"
	case FuelPetrol:
		return "Petrol"
	case FuelLPG:
		return "LPG"
	case FuelCNG:
		return "CNG"
	case FuelHybrid:
		return "Hybrid"
	case FuelPHEV:
		return "Plug-in Hybrid Electric Vehicle"
	case FuelBEV:
		return "Battery Electric Vehicle"
	case FuelUnknown:
		return "Unknown"
	}
	return ""
}
