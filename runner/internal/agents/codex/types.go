package codex

import "encoding/json"

type initializeParams struct {
	ClientInfo   initializeClientInfo   `json:"clientInfo"`
	Capabilities initializeCapabilities `json:"capabilities"`
}

type initializeClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type initializeCapabilities struct {
	ExperimentalAPI    bool `json:"experimentalApi"`
	RequestAttestation bool `json:"requestAttestation"`
}

type threadStartResult struct {
	Model  string `json:"model"`
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

type modelListParams struct {
	Cursor *string `json:"cursor,omitempty"`
	Limit  uint32  `json:"limit"`
}

type modelListResponse struct {
	Data       []codexModel `json:"data"`
	NextCursor *string      `json:"nextCursor"`
}

type codexModel struct {
	Model  string `json:"model"`
	Hidden bool   `json:"hidden"`
}

type threadSettingsUpdateParams struct {
	ThreadID string `json:"threadId"`
	Model    string `json:"model"`
}

type turnInterruptParams struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

type backgroundTerminalListParams struct {
	ThreadID string `json:"threadId"`
}

type backgroundTerminalListResponse struct {
	Data []struct {
		ProcessID string `json:"processId"`
	} `json:"data"`
}

type backgroundTerminalTerminateParams struct {
	ThreadID  string `json:"threadId"`
	ProcessID string `json:"processId"`
}

type backgroundTerminalTerminateResponse struct {
	Terminated bool `json:"terminated"`
}

type turnStartedParams struct {
	Turn struct {
		ID string `json:"id"`
	} `json:"turn"`
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

type itemOutputDelta struct {
	ItemID string `json:"itemId"`
	Delta  string `json:"delta"`
}

type fileUpdateChange struct {
	Path string          `json:"path"`
	Kind json.RawMessage `json:"kind"`
	Diff string          `json:"diff"`
}

type codexItem struct {
	ID               string                    `json:"id"`
	Type             string                    `json:"type"`
	Status           string                    `json:"status,omitempty"`
	Text             string                    `json:"text,omitempty"`
	Content          []agentMessageContentPart `json:"content,omitempty"`
	Message          string                    `json:"message,omitempty"`
	Command          json.RawMessage           `json:"command,omitempty"`
	CWD              string                    `json:"cwd,omitempty"`
	ExitCode         *int                      `json:"exitCode,omitempty"`
	AggregatedOutput string                    `json:"aggregatedOutput,omitempty"`
	ToolName         string                    `json:"toolName,omitempty"`
	Arguments        json.RawMessage           `json:"arguments,omitempty"`
	FilePath         string                    `json:"filePath,omitempty"`
	Changes          []fileUpdateChange        `json:"changes,omitempty"`
}

type itemStartedParams struct {
	Item codexItem `json:"item"`
}

type itemCompletedParams struct {
	Item codexItem `json:"item"`
}

type fileChangePatchUpdatedParams struct {
	ItemID  string             `json:"itemId"`
	Changes []fileUpdateChange `json:"changes"`
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
		ID     string `json:"id"`
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
