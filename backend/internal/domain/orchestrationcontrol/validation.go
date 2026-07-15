package orchestrationcontrol

import (
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

func validateBaseState(
	operation PlanOperation,
	baseUID string,
	baseResourceVersion int64,
) error {
	switch operation {
	case PlanOperationCreate:
		if baseUID != "" || baseResourceVersion != 0 {
			return invalid("baseState", "must be empty for create")
		}
	case PlanOperationUpdate:
		if err := validateUUID("baseUid", baseUID); err != nil {
			return err
		}
		if baseResourceVersion <= 0 {
			return invalid("baseResourceVersion", "must be positive for update")
		}
	default:
		return invalid("operation", "must be create or update")
	}
	return nil
}

func validatePlanTimes(createdAt, expiresAt time.Time) error {
	if createdAt.IsZero() || expiresAt.IsZero() {
		return invalid("plan timestamps", "must not be zero")
	}
	if !expiresAt.After(createdAt) {
		return invalid("plan.expiresAt", "must be after creation")
	}
	if _, err := createdAt.MarshalJSON(); err != nil {
		return invalid("plan.createdAt", "must be encodable")
	}
	if _, err := expiresAt.MarshalJSON(); err != nil {
		return invalid("plan.expiresAt", "must be encodable")
	}
	return nil
}

func validateTimeRange(field string, start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return invalid(field+" timestamps", "must not be zero")
	}
	if end.Before(start) {
		return invalid(field+" timestamps", "must be ordered")
	}
	if _, err := start.MarshalJSON(); err != nil {
		return invalid(field+".createdAt", "must be encodable")
	}
	if _, err := end.MarshalJSON(); err != nil {
		return invalid(field+".updatedAt", "must be encodable")
	}
	return nil
}

func validateJSONPointer(field, path string) error {
	if path == "" || path[0] != '/' || len(path) > 512 || !utf8.ValidString(path) {
		return invalid(field, "must be a stable JSON pointer")
	}
	for index := 0; index < len(path); index++ {
		if path[index] != '~' {
			continue
		}
		if index+1 >= len(path) || (path[index+1] != '0' && path[index+1] != '1') {
			return invalid(field, "must use valid JSON pointer escaping")
		}
		index++
	}
	for _, char := range path {
		if unicode.IsControl(char) || unicode.Is(unicode.Bidi_Control, char) {
			return invalid(field, "must not contain control characters")
		}
	}
	return nil
}

func validateSafeText(field, value string, maxRunes int, allowEmpty bool) error {
	if (!allowEmpty && value == "") || !utf8.ValidString(value) ||
		utf8.RuneCountInString(value) > maxRunes {
		return invalid(field, "must be bounded valid text")
	}
	for _, char := range value {
		if unicode.IsControl(char) || unicode.Is(unicode.Bidi_Control, char) {
			return invalid(field, "must not contain control characters")
		}
	}
	return nil
}

func fmtInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}
