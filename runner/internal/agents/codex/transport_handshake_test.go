package codex

import (
	"bufio"
	"encoding/json"
	"testing"
)

func TestTransportHandshakeEnablesExperimentalAPI(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()

	go func() {
		scanner := bufio.NewScanner(fixture.StdinPR)
		if !scanner.Scan() {
			t.Error("initialize request was not written")
			return
		}
		var request struct {
			ID     int64            `json:"id"`
			Method string           `json:"method"`
			Params initializeParams `json:"params"`
		}
		requireNoError(t, json.Unmarshal(scanner.Bytes(), &request))
		if request.Method != "initialize" {
			t.Errorf("method = %q", request.Method)
			return
		}
		if !request.Params.Capabilities.ExperimentalAPI {
			t.Error("experimentalApi capability is disabled")
			return
		}
		if request.Params.Capabilities.RequestAttestation {
			t.Error("requestAttestation capability is enabled")
			return
		}
		writeResponse(fixture.PW, request.ID, map[string]any{}, nil)

		if !scanner.Scan() {
			t.Error("initialized notification was not written")
			return
		}
		if !scanner.Scan() {
			t.Error("model/list request was not written")
			return
		}
		var modelRequest struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
		}
		requireNoError(t, json.Unmarshal(scanner.Bytes(), &modelRequest))
		if modelRequest.Method != "model/list" {
			t.Errorf("method = %q", modelRequest.Method)
			return
		}
		writeModelListResponse(fixture.PW, modelRequest.ID)
	}()

	_, err := fixture.transport.Handshake(fixture.transport.ctx)
	requireNoError(t, err)
}
