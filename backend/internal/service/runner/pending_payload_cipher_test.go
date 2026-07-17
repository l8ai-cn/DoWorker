package runner

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPendingPayloadCipherEncryptsAndAuthenticates(t *testing.T) {
	plaintext := []byte("private-key-secret")
	cipher := newPendingPayloadCipher(crypto.NewEncryptor("queue-key"))

	first, err := cipher.encrypt(plaintext)
	require.NoError(t, err)
	second, err := cipher.encrypt(plaintext)
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.NotContains(t, string(first), string(plaintext))
	assert.Contains(t, string(first), agentpod.PendingPayloadPrefix)
	opened, err := cipher.decrypt(first)
	require.NoError(t, err)
	assert.Equal(t, plaintext, opened)
}

func TestPendingPayloadCipherRejectsPlaintextTamperingAndWrongKey(t *testing.T) {
	cipher := newPendingPayloadCipher(crypto.NewEncryptor("queue-key"))
	encrypted, err := cipher.encrypt([]byte("secret"))
	require.NoError(t, err)

	_, err = cipher.decrypt([]byte("plaintext"))
	assert.ErrorIs(t, err, errPendingPayloadUnencrypted)

	encrypted[len(encrypted)-1] ^= 1
	_, err = cipher.decrypt(encrypted)
	require.Error(t, err)

	wrongCipher := newPendingPayloadCipher(crypto.NewEncryptor("wrong-key"))
	fresh, err := cipher.encrypt([]byte("secret"))
	require.NoError(t, err)
	_, err = wrongCipher.decrypt(fresh)
	require.Error(t, err)
}
