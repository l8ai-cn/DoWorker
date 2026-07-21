package knowledgebase

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceSecretsRoundTrip(t *testing.T) {
	svc := &Service{secrets: crypto.NewEncryptor("test-key")}
	raw := json.RawMessage(`{"app_id":"cli_123","app_secret":"s3cret","space_id":"7"}`)

	encrypted, err := svc.encryptSourceSecrets(raw)
	require.NoError(t, err)

	var cfg map[string]string
	require.NoError(t, json.Unmarshal(encrypted, &cfg))
	assert.Equal(t, "cli_123", cfg["app_id"])
	assert.Equal(t, "7", cfg["space_id"])
	assert.True(t, strings.HasPrefix(cfg["app_secret"], "enc:v1:"))
	assert.NotContains(t, cfg["app_secret"], "s3cret")

	// re-encrypt is idempotent
	again, err := svc.encryptSourceSecrets(encrypted)
	require.NoError(t, err)
	assert.JSONEq(t, string(encrypted), string(again))

	decrypted, err := svc.decryptSourceSecrets(encrypted)
	require.NoError(t, err)
	var out map[string]string
	require.NoError(t, json.Unmarshal(decrypted, &out))
	assert.Equal(t, "s3cret", out["app_secret"])
}

func TestSourceSecretsNilEncryptorPassthrough(t *testing.T) {
	svc := &Service{}
	raw := json.RawMessage(`{"app_secret":"plain"}`)

	out, err := svc.encryptSourceSecrets(raw)
	require.NoError(t, err)
	assert.Equal(t, raw, out)
}

func TestRedactSourceSecrets(t *testing.T) {
	raw := json.RawMessage(`{"app_id":"a","app_secret":"enc:v1:abc","access_token":"tok"}`)
	out := RedactedSourceConfigJSON(raw)
	assert.Contains(t, out, `"app_secret":"***"`)
	assert.Contains(t, out, `"access_token":"***"`)
	assert.Contains(t, out, `"app_id":"a"`)
}

func TestMergeSourceConfigUpdatePreservesSecrets(t *testing.T) {
	svc := &Service{}
	existing := json.RawMessage(`{"app_id":"a","app_secret":"enc:v1:abc","space_id":"sp"}`)
	incoming := json.RawMessage(`{"app_id":"a2","app_secret":"","space_id":"sp2"}`)

	merged, err := svc.mergeSourceConfigUpdate(existing, incoming)
	require.NoError(t, err)

	var cfg map[string]string
	require.NoError(t, json.Unmarshal(merged, &cfg))
	assert.Equal(t, "a2", cfg["app_id"])
	assert.Equal(t, "enc:v1:abc", cfg["app_secret"])
	assert.Equal(t, "sp2", cfg["space_id"])
}
