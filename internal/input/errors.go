package input

const (
	CodeWrongColumnCount             = "wrong_column_count"
	CodeHeaderInBody                 = "header_in_body"
	CodeEmptyField                   = "empty_field"
	CodeInvalidMonth                 = "invalid_month"
	CodeDuplicateMonth               = "duplicate_month"
	CodeInvalidNumber                = "invalid_number"
	CodeNegativeValue                = "negative_value"
	CodeUnitInNumber                 = "unit_in_number"
	CodeInvalidFuelType              = "invalid_fuel_type"
	CodeInvalidVehicleType           = "invalid_vehicle_type"
	CodeInvalidVehicleSizeClass      = "invalid_vehicle_size_class"
	CodeIncompatibleVehicleSizeClass = "incompatible_vehicle_size_class"
	CodeInvalidRefrigerant           = "invalid_refrigerant"
	CodeUnsupportedFuelForMethod     = "unsupported_fuel_for_method"

	CodeZeroValue             = "zero_value"
	CodeDuplicateFuelType     = "duplicate_fuel_type"
	CodeDuplicatePlate        = "duplicate_plate"
	CodeDuplicateRefrigerant  = "duplicate_refrigerant"
	CodeUnknownFuelType       = "unknown_fuel_type"
	CodeBEVScope2NotEstimated = "bev_scope2_not_estimated"
	CodePHEVAverageFactor     = "phev_average_factor"
)

type ParseIssue struct {
	Code    string
	Message string
}

func issue(code, message string) ParseIssue {
	return ParseIssue{Code: code, Message: message}
}
