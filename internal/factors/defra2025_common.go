package factors

const (
	defra2025FactorSetID = "factor_set_defra_2025"
	defra2025Name        = "DEFRA/DESNZ 2025"
	defra2025Source      = "DEFRA"
	defra2025Year        = 2025
	defra2025Version     = "2025"
)

type defraRow struct {
	Scope         int
	OriginalScope string
	Level1        string
	Level2        string
	Level3        string
	Level4        string
	ColumnText    string
	UOM           string
	GHGUnit       string

	ConversionFactorText string
	ConversionFactor     float64
}
