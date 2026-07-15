package orchestrationcontrol

import (
	"errors"
	"fmt"
)

var (
	ErrInvalid  = errors.New("orchestration control invalid")
	ErrNotFound = errors.New("orchestration control not found")
	ErrConflict = errors.New("orchestration control conflict")
	ErrStale    = errors.New("orchestration control stale")
	ErrExpired  = errors.New("orchestration control expired")
	ErrConsumed = errors.New("orchestration control consumed")
	ErrCorrupt  = errors.New("orchestration control corrupt")
)

func invalid(field, requirement string) error {
	return fmt.Errorf("%w: %s %s", ErrInvalid, field, requirement)
}

func corrupt(field, requirement string) error {
	return fmt.Errorf("%w: %s %s", ErrCorrupt, field, requirement)
}
