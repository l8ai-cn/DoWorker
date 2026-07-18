package runner

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
)

var (
	errPendingPayloadCipherMissing = errors.New("pending command payload cipher is not configured")
	errPendingPayloadUnencrypted   = errors.New("pending command payload is not encrypted")
)

type pendingPayloadCipher struct {
	encryptor *crypto.Encryptor
}

func newPendingPayloadCipher(encryptor *crypto.Encryptor) *pendingPayloadCipher {
	return &pendingPayloadCipher{encryptor: encryptor}
}

func (c *pendingPayloadCipher) encrypt(payload []byte) ([]byte, error) {
	if c == nil || c.encryptor == nil {
		return nil, errPendingPayloadCipherMissing
	}
	encrypted, err := c.encryptor.Encrypt(string(payload))
	if err != nil {
		return nil, fmt.Errorf("encrypt pending command payload: %w", err)
	}
	return append([]byte(agentpod.PendingPayloadPrefix), encrypted...), nil
}

func (c *pendingPayloadCipher) decrypt(payload []byte) ([]byte, error) {
	if c == nil || c.encryptor == nil {
		return nil, errPendingPayloadCipherMissing
	}
	if !bytes.HasPrefix(payload, []byte(agentpod.PendingPayloadPrefix)) {
		return nil, errPendingPayloadUnencrypted
	}
	plaintext, err := c.encryptor.Decrypt(string(payload[len(agentpod.PendingPayloadPrefix):]))
	if err != nil {
		return nil, fmt.Errorf("decrypt pending command payload: %w", err)
	}
	return []byte(plaintext), nil
}

func (q *PendingCommandQueue) SealPayload(payload []byte) ([]byte, error) {
	if q == nil {
		return nil, errPendingPayloadCipherMissing
	}
	return q.payloadCipher.encrypt(payload)
}

func (q *PendingCommandQueue) PayloadMatches(envelope, plaintext []byte) (bool, error) {
	if q == nil {
		return false, errPendingPayloadCipherMissing
	}
	decrypted, err := q.payloadCipher.decrypt(envelope)
	if err != nil {
		return false, err
	}
	return bytes.Equal(decrypted, plaintext), nil
}
