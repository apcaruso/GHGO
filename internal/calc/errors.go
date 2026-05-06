package calc

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidOptions    = errors.New("invalid calculation options")
	ErrNoActiveRecords   = errors.New("no active activity records")
	ErrInvalidSettings   = errors.New("invalid calculation settings")
	ErrUnsupportedRecord = errors.New("unsupported activity record")
	ErrMissingFactor     = errors.New("missing emission factor")
)

func invalidOptions(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrInvalidOptions)...)
}

func invalidSettings(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrInvalidSettings)...)
}

func unsupportedRecord(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrUnsupportedRecord)...)
}
