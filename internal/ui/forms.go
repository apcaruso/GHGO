package ui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ghgo/internal/domain"
)

func newID(prefix string) (domain.ID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return domain.ID(prefix + "_" + hex.EncodeToString(b[:])), nil
}

func parseYear(value string) (int, error) {
	year, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || year <= 0 {
		return 0, fmt.Errorf("year must be a positive number")
	}
	return year, nil
}

func parseDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("date %q must use YYYY-MM-DD", value)
	}
	return date, nil
}

func validateCountryCode(value string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(value))
	if len(code) != 2 {
		return "", fmt.Errorf("country code must be two letters")
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return "", fmt.Errorf("country code must contain letters only")
		}
	}
	return code, nil
}

func mobileMethodFromLabel(label string) domain.MobileMethod {
	switch label {
	case "Fuel consumed":
		return domain.MobileMethodFuelBased
	case "Distance travelled":
		return domain.MobileMethodDistanceBased
	}
	return ""
}

func mobileMethodLabel(method domain.MobileMethod) string {
	switch method {
	case domain.MobileMethodFuelBased:
		return "Fuel consumed"
	case domain.MobileMethodDistanceBased:
		return "Distance travelled"
	}
	return ""
}
