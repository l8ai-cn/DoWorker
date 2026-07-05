package knowledgebase

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
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
