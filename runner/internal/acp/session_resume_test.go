package acp

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResumer struct {
	cwd, resumeID string
	called        bool
}

func (s *stubResumer) Initialize(context.Context, io.Writer, io.Reader, io.Reader) error { return nil }
func (s *stubResumer) Handshake(context.Context) (string, error)                         { return "", nil }
func (s *stubResumer) NewSession(string, map[string]any) (string, error)                   { return "new-id", nil }
func (s *stubResumer) ResumeSession(cwd string, _ map[string]any, externalSessionID string) (string, error) {
	s.called = true
	s.cwd = cwd
	s.resumeID = externalSessionID
	return externalSessionID, nil
}
func (s *stubResumer) SendPrompt(string, string) error                        { return nil }
func (s *stubResumer) RespondToPermission(string, bool, map[string]any) error { return nil }
func (s *stubResumer) CancelSession(string) error                               { return nil }
func (s *stubResumer) SendControlRequest(string, string, map[string]any) (map[string]any, error) {
	return nil, ErrControlNotSupported
}
func (s *stubResumer) SupportedPermissionModes() []string { return nil }
func (s *stubResumer) ReadLoop(context.Context)           {}
func (s *stubResumer) Close()                             {}

func TestResumeOrNewSession_UsesResumeWhenSet(t *testing.T) {
	s := &stubResumer{}
	id, err := resumeOrNewSession(s, "/tmp/ws", nil, "vendor-sess-42")
	require.NoError(t, err)
	assert.True(t, s.called)
	assert.Equal(t, "vendor-sess-42", id)
	assert.Equal(t, "/tmp/ws", s.cwd)
}

func TestResumeOrNewSession_FallsBackToNew(t *testing.T) {
	s := &stubResumer{}
	id, err := resumeOrNewSession(s, "/tmp/ws", nil, "")
	require.NoError(t, err)
	assert.False(t, s.called)
	assert.Equal(t, "new-id", id)
}
