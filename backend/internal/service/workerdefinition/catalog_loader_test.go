package workerdefinition

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"testing"

	"github.com/anthropics/agentsmesh/agentfile/extract"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	agentfileschema "github.com/anthropics/agentsmesh/agentfile/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLoadsEveryFormalWorkerDefinition(t *testing.T) {
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))

	require.NoError(t, err)
	assert.Equal(t, formalWorkerSlugs, catalog.Slugs())
	for _, slug := range catalog.Slugs() {
		definition, ok := catalog.Get(slug)
		require.True(t, ok)
		assert.NotEmpty(t, definition.DefinitionHash)
		assert.NotEmpty(t, definition.AgentFile)
		assert.NotEmpty(t, definition.AdapterID)
		assert.NotEmpty(t, definition.Image.Runtime)
		assert.NotEmpty(t, definition.Image.VersionProbe)
		program, parseErrors := parser.Parse(definition.AgentFile)
		require.Empty(t, parseErrors, slug)
		assert.Equal(t, definition.Executable, extract.Extract(program).Agent.Command, slug)
		if definition.ModelRequirement.Required {
			assert.NotEmpty(t, definition.ModelRequirement.ProtocolAdapters)
		} else {
			assert.Empty(t, definition.ModelRequirement.ProtocolAdapters)
		}
		for _, binding := range definition.CredentialBindings {
			assert.NotEmpty(t, binding.ID)
			assert.NotEmpty(t, binding.Source.Ref)
			assert.Equal(t, "env", binding.Target.Kind)
		}
	}
}

func TestCursorDefinitionDoesNotRequirePlatformModelResource(t *testing.T) {
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))

	require.NoError(t, err)
	cursor, ok := catalog.Get("cursor-cli")
	require.True(t, ok)
	assert.False(t, cursor.ModelRequirement.Required)
	assert.Empty(t, cursor.ModelRequirement.ProtocolAdapters)
}

func TestGeminiDefinitionUsesGeminiAPIKey(t *testing.T) {
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))

	require.NoError(t, err)
	gemini, ok := catalog.Get("gemini-cli")
	require.True(t, ok)
	require.Len(t, gemini.CredentialBindings, 1)
	assert.Equal(t, "GEMINI_API_KEY", gemini.CredentialBindings[0].Target.Name)
	assert.Contains(t, gemini.AgentFile, "ENV GEMINI_API_KEY SECRET OPTIONAL")
}

func TestSeedanceDefinitionRequiresDoubaoVideoModel(t *testing.T) {
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))

	require.NoError(t, err)
	seedance, ok := catalog.Get("seedance-expert")
	require.True(t, ok)
	require.Len(t, seedance.ToolModelRequirements, 1)
	requirement := seedance.ToolModelRequirements[0]
	assert.Equal(t, "seedance-video", requirement.ID)
	assert.Equal(t, []string{"doubao"}, requirement.ProviderKeys)
	assert.Equal(t, []string{"openai-compatible"}, requirement.ProtocolAdapters)
	assert.Equal(t, "video", requirement.Modality)
	assert.Equal(t, "video-generation", requirement.Capability)
	assert.Equal(t, "SEEDANCE_API_KEY", requirement.Environment.APIKey)
	assert.Equal(t, "SEEDANCE_BASE_URL", requirement.Environment.BaseURL)
	assert.Equal(t, "SEEDANCE_MODEL", requirement.Environment.ModelID)
	assert.Contains(t, seedance.AgentFile, `ENV DO_AGENT_HOME = sandbox.root + "/do-agent-home"`)
	assert.NotContains(t, seedance.AgentFile, "seedance-expert-home")
}

func TestMiniMaxDefinitionUsesOneShotChatCommand(t *testing.T) {
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))

	require.NoError(t, err)
	minimax, ok := catalog.Get("minimax-cli")
	require.True(t, ok)

	assert.Contains(t, minimax.AgentFile, "PROMPT_POSITION append")
	assert.Contains(t, minimax.AgentFile, "arg \"text\"")
	assert.Contains(t, minimax.AgentFile, "arg \"chat\"")
	assert.Contains(t, minimax.AgentFile, "arg \"--non-interactive\"")
	assert.Contains(t, minimax.AgentFile, "arg \"--message\"")
	assert.NotContains(t, minimax.AgentFile, "arg \"repl\"")
}

func TestCatalogSeparatesCredentialBundleFieldsFromModelResourceFields(t *testing.T) {
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))

	require.NoError(t, err)
	cursorFields, cursorFound := catalog.CredentialBundleFields("cursor-cli")
	claudeFields, claudeFound := catalog.CredentialBundleFields("claude-code")
	_, unknownFound := catalog.CredentialBundleFields("unknown")

	assert.True(t, cursorFound)
	assert.Equal(t, []string{"CURSOR_API_KEY"}, cursorFields)
	assert.True(t, claudeFound)
	assert.Empty(t, claudeFields)
	assert.False(t, unknownFound)
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	require.NoError(t, err)
	return root
}

func TestCatalogSlugsAreSorted(t *testing.T) {
	sorted := append([]string{}, formalWorkerSlugs...)
	sort.Strings(sorted)
	assert.Equal(t, sorted, formalWorkerSlugs)
}

func TestValidateDefinitionRejectsLiteralCredentialValue(t *testing.T) {
	document := definitionFile{
		SchemaVersion:     1,
		Slug:              "codex-cli",
		DefinitionVersion: "1",
		Executable:        "codex",
		AdapterID:         "codex-app-server",
		InteractionModes:  []string{"pty"},
		ModelRequirement:  json.RawMessage(`{"required":true,"protocol_adapters":["openai-compatible"]}`),
		CredentialBindings: []json.RawMessage{
			json.RawMessage(`{"id":"openai","source":{"kind":"model_resource","ref":"openai-compatible","value":"secret"},"target":{"kind":"env","name":"OPENAI_API_KEY"}}`),
		},
		ConfigDocuments: []json.RawMessage{},
		Image:           json.RawMessage(`{"runtime":"docker"}`),
	}

	err := validateDefinition("codex-cli", document)

	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid credential binding")
}

func TestValidateDefinitionRejectsInvalidAdapterID(t *testing.T) {
	document := definitionFile{
		SchemaVersion:      1,
		Slug:               "codex-cli",
		DefinitionVersion:  "1",
		Executable:         "codex",
		AdapterID:          "Codex.App",
		InteractionModes:   []string{"pty"},
		ModelRequirement:   json.RawMessage(`{"required":true,"protocol_adapters":["openai-compatible"]}`),
		CredentialBindings: []json.RawMessage{},
		ConfigDocuments:    []json.RawMessage{},
		Image:              json.RawMessage(`{"runtime":"codex-cli","version_probe":["codex","--version"]}`),
	}

	err := validateDefinition("codex-cli", document)

	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid adapter_id")
}

func TestValidateCredentialBindingSchemaRejectsUnboundAgentFileField(t *testing.T) {
	schema, err := agentfileschema.FromSource("ENV CURSOR_API_KEY SECRET\n")
	require.NoError(t, err)

	err = validateCredentialBindingSchema(ModelRequirement{}, schema, nil, nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, "has no binding")
}

func TestDefinitionBundleHashIncludesAgentFile(t *testing.T) {
	definition := []byte(`{"slug":"codex-cli"}`)

	assert.NotEqual(
		t,
		definitionBundleHash(definition, []byte("AGENT codex\n")),
		definitionBundleHash(definition, []byte("AGENT codex\nMODE pty\n")),
	)
}
