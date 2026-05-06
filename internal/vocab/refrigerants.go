package vocab

import "strings"

type Refrigerant string

const (
	RefrigerantR134a Refrigerant = "R134a"
	RefrigerantR410A Refrigerant = "R410A"
	RefrigerantR407C Refrigerant = "R407C"
	RefrigerantR404A Refrigerant = "R404A"
	RefrigerantR32   Refrigerant = "R32"
	RefrigerantR22   Refrigerant = "R22"
)

var refrigerantAliases = map[string]Refrigerant{
	"r134a": RefrigerantR134a,
	"r410a": RefrigerantR410A,
	"410a":  RefrigerantR410A,
	"r407c": RefrigerantR407C,
	"407c":  RefrigerantR407C,
	"r404a": RefrigerantR404A,
	"404a":  RefrigerantR404A,
	"r32":   RefrigerantR32,
	"32":    RefrigerantR32,
	"r22":   RefrigerantR22,
	"22":    RefrigerantR22,
}

func NormalizeRefrigerant(s string) (Refrigerant, bool) {
	key := strings.ReplaceAll(Key(s), " ", "")
	r, ok := refrigerantAliases[key]
	return r, ok
}

func (r Refrigerant) Valid() bool {
	switch r {
	case RefrigerantR134a, RefrigerantR410A, RefrigerantR407C, RefrigerantR404A, RefrigerantR32, RefrigerantR22:
		return true
	}
	return false
}

func (r Refrigerant) EnglishLabel() string {
	if !r.Valid() {
		return ""
	}
	return string(r)
}
