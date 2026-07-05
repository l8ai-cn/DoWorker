package acp

// externalSessionResumer is implemented by transports that support vendor
// session/resume (DoAgent, standard ACP). Optional — absent transports fall
// back to session/new when a resume id is configured.
type externalSessionResumer interface {
	ResumeSession(cwd string, mcpServers map[string]any, externalSessionID string) (string, error)
}

func resumeOrNewSession(t Transport, cwd string, mcpServers map[string]any, resumeID string) (string, error) {
	if resumeID != "" {
		if r, ok := t.(externalSessionResumer); ok {
			return r.ResumeSession(cwd, mcpServers, resumeID)
		}
	}
	return t.NewSession(cwd, mcpServers)
}
