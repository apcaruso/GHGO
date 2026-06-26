package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/store"
)

func checkStore(ctx context.Context, st *store.Store) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if st == nil {
		return invalidOptions("store is required")
	}
	return nil
}

func cleanText(value string) string {
	return strings.TrimSpace(value)
}

func requiredID(name string, value string) (domain.ID, error) {
	value = cleanText(value)
	if value == "" {
		return "", invalidOptions("%s is required", name)
	}
	return domain.ID(value), nil
}

func validateCountryCode(value string) (string, error) {
	code := strings.ToUpper(cleanText(value))
	if len(code) != 2 {
		return "", invalidOptions("country code must be two letters")
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return "", invalidOptions("country code must contain letters only")
		}
	}
	return code, nil
}

func yearBounds(year int) (time.Time, time.Time) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	return start, end
}

func newID(prefix string) (domain.ID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return domain.ID(prefix + "_" + hex.EncodeToString(b[:])), nil
}
