package workerdefinition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	agentfileschema "github.com/anthropics/agentsmesh/agentfile/schema"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func loadDefinition(repoRoot string, entry catalogWorker) (Definition, error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(entry.DefinitionPath))
	raw, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, fmt.Errorf("read Worker definition %q: %w", entry.Slug, err)
	}
	var document definitionFile
	if err := json.Unmarshal(raw, &document); err != nil {
		return Definition{}, fmt.Errorf("decode Worker definition %q: %w", entry.Slug, err)
	}
	if err := validateDefinition(entry.Slug, document); err != nil {
		return Definition{}, err
	}
	agentFile, err := os.ReadFile(filepath.Join(filepath.Dir(path), "AgentFile"))
	if err != nil {
		return Definition{}, fmt.Errorf("read Worker AgentFile %q: %w", entry.Slug, err)
	}
	hash := definitionBundleHash(raw, agentFile)
	if entry.DefinitionHash != "sha256:"+hash {
		return Definition{}, fmt.Errorf("Worker definition %q hash does not match catalog", entry.Slug)
	}
	agentSchema, err := agentfileschema.FromSource(string(agentFile))
	if err != nil {
		return Definition{}, fmt.Errorf("parse Worker AgentFile %q: %w", entry.Slug, err)
	}
	bindings, err := decodeCredentialBindings(document.CredentialBindings)
	if err != nil {
		return Definition{}, fmt.Errorf("decode Worker credentials %q: %w", entry.Slug, err)
	}
	modelRequirement, err := decodeModelRequirement(document.ModelRequirement)
	if err != nil {
		return Definition{}, fmt.Errorf(
			"decode Worker model requirement %q: %w",
			entry.Slug,
			err,
		)
	}
	toolModelRequirements, err := decodeToolModelRequirements(document.ToolModelRequirements)
	if err != nil {
		return Definition{}, fmt.Errorf(
			"decode Worker tool model requirements %q: %w",
			entry.Slug,
			err,
		)
	}
	if err := validateCredentialBindingSchema(
		modelRequirement,
		agentSchema,
		bindings,
		toolModelRequirements,
	); err != nil {
		return Definition{}, fmt.Errorf(
			"validate Worker credential bindings %q: %w",
			entry.Slug,
			err,
		)
	}
	configDocuments, err := decodeConfigDocuments(document.ConfigDocuments)
	if err != nil {
		return Definition{}, fmt.Errorf("decode Worker config documents %q: %w", entry.Slug, err)
	}
	image, err := decodeImage(document.Image)
	if err != nil {
		return Definition{}, fmt.Errorf("decode Worker image %q: %w", entry.Slug, err)
	}
	return Definition{
		Slug: entry.Slug, Version: document.DefinitionVersion,
		Executable: document.Executable, AdapterID: document.AdapterID,
		DefinitionHash: hash, AgentFile: string(agentFile),
		Modes:                 append([]string{}, document.InteractionModes...),
		ModelRequirement:      modelRequirement,
		ToolModelRequirements: toolModelRequirements,
		CredentialBindings:    bindings,
		ConfigDocuments:       configDocuments,
		Image:                 image,
	}, nil
}

func validateDefinition(slug string, document definitionFile) error {
	if document.SchemaVersion != 1 || document.Slug != slug ||
		document.DefinitionVersion == "" || document.Executable == "" ||
		document.AdapterID == "" || len(document.InteractionModes) == 0 ||
		len(document.ModelRequirement) == 0 ||
		document.CredentialBindings == nil || document.ConfigDocuments == nil ||
		len(document.Image) == 0 {
		return fmt.Errorf("Worker definition %q is incomplete", slug)
	}
	if err := slugkit.Validate(document.AdapterID); err != nil {
		return fmt.Errorf("Worker definition %q has invalid adapter_id: %w", slug, err)
	}
	for _, mode := range document.InteractionModes {
		if mode != "pty" && mode != "acp" {
			return fmt.Errorf("Worker definition %q has invalid interaction mode %q", slug, mode)
		}
	}
	if err := validateCredentialBindings(document.CredentialBindings); err != nil {
		return fmt.Errorf("Worker definition %q has invalid credential binding: %w", slug, err)
	}
	if _, err := decodeModelRequirement(document.ModelRequirement); err != nil {
		return fmt.Errorf("Worker definition %q has invalid model requirement: %w", slug, err)
	}
	if _, err := decodeToolModelRequirements(document.ToolModelRequirements); err != nil {
		return fmt.Errorf("Worker definition %q has invalid tool model requirement: %w", slug, err)
	}
	if err := validateConfigDocuments(document.ConfigDocuments); err != nil {
		return fmt.Errorf("Worker definition %q has invalid config document: %w", slug, err)
	}
	return nil
}

type configDocumentDocument struct {
	ID         string `json:"id"`
	Format     string `json:"format"`
	TargetPath string `json:"target_path"`
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
		if document.ID == "" || document.Format == "" || document.TargetPath == "" {
			return nil, fmt.Errorf("config document must declare id, format, and target_path")
		}
		documents = append(documents, ConfigDocument{
			ID: document.ID, Format: document.Format, TargetPath: document.TargetPath,
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

func definitionBundleHash(definition, agentFile []byte) string {
	hash := sha256.New()
	_, _ = hash.Write(definition)
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(agentFile)
	return hex.EncodeToString(hash.Sum(nil))
}
