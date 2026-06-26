package report

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidOptions     = errors.New("invalid report options")
	ErrMixedMobileMethods = errors.New("mixed mobile methods")
)

func invalidOptions(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrInvalidOptions)...)
}
