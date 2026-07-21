package grok

import (
	"encoding/json"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func TestAuthenticateUsesXAIAPIKeyMethod(t *testing.T) {
	t.Setenv("XAI_API_KEY", "xai-test")
	requester := &recordingRequester{}

	err := authenticate(requester, json.RawMessage(`{
		"authMethods": [{"id": "xai.api_key"}]
	}`))
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if requester.method != "authenticate" {
		t.Fatalf("method = %q, want authenticate", requester.method)
	}
	if requester.params["methodId"] != "xai.api_key" {
		t.Fatalf("methodId = %v, want xai.api_key", requester.params["methodId"])
	}
}

func TestAuthenticateRejectsMissingAPIKey(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")
	err := authenticate(&recordingRequester{}, json.RawMessage(`{
		"authMethods": [{"id": "xai.api_key"}]
	}`))
	if err == nil {
		t.Fatal("authenticate succeeded without XAI_API_KEY")
	}
}

func TestAuthenticateRejectsUnavailableAPIKeyMethod(t *testing.T) {
	t.Setenv("XAI_API_KEY", "xai-test")
	err := authenticate(&recordingRequester{}, json.RawMessage(`{
		"authMethods": [{"id": "xai.oauth"}]
	}`))
	if err == nil {
		t.Fatal("authenticate succeeded without xai.api_key method")
	}
}

type recordingRequester struct {
	method string
	params map[string]any
}

func (r *recordingRequester) Request(method string, params any) (json.RawMessage, error) {
	r.method = method
	r.params = params.(map[string]any)
	return json.RawMessage(`{}`), nil
}

var _ acp.HandshakeRequester = (*recordingRequester)(nil)
