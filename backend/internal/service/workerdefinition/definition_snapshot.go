package workerdefinition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	agentfileschema "github.com/anthropics/agentsmesh/agentfile/schema"
)

func ParseSnapshot(source []byte, agentFile string) (Definition, error) {
	var document definitionFile
	if err := json.Unmarshal(source, &document); err != nil {
		return Definition{}, err
	}
	if err := validateDefinition(document.Slug, document); err != nil {
		return Definition{}, err
	}
	schema, err := agentfileschema.FromSource(agentFile)
	if err != nil {
		return Definition{}, fmt.Errorf("parse Worker AgentFile: %w", err)
	}
	bindings, err := decodeCredentialBindings(document.CredentialBindings)
	if err != nil {
		return Definition{}, err
	}
	credentialGroups, err := decodeCredentialRequirementGroups(
		document.CredentialRequirementGroups,
	)
	if err != nil {
		return Definition{}, err
	}
	if err := validateCredentialRequirementGroupTargets(credentialGroups, bindings); err != nil {
		return Definition{}, err
	}
	model, err := decodeModelRequirement(document.ModelRequirement)
	if err != nil {
		return Definition{}, err
	}
	tools, err := decodeToolModelRequirements(document.ToolModelRequirements)
	if err != nil {
		return Definition{}, err
	}
	if err := validateCredentialBindingSchema(model, schema, bindings, tools); err != nil {
		return Definition{}, err
	}
	configDocuments, err := decodeConfigDocuments(document.ConfigDocuments)
	if err != nil {
		return Definition{}, err
	}
	image, err := decodeImage(document.Image)
	if err != nil {
		return Definition{}, err
	}
	return Definition{
		Slug: document.Slug, Version: document.DefinitionVersion,
		Internal:   document.Internal,
		Executable: document.Executable, AdapterID: document.AdapterID,
		DefinitionHash:   definitionBundleHash(source, []byte(agentFile)),
		DefinitionSource: append([]byte{}, source...), AgentFile: agentFile,
		Modes:                       append([]string{}, document.InteractionModes...),
		ModelRequirement:            model,
		ToolModelRequirements:       tools,
		CredentialBindings:          bindings,
		CredentialRequirementGroups: credentialGroups,
		ConfigDocuments:             configDocuments,
		Image:                       image,
	}, nil
}

func ValidateIntegrity(definition Definition) error {
	reparsed, err := ParseSnapshot(
		definition.DefinitionSource,
		definition.AgentFile,
	)
	if err != nil {
		return fmt.Errorf("worker definition snapshot is invalid: %w", err)
	}
	if !reflect.DeepEqual(reparsed, definition) {
		return fmt.Errorf("worker definition snapshot projection does not match its source")
	}
	return nil
}

func definitionBundleHash(definition, agentFile []byte) string {
	hash := sha256.New()
	_, _ = hash.Write(definition)
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(agentFile)
	return hex.EncodeToString(hash.Sum(nil))
}
