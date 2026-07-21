package acp

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestParseAgentsmeshExtensions(t *testing.T) {
	t.Run("agent advertises exact control capabilities", func(t *testing.T) {
		raw := json.RawMessage(`{"agentcloudExtensions":{"controlRequest":true,"permissionModes":["bypass","ask_dangerous"],"artifactActions":["image.edit","presentation.export"]}}`)
		extensions, err := parseAgentsmeshExtensions(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !extensions.ControlRequest {
			t.Error("controlRequest = false, want true")
		}
		if !slices.Equal(extensions.PermissionModes, []string{"bypass", "ask_dangerous"}) {
			t.Errorf("permissionModes = %v", extensions.PermissionModes)
		}
		if !slices.Equal(extensions.ArtifactActions, []string{"image.edit", "presentation.export"}) {
			t.Errorf("artifactActions = %v", extensions.ArtifactActions)
		}
	})

	t.Run("no extension block leaves both zero", func(t *testing.T) {
		extensions, err := parseAgentsmeshExtensions(json.RawMessage(`{"protocolVersion":1}`))
		if err != nil {
			t.Fatal(err)
		}
		if extensions.ControlRequest ||
			extensions.PermissionModes != nil ||
			extensions.ArtifactActions != nil {
			t.Errorf("got %+v, want zero extensions", extensions)
		}
	})

	t.Run("artifact actions require control request support", func(t *testing.T) {
		_, err := parseAgentsmeshExtensions(json.RawMessage(
			`{"agentcloudExtensions":{"artifactActions":["image.edit"]}}`,
		))
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("invalid actions fail the handshake", func(t *testing.T) {
		for _, raw := range []json.RawMessage{
			json.RawMessage(`{"agentcloudExtensions":{"controlRequest":true,"artifactActions":["image.edit","image.edit"]}}`),
			json.RawMessage(`{"agentcloudExtensions":{"controlRequest":true,"artifactActions":["Image Edit"]}}`),
			json.RawMessage(`not json`),
		} {
			if _, err := parseAgentsmeshExtensions(raw); err == nil {
				t.Fatalf("expected validation error for %s", raw)
			}
		}
	})
}
