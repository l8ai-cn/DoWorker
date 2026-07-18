package workbench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func decodeStrictJSON(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func validateArtifactDeclarationHeader(declaration artifactDeclaration) error {
	if declaration.SchemaVersion != artifactDeclarationSchema {
		return fmt.Errorf(
			"schema_version must be %q",
			artifactDeclarationSchema,
		)
	}
	if declaration.Revision == 0 {
		return fmt.Errorf("revision must be positive")
	}
	for field, value := range map[string]string{
		"artifact_id":               declaration.ArtifactID,
		"primary_representation_id": declaration.PrimaryRepresentationID,
	} {
		if len(value) < 2 || len(value) > 100 ||
			!artifactDeclarationIdentifier.MatchString(value) {
			return fmt.Errorf("%s must be a 2-100 character identifier", field)
		}
	}
	if !validDeclarationLabel(declaration.Role, 64) {
		return fmt.Errorf("role is invalid")
	}
	if !validDeclarationLabel(declaration.Producer.Namespace, 100) ||
		!validDeclarationLabel(declaration.Producer.Type, 100) {
		return fmt.Errorf("producer namespace and type are required")
	}
	if declaration.Producer.Namespace == "seedance" &&
		declaration.Producer.Type == "video.generate" &&
		!validDeclarationLabel(declaration.Producer.ID, 200) {
		return fmt.Errorf("seedance video.generate requires producer.id")
	}
	if declaration.Producer.ToolExecutionID != "" {
		return fmt.Errorf("producer.tool_execution_id is assigned by Runner")
	}
	if len(declaration.Representations) == 0 ||
		len(declaration.Representations) > 64 {
		return fmt.Errorf("representations must contain 1-64 entries")
	}
	return nil
}

func validDeclarationLabel(value string, maximum int) bool {
	return value != "" && len(value) <= maximum &&
		strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, "\r\n\t")
}
