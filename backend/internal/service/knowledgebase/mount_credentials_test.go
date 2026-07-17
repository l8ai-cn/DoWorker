package knowledgebase

import (
	"encoding/json"
	"strings"
	"testing"

	kbdomain "github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestMountDeployKeysAreDistinctValidAndEncrypted(t *testing.T) {
	keys, err := newMountDeployKeys()
	require.NoError(t, err)
	assert.NotEqual(t, keys.readOnlyPrivate, keys.readWritePrivate)
	assert.True(t, strings.HasPrefix(keys.readOnlyPublic, "ssh-ed25519 "))
	assert.True(t, strings.HasPrefix(keys.readWritePublic, "ssh-ed25519 "))
	_, err = ssh.ParseRawPrivateKey([]byte(keys.readOnlyPrivate))
	require.NoError(t, err)
	_, err = ssh.ParseRawPrivateKey([]byte(keys.readWritePrivate))
	require.NoError(t, err)

	raw, err := addMountDeployKeys(json.RawMessage(`{"folder_id":"docs"}`), keys)
	require.NoError(t, err)
	service := &Service{secrets: crypto.NewEncryptor("mount-key-test-secret")}
	encrypted, err := service.encryptSourceSecrets(raw)
	require.NoError(t, err)
	assert.NotContains(t, string(encrypted), keys.readOnlyPrivate)
	assert.NotContains(t, string(encrypted), keys.readWritePrivate)

	ro, err := service.mountPrivateKey(encrypted, kbdomain.MountModeReadOnly)
	require.NoError(t, err)
	rw, err := service.mountPrivateKey(encrypted, kbdomain.MountModeReadWrite)
	require.NoError(t, err)
	assert.Equal(t, keys.readOnlyPrivate, ro)
	assert.Equal(t, keys.readWritePrivate, rw)

	redacted := RedactedSourceConfigJSON(encrypted)
	assert.NotContains(t, redacted, readOnlyDeployKeyConfig)
	assert.NotContains(t, redacted, readWriteDeployKeyConfig)
	assert.Contains(t, redacted, `"folder_id":"docs"`)
}

func TestMountPrivateKeyRejectsLegacyKnowledgeBaseWithoutDeployKey(t *testing.T) {
	service := &Service{secrets: crypto.NewEncryptor("mount-key-test-secret")}
	_, err := service.mountPrivateKey(json.RawMessage(`{}`), kbdomain.MountModeReadOnly)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotConfigured)
}
