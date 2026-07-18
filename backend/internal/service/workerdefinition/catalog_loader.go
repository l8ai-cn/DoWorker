package workerdefinition

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Catalog struct {
	definitions map[string]Definition
	slugs       []string
}

type Definition struct {
	Slug                        string
	Version                     string
	Executable                  string
	AdapterID                   string
	Internal                    bool
	DefinitionHash              string
	DefinitionSource            []byte
	AgentFile                   string
	Modes                       []string
	ModelRequirement            ModelRequirement
	ToolModelRequirements       []ToolModelRequirement
	CredentialBindings          []CredentialBinding
	CredentialRequirementGroups []CredentialRequirementGroup
	ConfigDocuments             []ConfigDocument
	Image                       Image
}

type ModelRequirement struct {
	Required         bool
	ProtocolAdapters []string
}

type ToolModelRequirement struct {
	ID               string
	ProviderKeys     []string
	ProtocolAdapters []string
	Modality         string
	Capability       string
	Environment      ToolModelEnvironment
}

type ToolModelEnvironment struct {
	APIKey  string
	BaseURL string
	ModelID string
}

type CredentialBinding struct {
	ID     string
	Source CredentialSource
	Target CredentialTarget
}

type CredentialSource struct {
	Kind string
	Ref  string
}

type CredentialTarget struct {
	Kind string
	Name string
}

type ConfigDocument struct {
	ID         string
	Format     string
	TargetPath string
	Required   bool
}

type Image struct {
	Runtime      string
	VersionProbe []string
}

type catalogFile struct {
	SchemaVersion int             `json:"schema_version"`
	WorkerTypes   []catalogWorker `json:"worker_types"`
}

type catalogWorker struct {
	Slug           string `json:"slug"`
	DefinitionPath string `json:"definition_path"`
	DefinitionHash string `json:"definition_hash"`
}

type definitionFile struct {
	SchemaVersion               int               `json:"schema_version"`
	Slug                        string            `json:"slug"`
	DefinitionVersion           string            `json:"definition_version"`
	Executable                  string            `json:"executable"`
	AdapterID                   string            `json:"adapter_id"`
	InteractionModes            []string          `json:"interaction_modes"`
	Internal                    bool              `json:"internal"`
	ModelRequirement            json.RawMessage   `json:"model_requirement"`
	ToolModelRequirements       []json.RawMessage `json:"tool_model_requirements"`
	CredentialBindings          []json.RawMessage `json:"credential_bindings"`
	CredentialRequirementGroups []json.RawMessage `json:"credential_requirement_groups"`
	ConfigDocuments             []json.RawMessage `json:"config_documents"`
	Image                       json.RawMessage   `json:"image"`
}

func Load(root string) (*Catalog, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve worker definitions path: %w", err)
	}
	if err := validateSchema(root); err != nil {
		return nil, err
	}
	rawCatalog, err := os.ReadFile(filepath.Join(root, "catalog.json"))
	if err != nil {
		return nil, fmt.Errorf("read worker definition catalog: %w", err)
	}
	var document catalogFile
	if err := json.Unmarshal(rawCatalog, &document); err != nil {
		return nil, fmt.Errorf("decode worker definition catalog: %w", err)
	}
	if document.SchemaVersion != 1 {
		return nil, fmt.Errorf("worker definition catalog schema_version must be 1")
	}
	if err := validateCatalogEntries(document.WorkerTypes); err != nil {
		return nil, err
	}
	repoRoot := filepath.Dir(filepath.Dir(root))
	definitions := make(map[string]Definition, len(document.WorkerTypes))
	slugs := make([]string, 0, len(document.WorkerTypes))
	for _, entry := range document.WorkerTypes {
		definition, err := loadDefinition(repoRoot, entry)
		if err != nil {
			return nil, err
		}
		definitions[definition.Slug] = definition
		slugs = append(slugs, definition.Slug)
	}
	return &Catalog{definitions: definitions, slugs: slugs}, nil
}

func (catalog *Catalog) Get(slug string) (Definition, bool) {
	if catalog == nil {
		return Definition{}, false
	}
	definition, ok := catalog.definitions[slug]
	return cloneDefinition(definition), ok
}

func (catalog *Catalog) Slugs() []string {
	if catalog == nil {
		return nil
	}
	return append([]string{}, catalog.slugs...)
}

func (catalog *Catalog) CredentialBundleFields(slug string) ([]string, bool) {
	definition, found := catalog.Get(slug)
	if !found {
		return nil, false
	}
	policy := BuildEnvironmentBundlePolicy(definition)
	return policy.CredentialBundleFields, true
}

func cloneDefinition(definition Definition) Definition {
	definition.DefinitionSource = append([]byte{}, definition.DefinitionSource...)
	definition.Modes = append([]string{}, definition.Modes...)
	definition.ModelRequirement.ProtocolAdapters = append(
		[]string{},
		definition.ModelRequirement.ProtocolAdapters...,
	)
	definition.ToolModelRequirements = append(
		[]ToolModelRequirement{},
		definition.ToolModelRequirements...,
	)
	for index := range definition.ToolModelRequirements {
		requirement := &definition.ToolModelRequirements[index]
		requirement.ProviderKeys = append([]string{}, requirement.ProviderKeys...)
		requirement.ProtocolAdapters = append([]string{}, requirement.ProtocolAdapters...)
	}
	definition.CredentialBindings = append(
		[]CredentialBinding{},
		definition.CredentialBindings...,
	)
	definition.CredentialRequirementGroups = cloneCredentialRequirementGroups(
		definition.CredentialRequirementGroups,
	)
	definition.ConfigDocuments = append([]ConfigDocument{}, definition.ConfigDocuments...)
	definition.Image.VersionProbe = append([]string{}, definition.Image.VersionProbe...)
	return definition
}
