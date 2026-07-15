package workerruntime

func defaultRuntimeImages() []CatalogRuntimeImage {
	return []CatalogRuntimeImage{
		{
			ID:              1,
			Slug:            "codex-cli-stable",
			Name:            "Codex CLI",
			Reference:       "repo.aiedulab.cn:8443/agentsmesh/runner-codex-cli@sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1",
			Digest:          "sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1",
			WorkerTypeSlugs: []string{"codex-cli"},
			Enabled:         true,
		},
		{
			ID:              2,
			Slug:            "claude-code-stable",
			Name:            "Claude Code",
			Reference:       "repo.aiedulab.cn:8443/agentsmesh/runner-claude-code@sha256:a9a02976dec14907be8eb6a7f68cd1adc5158099645244be733546b0f3e7041f",
			Digest:          "sha256:a9a02976dec14907be8eb6a7f68cd1adc5158099645244be733546b0f3e7041f",
			WorkerTypeSlugs: []string{"claude-code"},
			Enabled:         true,
		},
		{
			ID:              3,
			Slug:            "gemini-cli-stable",
			Name:            "Gemini CLI",
			Reference:       "repo.aiedulab.cn:8443/agentsmesh/runner-gemini-cli@sha256:852dba55bcc3213c72a7ee94e9c2da29a44e2ba0d5a9c0a8c15fea5adb8c6cd4",
			Digest:          "sha256:852dba55bcc3213c72a7ee94e9c2da29a44e2ba0d5a9c0a8c15fea5adb8c6cd4",
			WorkerTypeSlugs: []string{"gemini-cli"},
			Enabled:         true,
		},
		{
			ID:              4,
			Slug:            "do-agent-stable",
			Name:            "DoAgent",
			Reference:       "agentsmesh/runner-do-agent@sha256:104e8c621400d35bc5fed37eaf2d059cca7fbe7c47f9c4fe3b59093a1c2b7cc5",
			Digest:          "sha256:104e8c621400d35bc5fed37eaf2d059cca7fbe7c47f9c4fe3b59093a1c2b7cc5",
			WorkerTypeSlugs: []string{"do-agent", "seedance-expert"},
			Enabled:         true,
		},
	}
}
