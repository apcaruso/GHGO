package vocab

type Vector string

const (
	VectorElectricity      Vector = "electricity"
	VectorNaturalGas       Vector = "natural_gas"
	VectorMobileCombustion Vector = "mobile_combustion"
	VectorRefrigerants     Vector = "refrigerants"
)

func (v Vector) Valid() bool {
	switch v {
	case VectorElectricity, VectorNaturalGas, VectorMobileCombustion, VectorRefrigerants:
		return true
	}
	return false
}

func (v Vector) EnglishLabel() string {
	switch v {
	case VectorElectricity:
		return "Electricity"
	case VectorNaturalGas:
		return "Natural gas"
	case VectorMobileCombustion:
		return "Mobile combustion"
	case VectorRefrigerants:
		return "Refrigerants"
	}
	return ""
}
