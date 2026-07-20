package config

import "time"

func loadMarketplaceConfig() MarketplaceConfig {
	return MarketplaceConfig{
		SyncInterval:    getEnvDuration("MARKETPLACE_SYNC_INTERVAL", time.Hour),
		RegistryEnabled: getEnvBool("MCP_REGISTRY_ENABLED", true),
		RegistryURL:     getEnv("MCP_REGISTRY_URL", "https://registry.modelcontextprotocol.io"),
	}
}

func loadKnowledgeBaseConfig() KnowledgeBaseConfig {
	return KnowledgeBaseConfig{
		GiteaURL:           getEnv("KB_GITEA_URL", ""),
		GiteaToken:         getEnv("KB_GITEA_TOKEN", ""),
		GiteaOrg:           getEnv("KB_GITEA_ORG", "am-kb"),
		CloneBaseURL:       getEnv("KB_GITEA_CLONE_URL", ""),
		RepositoryBaseURLs: getEnvList("KB_GITEA_REPOSITORY_BASE_URLS", nil),
		SSHCloneBaseURL:    getEnv("KB_GITEA_SSH_URL", ""),
		SSHKnownHosts:      getEnv("KB_GITEA_KNOWN_HOSTS", ""),
		SyncInterval:       getEnvDuration("KB_SYNC_INTERVAL", time.Hour),
	}
}

func loadRelayConfig() RelayConfig {
	return RelayConfig{
		BaseDomain: getEnv("RELAY_BASE_DOMAIN", ""),
		DNS: DNSConfig{
			Provider:              getEnv("DNS_PROVIDER", ""),
			CloudflareAPIToken:    getEnv("CLOUDFLARE_API_TOKEN", ""),
			CloudflareZoneID:      getEnv("CLOUDFLARE_ZONE_ID", ""),
			AliyunAccessKeyID:     getEnv("ALIYUN_ACCESS_KEY_ID", ""),
			AliyunAccessKeySecret: getEnv("ALIYUN_ACCESS_KEY_SECRET", ""),
		},
		ACME: ACMEConfig{
			Enabled:      getEnvBool("ACME_ENABLED", false),
			Email:        getEnv("ACME_EMAIL", ""),
			DirectoryURL: getEnv("ACME_DIRECTORY_URL", ""),
			StorageDir:   getEnv("ACME_STORAGE_DIR", "/var/lib/agentsmesh/acme"),
			Staging:      getEnvBool("ACME_STAGING", false),
		},
	}
}
