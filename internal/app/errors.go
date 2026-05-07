package app

import (
	"errors"
	"fmt"
)

var ErrInvalidOptions = errors.New("invalid app options")

func invalidOptions(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrInvalidOptions)...)
}
