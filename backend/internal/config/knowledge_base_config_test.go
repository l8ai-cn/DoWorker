package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnowledgeBaseConfigEnabledRequiresPinnedSSHConfiguration(t *testing.T) {
	config := KnowledgeBaseConfig{
		GiteaURL:        "http://gitea:3000",
		GiteaToken:      "service-token",
		SSHCloneBaseURL: "ssh://git@gitea:22",
		SSHKnownHosts:   "gitea ssh-ed25519 host-key",
	}
	assert.True(t, config.Enabled())

	config.SSHKnownHosts = ""
	assert.False(t, config.Enabled())

	config.SSHKnownHosts = "gitea ssh-ed25519 host-key"
	config.SSHCloneBaseURL = ""
	assert.False(t, config.Enabled())
}

func TestLoadKnowledgeBaseConfigReadsTrustedRepositoryOrigins(t *testing.T) {
	t.Setenv("KB_GITEA_REPOSITORY_BASE_URLS", "http://gitea:3000, http://gitea.internal")

	config := loadKnowledgeBaseConfig()

	assert.Equal(t, []string{"http://gitea:3000", "http://gitea.internal"}, config.RepositoryBaseURLs)
}
