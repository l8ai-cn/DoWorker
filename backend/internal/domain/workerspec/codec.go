package workerspec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func EncodeSpec(spec Spec) ([]byte, error) {
	normalized, err := NormalizeAndValidate(spec)
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

func DecodeSpec(data []byte) (Spec, error) {
	if err := requireV1(data); err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err := decodeStrict(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("decode workerspec: %w", err)
	}
	return NormalizeAndValidate(spec)
}

func EncodeSummary(summary Summary) ([]byte, error) {
	if err := ValidateSummary(summary); err != nil {
		return nil, err
	}
	return json.Marshal(summary)
}

func DecodeSummary(data []byte) (Summary, error) {
	if err := requireV1(data); err != nil {
		return Summary{}, err
	}
	var summary Summary
	if err := decodeStrict(data, &summary); err != nil {
		return Summary{}, fmt.Errorf("decode workerspec summary: %w", err)
	}
	if err := ValidateSummary(summary); err != nil {
		return Summary{}, err
	}
	return summary, nil
}

func requireV1(data []byte) error {
	var envelope struct {
		Version Version `json:"version"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode workerspec version: %w", err)
	}
	if envelope.Version != VersionV1 {
		return fmt.Errorf("%w: %d", ErrUnsupportedVersion, envelope.Version)
	}
	return nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	err := decoder.Decode(&trailing)
	switch {
	case errors.Is(err, io.EOF):
		return nil
	case err == nil:
		return fmt.Errorf("trailing JSON data")
	default:
		return err
	}
}
