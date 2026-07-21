package workerdefinition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func loadDefinition(repoRoot string, entry catalogWorker) (Definition, error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(entry.DefinitionPath))
	raw, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, fmt.Errorf("read Worker definition %q: %w", entry.Slug, err)
	}
	agentFile, err := os.ReadFile(filepath.Join(filepath.Dir(path), "AgentFile"))
	if err != nil {
		return Definition{}, fmt.Errorf("read Worker AgentFile %q: %w", entry.Slug, err)
	}
	definition, err := ParseSnapshot(raw, string(agentFile))
	if err != nil {
		return Definition{}, fmt.Errorf("decode Worker definition %q: %w", entry.Slug, err)
	}
	if definition.Slug != entry.Slug {
		return Definition{}, fmt.Errorf("worker definition %q slug does not match catalog", entry.Slug)
	}
	if entry.DefinitionHash != "sha256:"+definition.DefinitionHash {
		return Definition{}, fmt.Errorf("worker definition %q hash does not match catalog", entry.Slug)
	}
	return definition, nil
}

func validateDefinition(slug string, document definitionFile) error {
	if document.SchemaVersion != 1 || document.Slug != slug ||
		document.DefinitionVersion == "" || document.Executable == "" ||
		document.AdapterID == "" || len(document.InteractionModes) == 0 ||
		len(document.ModelRequirement) == 0 ||
		document.CredentialBindings == nil || document.ConfigDocuments == nil ||
		len(document.Image) == 0 {
		return fmt.Errorf("worker definition %q is incomplete", slug)
	}
	if err := slugkit.Validate(document.AdapterID); err != nil {
		return fmt.Errorf("worker definition %q has invalid adapter_id: %w", slug, err)
	}
	for _, mode := range document.InteractionModes {
		if mode != "pty" && mode != "acp" {
			return fmt.Errorf("worker definition %q has invalid interaction mode %q", slug, mode)
		}
	}
	if err := validateCredentialBindings(document.CredentialBindings); err != nil {
		return fmt.Errorf("worker definition %q has invalid credential binding: %w", slug, err)
	}
	if err := validateCredentialRequirementGroups(
		document.CredentialRequirementGroups,
	); err != nil {
		return fmt.Errorf(
			"worker definition %q has invalid credential requirement group: %w",
			slug,
			err,
		)
	}
	if _, err := decodeModelRequirement(document.ModelRequirement); err != nil {
		return fmt.Errorf("worker definition %q has invalid model requirement: %w", slug, err)
	}
	if _, err := decodeToolModelRequirements(document.ToolModelRequirements); err != nil {
		return fmt.Errorf("worker definition %q has invalid tool model requirement: %w", slug, err)
	}
	if err := validateConfigDocuments(document.ConfigDocuments); err != nil {
		return fmt.Errorf("worker definition %q has invalid config document: %w", slug, err)
	}
	return nil
}

type configDocumentDocument struct {
	ID         string `json:"id"`
	Format     string `json:"format"`
	TargetPath string `json:"target_path"`
	Required   *bool  `json:"required"`
}

func validateConfigDocuments(documents []json.RawMessage) error {
	_, err := decodeConfigDocuments(documents)
	return err
}

func decodeConfigDocuments(rawDocuments []json.RawMessage) ([]ConfigDocument, error) {
	documents := make([]ConfigDocument, 0, len(rawDocuments))
	for _, raw := range rawDocuments {
		var document configDocumentDocument
		if err := decodeStrict(raw, &document); err != nil {
			return nil, err
		}
		if document.ID == "" || document.Format != "json" ||
			document.TargetPath == "" || document.Required == nil {
			return nil, fmt.Errorf(
				"config document must declare id, format, target_path, and required",
			)
		}
		documents = append(documents, ConfigDocument{
			ID:         document.ID,
			Format:     document.Format,
			TargetPath: document.TargetPath,
			Required:   *document.Required,
		})
	}
	return documents, nil
}

func decodeImage(raw json.RawMessage) (Image, error) {
	var image struct {
		Runtime      string   `json:"runtime"`
		VersionProbe []string `json:"version_probe"`
	}
	if err := decodeStrict(raw, &image); err != nil {
		return Image{}, err
	}
	if image.Runtime == "" || len(image.VersionProbe) == 0 {
		return Image{}, fmt.Errorf("image must declare runtime and version_probe")
	}
	for _, argument := range image.VersionProbe {
		if argument == "" {
			return Image{}, fmt.Errorf("image version_probe cannot contain empty arguments")
		}
	}
	return Image{
		Runtime: image.Runtime, VersionProbe: append([]string{}, image.VersionProbe...),
	}, nil
}

func decodeStrict(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
