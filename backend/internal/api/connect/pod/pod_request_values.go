package podconnect

import (
	"errors"
	"strings"

	"connectrpc.com/connect"
)

func normalizeAlias(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validateAlias(value *string) error {
	if value != nil && len(*value) > 100 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("alias must be 100 characters or less"))
	}
	return nil
}

func optionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func optionalInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func optionalBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
