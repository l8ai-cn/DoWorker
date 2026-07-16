package config

import "time"

// PKIConfig holds PKI (certificate) configuration for Runner mTLS authentication
// Required for Runner communication via gRPC + mTLS
type PKIConfig struct {
	CACertFile     string // Path to CA certificate file (required)
	CAKeyFile      string // Path to CA private key file (required)
	ServerCertFile string // Path to server certificate file (optional, generated if not set)
	ServerKeyFile  string // Path to server private key file (optional)
	ValidityDays   int    // Certificate validity period in days (default: 365)
}

type GRPCConfig struct {
	Address  string // gRPC server listen address (default: :9090)
	Endpoint string // Public gRPC endpoint URL for Runners (e.g., grpcs://api.agentsmesh.cn:9443)
}

type AdminConfig struct {
	Enabled bool // Enable admin console
}

func (c AdminConfig) IsEnabled() bool {
	return c.Enabled
}

type StorageConfig struct {
	Endpoint       string   // S3 endpoint (empty for AWS, set for MinIO/OSS)
	PublicEndpoint string   // Public endpoint for browser access (if different from Endpoint)
	RunnerEndpoint string   // Endpoint reachable by runner pods for presigned resource downloads (empty = Endpoint)
	Region         string   // AWS region or equivalent
	Bucket         string   // Bucket name
	AccessKey      string   // Access key ID
	SecretKey      string   // Secret access key
	UseSSL         bool     // Use HTTPS
	PublicUseSSL   bool     // Use HTTPS for browser-facing URLs
	UsePathStyle   bool     // Use path-style URLs (required for MinIO)
	MaxFileSize    int64    // Max file size in MB
	AllowedTypes   []string // Allowed MIME types
}

type EmailConfig struct {
	Provider    string // "resend" or "console"
	ResendKey   string
	FromAddress string
}

// KnowledgeBaseConfig points at the internal Gitea instance hosting KB
// repositories. When URL or token is empty the KB feature is disabled
// (same nil-service pattern as StorageConfig / file service).
type KnowledgeBaseConfig struct {
	GiteaURL     string // Gitea base URL as reachable from the backend
	GiteaToken   string // admin-scoped service token for repo provisioning
	GiteaOrg     string // Gitea org namespace owning all KB repos
	CloneBaseURL string // clone base URL as reachable from runners

	SyncInterval time.Duration // external-source (feishu/dingtalk/google) sync cadence
}

func (c KnowledgeBaseConfig) Enabled() bool {
	return c.GiteaURL != "" && c.GiteaToken != ""
}
