package codex

import "encoding/json"

type threadStartResult struct {
	Thread struct {
		ID string `json:"id"`
	} `json:"thread"`
}

type turnStartParams struct {
	ThreadID string      `json:"threadId"`
	Input    []turnInput `json:"input"`
}

type turnInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type turnInterruptParams struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId,omitempty"`
}

type approvalRequestParams struct {
	Command     string `json:"command,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Message     string `json:"message,omitempty"`
}

type agentMessageDelta struct {
	ItemID string `json:"itemId"`
	Delta  string `json:"delta"`
}

type reasoningDelta struct {
	ItemID string `json:"itemId"`
	Delta  string `json:"delta"`
}

type planDelta struct {
	ItemID string `json:"itemId"`
	Delta  string `json:"delta"`
}

type itemStartedParams struct {
	Item struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		// Command shape varies across Codex builds (string vs array); it is not
		// consumed here, so keep it raw to avoid dropping the whole item/started
		// on a type mismatch.
		Command  json.RawMessage `json:"command,omitempty"`
		ToolName string          `json:"toolName,omitempty"`
		FilePath string          `json:"filePath,omitempty"`
	} `json:"item"`
}

type itemCompletedParams struct {
	Item struct {
		ID               string                    `json:"id"`
		Type             string                    `json:"type"`
		Status           string                    `json:"status,omitempty"`
		Text             string                    `json:"text,omitempty"`
		Content          []agentMessageContentPart `json:"content,omitempty"`
		Message          string                    `json:"message,omitempty"`
		ExitCode         *int                      `json:"exitCode,omitempty"`
		AggregatedOutput string                    `json:"aggregatedOutput,omitempty"`
		ToolName         string                    `json:"toolName,omitempty"`
		FilePath         string                    `json:"filePath,omitempty"`
	} `json:"item"`
}

type errorNotificationParams struct {
	Message string `json:"message"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type threadStatusChangedParams struct {
	Status struct {
		Type string `json:"type"`
	} `json:"status"`
}

type turnCompletedParams struct {
	Turn struct {
		Status string `json:"status"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
		Usage *struct {
			InputTokens       int64 `json:"input_tokens"`
			CachedInputTokens int64 `json:"cached_input_tokens"`
			OutputTokens      int64 `json:"output_tokens"`
		} `json:"usage,omitempty"`
	} `json:"turn"`
}
