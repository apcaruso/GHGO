package vocab

type Method string

const (
	MethodLocationBased Method = "location_based"
	MethodMarketBased   Method = "market_based"
	MethodFuelBased     Method = "fuel_based"
	MethodDistanceBased Method = "distance_based"
	MethodDirectGWP     Method = "direct_gwp"
)

func (m Method) Valid() bool {
	switch m {
	case MethodLocationBased, MethodMarketBased, MethodFuelBased, MethodDistanceBased, MethodDirectGWP:
		return true
	}
	return false
}

func (m Method) EnglishLabel() string {
	switch m {
	case MethodLocationBased:
		return "Location-based"
	case MethodMarketBased:
		return "Market-based"
	case MethodFuelBased:
		return "Fuel-based"
	case MethodDistanceBased:
		return "Distance-based"
	case MethodDirectGWP:
		return "Direct GWP"
	}
	return ""
}
