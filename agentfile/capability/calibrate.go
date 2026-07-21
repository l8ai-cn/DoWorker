package capability

import (
	"encoding/json"
	"log/slog"
	"strings"
)

func ProbeRuntimeFromInitialize(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var body struct {
		Capabilities struct {
			Permissions bool `json:"permissions"`
			Streaming   bool `json:"streaming"`
		} `json:"capabilities"`
		AgentsmeshExtensions struct {
			ControlRequest bool `json:"controlRequest"`
		} `json:"agentcloudExtensions"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil
	}
	runtime := make(map[string]string)
	if body.Capabilities.Permissions {
		runtime["permission"] = "acp"
	}
	if body.Capabilities.Streaming {
		runtime["streaming"] = "true"
	} else {
		runtime["streaming"] = "false"
	}
	if body.AgentsmeshExtensions.ControlRequest {
		runtime["interrupt"] = "true"
	}
	if len(runtime) == 0 {
		return nil
	}
	return runtime
}

func LogDeclaredRuntimeMismatches(logger *slog.Logger, podKey string, declared, runtime map[string]string) {
	if logger == nil || len(declared) == 0 || len(runtime) == 0 {
		return
	}
	for axis, declaredVal := range declared {
		runtimeVal, ok := runtime[axis]
		if !ok {
			continue
		}
		if strings.EqualFold(declaredVal, runtimeVal) {
			continue
		}
		logger.Warn("capability mismatch: declared != runtime",
			"pod_key", podKey, "axis", axis,
			"declared", declaredVal, "runtime", runtimeVal)
	}
}
