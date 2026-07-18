package workerdefinition

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSnapshotRejectsNonJSONConfigDocument(t *testing.T) {
	source := []byte(
		`{"schema_version":1,"slug":"codex-cli","definition_version":"1",` +
			`"executable":"codex","adapter_id":"codex-app-server",` +
			`"interaction_modes":["pty"],` +
			`"model_requirement":{"required":false,"protocol_adapters":[]},` +
			`"credential_bindings":[],"config_documents":[` +
			`{"id":"settings","format":"yaml","target_path":"settings.yaml"}],` +
			`"image":{"runtime":"codex-cli","version_probe":["codex","--version"]}}`,
	)

	_, err := ParseSnapshot(source, "AGENT codex\nMODE pty\n")

	require.ErrorContains(t, err, "config document must declare")
}

func TestParseSnapshotRejectsCredentialGroupWithUnknownTarget(t *testing.T) {
	source := []byte(
		`{"schema_version":1,"slug":"codex-cli","definition_version":"1",` +
			`"executable":"codex","adapter_id":"codex-app-server",` +
			`"interaction_modes":["pty"],` +
			`"model_requirement":{"required":false,"protocol_adapters":[]},` +
			`"credential_bindings":[{"id":"openai","source":{"kind":"credential_bundle","ref":"codex-cli"},"target":{"kind":"env","name":"OPENAI_API_KEY"}}],` +
			`"credential_requirement_groups":[{"id":"provider-api-key","any_of":["OPENAI_API_KEY","ANTHROPIC_API_KEY"]}],` +
			`"config_documents":[],"image":{"runtime":"codex-cli","version_probe":["codex","--version"]}}`,
	)

	_, err := ParseSnapshot(source, "AGENT codex\nENV OPENAI_API_KEY SECRET OPTIONAL\n")

	require.ErrorContains(t, err, `references undeclared target "ANTHROPIC_API_KEY"`)
}
