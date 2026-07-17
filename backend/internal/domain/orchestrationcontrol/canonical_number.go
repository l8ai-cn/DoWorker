package orchestrationcontrol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func normalizeJSONNumbers(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if err := normalizeJSONNumbers(child); err != nil {
				return err
			}
			if number, ok := child.(json.Number); ok {
				normalized, err := normalizeJSONNumber(number.String())
				if err != nil {
					return err
				}
				typed[key] = json.Number(normalized)
			}
		}
	case []any:
		for index, child := range typed {
			if err := normalizeJSONNumbers(child); err != nil {
				return err
			}
			if number, ok := child.(json.Number); ok {
				normalized, err := normalizeJSONNumber(number.String())
				if err != nil {
					return err
				}
				typed[index] = json.Number(normalized)
			}
		}
	}
	return nil
}

func normalizeJSONNumber(source string) (string, error) {
	sign := ""
	if strings.HasPrefix(source, "-") {
		sign, source = "-", source[1:]
	}
	mantissa, exponentText, _ := strings.Cut(source, "e")
	if exponentText == "" {
		mantissa, exponentText, _ = strings.Cut(source, "E")
	}
	exponent := 0
	var err error
	if exponentText != "" {
		exponent, err = strconv.Atoi(exponentText)
		if err != nil {
			return "", errors.New("number exponent is out of range")
		}
		if exponent > maxCanonicalJSONBytes || exponent < -maxCanonicalJSONBytes {
			return "", errors.New("number exponent is out of range")
		}
	}
	integer, fraction, hasFraction := strings.Cut(mantissa, ".")
	digits := integer + fraction
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return "0", nil
	}
	if hasFraction {
		exponent -= len(fraction)
	}
	for strings.HasSuffix(digits, "0") {
		digits = strings.TrimSuffix(digits, "0")
		exponent++
	}
	if exponent >= 0 {
		if len(digits)+exponent > maxCanonicalJSONBytes {
			return "", errors.New("number expansion is too large")
		}
		return sign + digits + strings.Repeat("0", exponent), nil
	}
	decimalPosition := len(digits) + exponent
	if decimalPosition > 0 {
		return sign + digits[:decimalPosition] + "." + digits[decimalPosition:], nil
	}
	zeroCount := -decimalPosition
	if len(digits)+zeroCount+2 > maxCanonicalJSONBytes {
		return "", errors.New("number exponent is out of range")
	}
	return sign + "0." + strings.Repeat("0", zeroCount) + digits, nil
}
