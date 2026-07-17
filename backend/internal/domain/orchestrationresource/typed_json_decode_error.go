package orchestrationresource

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const maxSanitizedTypedJSONErrorBytes = 480

var (
	ErrTypedJSONSyntax        = errors.New("typed JSON syntax error")
	ErrTypedJSONType          = errors.New("typed JSON type error")
	ErrTypedJSONUnknownField  = errors.New("typed JSON unknown field")
	ErrTypedJSONStringTag     = errors.New("typed JSON string-tag error")
	ErrTypedJSONInvalidTarget = errors.New("typed JSON invalid target")
	ErrTypedJSONDecode        = errors.New("typed JSON decode error")
)

type safeTypedJSONError struct {
	kind    error
	message string
}

func (err *safeTypedJSONError) Error() string {
	return err.message
}

func (err *safeTypedJSONError) Unwrap() error {
	return err.kind
}

func sanitizeTypedJSONDecodeError(err error) error {
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return boundedTypedJSONError(ErrTypedJSONSyntax, fmt.Sprintf(
			"typed JSON syntax error at offset %d",
			syntaxError.Offset,
		))
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return boundedTypedJSONError(
			ErrTypedJSONSyntax,
			"typed JSON syntax error: unexpected end of input",
		)
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		return sanitizeUnmarshalTypeError(typeError)
	}
	var invalidTarget *json.InvalidUnmarshalError
	if errors.As(err, &invalidTarget) {
		return sanitizeInvalidUnmarshalError(invalidTarget)
	}

	message := err.Error()
	if strings.HasPrefix(message, "json: unknown field ") {
		return boundedTypedJSONError(ErrTypedJSONUnknownField, "typed JSON unknown field")
	}
	if isStringTagError(message) {
		return boundedTypedJSONError(ErrTypedJSONStringTag, "typed JSON string-tag error")
	}
	return boundedTypedJSONError(ErrTypedJSONDecode, "typed JSON decode error")
}

func sanitizeUnmarshalTypeError(err *json.UnmarshalTypeError) error {
	typeName := "<nil>"
	if err.Type != nil {
		typeName = err.Type.String()
	}
	message := fmt.Sprintf(
		"typed JSON type error: %s cannot decode into type %s",
		safeJSONValueKind(err.Value),
		summarizeValue(typeName),
	)
	if err.Field != "" {
		message += " at field " + summarizeValue(err.Field)
	}
	if err.Struct != "" {
		message += " in struct " + summarizeValue(err.Struct)
	}
	message += fmt.Sprintf(" at offset %d", err.Offset)
	return boundedTypedJSONError(ErrTypedJSONType, message)
}

func sanitizeInvalidUnmarshalError(err *json.InvalidUnmarshalError) error {
	typeName := "<nil>"
	if err.Type != nil {
		typeName = err.Type.String()
	}
	return boundedTypedJSONError(
		ErrTypedJSONInvalidTarget,
		"typed JSON internal target error: invalid target type "+summarizeValue(typeName),
	)
}

func safeJSONValueKind(value string) string {
	kind, _, _ := strings.Cut(value, " ")
	switch kind {
	case "number", "string", "object", "array", "bool", "null":
		return kind
	default:
		return "value"
	}
}

func isStringTagError(message string) bool {
	return strings.HasPrefix(message, "json: invalid use of ,string struct tag") ||
		strings.HasPrefix(message, "json: invalid number literal, trying to unmarshal")
}

func boundedTypedJSONError(kind error, message string) error {
	valid := strings.ToValidUTF8(message, "\uFFFD")
	if len(valid) <= maxSanitizedTypedJSONErrorBytes {
		return &safeTypedJSONError{kind: kind, message: valid}
	}
	const suffix = "..."
	limit := maxSanitizedTypedJSONErrorBytes - len(suffix)
	for limit > 0 && !utf8.ValidString(valid[:limit]) {
		limit--
	}
	return &safeTypedJSONError{kind: kind, message: valid[:limit] + suffix}
}
