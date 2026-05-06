package vocab

type Unit string

const (
	UnitKWh   Unit = "kWh"
	UnitSmc   Unit = "Smc"
	UnitLitre Unit = "L"
	UnitKm    Unit = "km"
	UnitKg    Unit = "kg"
)

var unitAliases = map[string]Unit{
	"kwh":                  UnitKWh,
	"kw h":                 UnitKWh,
	"smc":                  UnitSmc,
	"standard cubic meter": UnitSmc,
	"standard cubic metre": UnitSmc,
	"l":                    UnitLitre,
	"litre":                UnitLitre,
	"liter":                UnitLitre,
	"litres":               UnitLitre,
	"liters":               UnitLitre,
	"lt":                   UnitLitre,
	"km":                   UnitKm,
	"kilometer":            UnitKm,
	"kilometre":            UnitKm,
	"kilometers":           UnitKm,
	"kilometres":           UnitKm,
	"kg":                   UnitKg,
	"kilogram":             UnitKg,
	"kilograms":            UnitKg,
}

func NormalizeUnit(s string) (Unit, bool) {
	u, ok := unitAliases[Key(s)]
	return u, ok
}

func (u Unit) Valid() bool {
	switch u {
	case UnitKWh, UnitSmc, UnitLitre, UnitKm, UnitKg:
		return true
	}
	return false
}
