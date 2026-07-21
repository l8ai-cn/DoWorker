package airesource

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingCipher struct{ encryptErr, decryptErr error }

func (cipher failingCipher) Encrypt(string) (string, error) { return "", cipher.encryptErr }
func (cipher failingCipher) Decrypt(string) (string, error) { return "", cipher.decryptErr }

func TestInternalCredentialMapsCannotBeSerialized(t *testing.T) {
	secret := "serialization-secret"
	values := []any{
		CreateConnectionInput{Credentials: map[string]string{"api_key": secret}},
		UpdateConnectionInput{Credentials: map[string]string{"api_key": secret}},
		ResolvedResource{Credentials: map[string]string{"api_key": secret}},
		ProbeInput{Credentials: map[string]string{"api_key": secret}},
	}
	for _, value := range values {
		encoded, err := json.Marshal(value)
		require.NoError(t, err)
		assert.NotContains(t, string(encoded), secret)
		assert.NotContains(t, string(encoded), "api_key")
	}
}

func TestCipherErrorsNeverExposeSensitiveMaterial(t *testing.T) {
	f := newFixture()
	leakedPlaintext := "plaintext-encryption-leak"
	service, err := NewService(Dependencies{Repository: f.repo, Cipher: failingCipher{encryptErr: errors.New(leakedPlaintext)}, Members: f.members, Prober: f.prober, Mutations: f.mutations, Endpoints: allowingEndpoints{}})
	require.NoError(t, err)
	_, err = service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", Credentials: map[string]string{"api_key": "input-secret"}})
	assert.ErrorIs(t, err, ErrEncrypt)
	assert.False(t, strings.Contains(err.Error(), leakedPlaintext) || strings.Contains(err.Error(), "input-secret"))

	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "decrypt-main", "stored-secret")
	resource := createResource(t, f, connection.ID, "model-b")
	leakedCiphertext := "ciphertext-decryption-leak"
	service, err = NewService(Dependencies{Repository: f.repo, Cipher: failingCipher{decryptErr: errors.New(leakedCiphertext)}, Members: f.members, Prober: f.prober, Mutations: f.mutations, Endpoints: allowingEndpoints{}})
	require.NoError(t, err)
	_, err = service.ResolveExact(context.Background(), actor(1), 0, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrDecrypt)
	assert.NotContains(t, err.Error(), leakedCiphertext)
}
