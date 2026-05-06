package vocab

type VehicleType string

const (
	VehicleCar       VehicleType = "car"
	VehicleVan       VehicleType = "van"
	VehicleMotorbike VehicleType = "motorbike"
)

var vehicleAliases = map[string]VehicleType{
	"car":           VehicleCar,
	"auto":          VehicleCar,
	"automobile":    VehicleCar,
	"passenger car": VehicleCar,
	"van":           VehicleVan,
	"furgone":       VehicleVan,
	"light van":     VehicleVan,
	"delivery van":  VehicleVan,
	"motorbike":     VehicleMotorbike,
	"motorcycle":    VehicleMotorbike,
	"moto":          VehicleMotorbike,
}

func NormalizeVehicleType(s string) (VehicleType, bool) {
	v, ok := vehicleAliases[Key(s)]
	return v, ok
}

func (v VehicleType) Valid() bool {
	switch v {
	case VehicleCar, VehicleVan, VehicleMotorbike:
		return true
	}
	return false
}

func (v VehicleType) EnglishLabel() string {
	switch v {
	case VehicleCar:
		return "Car"
	case VehicleVan:
		return "Van"
	case VehicleMotorbike:
		return "Motorbike"
	}
	return ""
}

type VehicleSizeClass string

const (
	SizeSmall    VehicleSizeClass = "small"
	SizeMedium   VehicleSizeClass = "medium"
	SizeLarge    VehicleSizeClass = "large"
	SizeAverage  VehicleSizeClass = "average"
	SizeClassI   VehicleSizeClass = "class_i"
	SizeClassII  VehicleSizeClass = "class_ii"
	SizeClassIII VehicleSizeClass = "class_iii"
)

var carSizeAliases = map[string]VehicleSizeClass{
	"small":        SizeSmall,
	"small car":    SizeSmall,
	"piccola":      SizeSmall,
	"medium":       SizeMedium,
	"medium car":   SizeMedium,
	"media":        SizeMedium,
	"large":        SizeLarge,
	"large car":    SizeLarge,
	"grande":       SizeLarge,
	"average":      SizeAverage,
	"average car":  SizeAverage,
	"media flotta": SizeAverage,
}

var vanSizeAliases = map[string]VehicleSizeClass{
	"class i":     SizeClassI,
	"class 1":     SizeClassI,
	"classe i":    SizeClassI,
	"classe 1":    SizeClassI,
	"class ii":    SizeClassII,
	"class 2":     SizeClassII,
	"classe ii":   SizeClassII,
	"classe 2":    SizeClassII,
	"class iii":   SizeClassIII,
	"class 3":     SizeClassIII,
	"classe iii":  SizeClassIII,
	"classe 3":    SizeClassIII,
	"average":     SizeAverage,
	"average van": SizeAverage,
	"medio":       SizeAverage,
}

var motorbikeSizeAliases = map[string]VehicleSizeClass{
	"small":   SizeSmall,
	"piccola": SizeSmall,
	"medium":  SizeMedium,
	"media":   SizeMedium,
	"large":   SizeLarge,
	"grande":  SizeLarge,
	"average": SizeAverage,
}

func NormalizeVehicleSizeClass(vehicleType VehicleType, s string) (VehicleSizeClass, bool) {
	key := Key(s)
	switch vehicleType {
	case VehicleCar:
		size, ok := carSizeAliases[key]
		return size, ok
	case VehicleVan:
		size, ok := vanSizeAliases[key]
		return size, ok
	case VehicleMotorbike:
		size, ok := motorbikeSizeAliases[key]
		return size, ok
	}
	return "", false
}

func (s VehicleSizeClass) Valid() bool {
	switch s {
	case SizeSmall, SizeMedium, SizeLarge, SizeAverage, SizeClassI, SizeClassII, SizeClassIII:
		return true
	}
	return false
}

func (s VehicleSizeClass) EnglishLabel(vehicleType VehicleType) string {
	if !VehicleSizeClassCompatible(vehicleType, s) {
		return ""
	}

	switch s {
	case SizeSmall:
		return "Small"
	case SizeMedium:
		return "Medium"
	case SizeLarge:
		return "Large"
	case SizeAverage:
		return "Average"
	case SizeClassI:
		return "Class I"
	case SizeClassII:
		return "Class II"
	case SizeClassIII:
		return "Class III"
	}
	return ""
}

func VehicleSizeClassCompatible(vehicleType VehicleType, size VehicleSizeClass) bool {
	switch vehicleType {
	case VehicleCar, VehicleMotorbike:
		switch size {
		case SizeSmall, SizeMedium, SizeLarge, SizeAverage:
			return true
		}
	case VehicleVan:
		switch size {
		case SizeClassI, SizeClassII, SizeClassIII, SizeAverage:
			return true
		}
	}
	return false
}
