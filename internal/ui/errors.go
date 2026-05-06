package ui

import (
	"errors"
	"fmt"
	"strings"
)

var ErrMissingPrerequisite = errors.New("missing prerequisite")

func required(name string) error {
	return fmt.Errorf("%s is required", name)
}

func prerequisite(message string) error {
	return fmt.Errorf("%s: %w", message, ErrMissingPrerequisite)
}

func cleanText(value string) string {
	return strings.TrimSpace(value)
}
