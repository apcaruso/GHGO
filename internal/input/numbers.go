package input

import (
	"strconv"
	"strings"
	"unicode"

	"ghgo/internal/vocab"
)

func ParsePositiveNumber(s string) (float64, []ParseIssue, bool) {
	s = vocab.NormalizeSpace(s)
	lower := strings.ToLower(s)
	if lower == "" || lower == "-" || lower == "n/a" || lower == "n.d." || lower == "na" || lower == "nd" {
		return 0, []ParseIssue{issue(CodeInvalidNumber, "Value is not a supported positive number.")}, false
	}

	if strings.HasPrefix(s, "-") {
		return 0, []ParseIssue{issue(CodeNegativeValue, "Value must not be negative.")}, false
	}

	if containsLetter(s) {
		return 0, []ParseIssue{issue(CodeUnitInNumber, "Numeric value must not include a unit.")}, false
	}

	if strings.Contains(s, " ") {
		return 0, []ParseIssue{issue(CodeInvalidNumber, "Value is not a supported positive number.")}, false
	}

	dotCount := strings.Count(s, ".")
	commaCount := strings.Count(s, ",")
	if dotCount+commaCount > 1 {
		return 0, []ParseIssue{issue(CodeInvalidNumber, "Value is not a supported positive number.")}, false
	}

	decimal := s
	if dotCount == 1 || commaCount == 1 {
		sep := "."
		if commaCount == 1 {
			sep = ","
		}

		parts := strings.Split(s, sep)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return 0, []ParseIssue{issue(CodeInvalidNumber, "Value is not a supported positive number.")}, false
		}
		if len(parts[1]) == 3 && len(parts[0]) >= 1 && len(parts[0]) <= 3 {
			return 0, []ParseIssue{issue(CodeInvalidNumber, "Ambiguous thousands separators are not supported.")}, false
		}

		if commaCount == 1 {
			decimal = strings.ReplaceAll(s, ",", ".")
		}
	}

	if !validDecimalChars(decimal) {
		return 0, []ParseIssue{issue(CodeInvalidNumber, "Value is not a supported positive number.")}, false
	}

	value, err := strconv.ParseFloat(decimal, 64)
	if err != nil {
		return 0, []ParseIssue{issue(CodeInvalidNumber, "Value is not a supported positive number.")}, false
	}
	if value < 0 {
		return 0, []ParseIssue{issue(CodeNegativeValue, "Value must not be negative.")}, false
	}
	if value == 0 {
		return value, []ParseIssue{issue(CodeZeroValue, "Value is zero.")}, true
	}

	return value, nil, true
}

func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func validDecimalChars(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			return false
		}
	}
	return true
}
