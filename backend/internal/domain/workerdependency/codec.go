package workerdependency

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

const MaxDocumentBytes = 1 << 20

var (
	ErrUnsupportedVersion = errors.New("worker dependency version is unsupported")
	ErrDocumentTooLarge   = errors.New("worker dependency document is too large")
)

func Encode(document Document) ([]byte, error) {
	if err := validateDocumentSizeBudget(document); err != nil {
		return nil, err
	}
	normalized, err := NormalizeAndValidate(document)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	encoded, err := control.CanonicalJSONObject(raw)
	if err != nil {
		return nil, fmt.Errorf("canonicalize worker dependency document: %w", err)
	}
	if len(encoded) > MaxDocumentBytes {
		return nil, ErrDocumentTooLarge
	}
	return encoded, nil
}

func EncodeAndDigest(document Document) ([]byte, string, error) {
	encoded, err := Encode(document)
	if err != nil {
		return nil, "", err
	}
	return encoded, digestBytes(encoded), nil
}

func Decode(data []byte) (Document, error) {
	if len(data) > MaxDocumentBytes {
		return Document{}, ErrDocumentTooLarge
	}
	if !utf8.Valid(data) {
		return Document{}, fmt.Errorf("worker dependency document must be UTF-8")
	}
	canonical, err := control.CanonicalJSONObject(data)
	if err != nil {
		return Document{}, fmt.Errorf("worker dependency document structure: %w", err)
	}
	if err := requireV1(canonical); err != nil {
		return Document{}, err
	}
	if err := requireDocumentFields(canonical); err != nil {
		return Document{}, err
	}
	var document Document
	if err := decodeStrict(canonical, &document); err != nil {
		return Document{}, fmt.Errorf("decode worker dependencies: %w", err)
	}
	if err := requireCollections(document); err != nil {
		return Document{}, err
	}
	return NormalizeAndValidate(document)
}

func Digest(document Document) (string, error) {
	_, digest, err := EncodeAndDigest(document)
	if err != nil {
		return "", err
	}
	return digest, nil
}

func DigestRuntimeValues(values []RuntimeValue) (string, error) {
	normalized := normalizeRuntimeValues(values)
	if err := validateRuntimeValues(normalized); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return digestBytes(encoded), nil
}

func TextDigest(value string) string {
	return digestBytes([]byte(value))
}

func requireV1(data []byte) error {
	var envelope struct {
		Version Version `json:"version"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode worker dependency version: %w", err)
	}
	if envelope.Version != VersionV1 {
		return fmt.Errorf("%w: %d", ErrUnsupportedVersion, envelope.Version)
	}
	return nil
}

func requireDocumentFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := decodeStrict(data, &fields); err != nil {
		return fmt.Errorf("decode worker dependency envelope: %w", err)
	}
	required := []string{
		"version",
		"organization_id",
		"namespace",
		"worker",
		"models",
		"repository",
		"skills",
		"knowledge_bases",
		"runtime_bundles",
		"secret_refs",
		"placement",
	}
	for _, field := range required {
		if _, exists := fields[field]; !exists {
			return fmt.Errorf("worker dependency field %q is required", field)
		}
	}
	return nil
}

func requireCollections(document Document) error {
	if document.Models.Tools == nil ||
		document.Worker.ModelManagedFields == nil ||
		document.Worker.CredentialBundleFields == nil ||
		document.Skills == nil ||
		document.KnowledgeBases == nil ||
		document.RuntimeBundles == nil ||
		document.SecretReferences == nil {
		return fmt.Errorf("worker dependency collections must be arrays")
	}
	for _, bundle := range document.RuntimeBundles {
		if bundle.Values == nil {
			return fmt.Errorf("worker dependency bundle values must be an array")
		}
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

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return fmt.Sprintf("sha256:%x", sum)
}
