package acp

import (
	"encoding/json"

	"github.com/anthropics/agentsmesh/agentfile/capability"
)

func (c *ACPClient) CalibrateDeclaredCapabilities(podKey string, declared map[string]string) {
	if c == nil || c.transport == nil || len(declared) == 0 {
		return
	}
	type initResultProvider interface {
		InitializeResult() json.RawMessage
	}
	prov, ok := c.transport.(initResultProvider)
	if !ok {
		return
	}
	runtime := capability.ProbeRuntimeFromInitialize(prov.InitializeResult())
	capability.LogDeclaredRuntimeMismatches(c.logger, podKey, declared, runtime)
}
