package codex

import (
	"encoding/json"
	"fmt"
	"time"
)

const interruptRequestTimeout = 10 * time.Second

func (t *transport) CancelSession(sessionID string) error {
	turnID := t.getActiveTurnID()
	if turnID != "" {
		if _, err := t.request(
			"turn/interrupt",
			turnInterruptParams{ThreadID: sessionID, TurnID: turnID},
		); err != nil {
			return err
		}
	}

	result, err := t.request(
		"thread/backgroundTerminals/list",
		backgroundTerminalListParams{ThreadID: sessionID},
	)
	if err != nil {
		return err
	}
	var terminals backgroundTerminalListResponse
	if err := json.Unmarshal(result, &terminals); err != nil {
		return fmt.Errorf("decode background terminal list: %w", err)
	}
	if turnID == "" && len(terminals.Data) == 0 {
		return fmt.Errorf("no active turn or background terminals to interrupt")
	}
	for _, terminal := range terminals.Data {
		if terminal.ProcessID == "" {
			return fmt.Errorf("background terminal missing process id")
		}
		if err := t.terminateBackgroundTerminal(sessionID, terminal.ProcessID); err != nil {
			return err
		}
	}
	return nil
}

func (t *transport) terminateBackgroundTerminal(sessionID, processID string) error {
	result, err := t.request(
		"thread/backgroundTerminals/terminate",
		backgroundTerminalTerminateParams{ThreadID: sessionID, ProcessID: processID},
	)
	if err != nil {
		return err
	}
	var response backgroundTerminalTerminateResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return fmt.Errorf("decode background terminal termination: %w", err)
	}
	if !response.Terminated {
		return fmt.Errorf("background terminal %s was not terminated", processID)
	}
	return nil
}

func (t *transport) request(method string, params any) (json.RawMessage, error) {
	pending, err := t.tracker.SendRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", method, err)
	}
	response, err := t.tracker.WaitResponse(pending, interruptRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("wait %s: %w", method, err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf(
			"%s error: code=%d message=%s",
			method,
			response.Error.Code,
			response.Error.Message,
		)
	}
	return response.Result, nil
}
