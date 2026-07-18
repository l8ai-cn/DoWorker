package secretguard

import "testing"

func TestContainsCredentialLiteral(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"OpenAI", "use sk-proj-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"Anthropic", "sk-ant-api03-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"GitHub", "ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"GitHub fine grained", "github_pat_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"GitLab", "glpat-aaaaaaaaaaaaaaaaaaaa"},
		{"Slack", "xoxb-" + "123456789012-123456789012-abcdefghijklmnop"},
		{"AWS", "AKIAIOSFODNN7EXAMPLE"},
		{"Google", "AIzaSyAaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"JWT", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature"},
		{"Bearer", "Authorization: Bearer abcdefghijklmnopqrstuvwxyz012345"},
		{"PEM", "-----BEGIN PRIVATE KEY-----\nZmFrZQ==\n-----END PRIVATE KEY-----"},
		{"short password assignment", "password=hunter2"},
		{"short pwd assignment", "pwd=hunter2"},
		{"short API key assignment", "use api_key: plaintext-value"},
		{"PostgreSQL URL", "postgres://loop_user:hunter2@db.internal/loops"},
		{"Redis URL", "redis://cache-user:hunter2@redis.internal/0"},
		{"whole base64", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef"},
		{"embedded base64", "prefix ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef suffix"},
		{"whole URL base64", "ABCDEFGHIJKLMNOPQRSTUVWXYZab-_12"},
		{"whole hex", "0123456789abcdef0123456789abcdef"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !ContainsCredentialLiteral(test.value) {
				t.Fatalf("ContainsCredentialLiteral(%q) = false, want true", test.value)
			}
		})
	}
}

func TestContainsCredentialLiteralAllowsOrdinaryText(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"model", "claude-3-7-sonnet-20250219"},
		{"URL", "https://api.anthropic.com/v1"},
		{"short bearer example", "send a Bearer token header"},
		{"token budget", "token budget=80000"},
		{"credential prefix inside word", "track task-ant-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa in docs"},
		{"short hex", "0123456789abcdef"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if ContainsCredentialLiteral(test.value) {
				t.Fatalf("ContainsCredentialLiteral(%q) = true, want false", test.value)
			}
		})
	}
}

func TestIsSensitiveName(t *testing.T) {
	for _, name := range []string{
		"api_key", "ACCESS_TOKEN", "auth-token", "refresh_token",
		"secret", "client_secret", "password", "credential", "private_key",
	} {
		t.Run(name, func(t *testing.T) {
			if !IsSensitiveName(name) {
				t.Fatalf("IsSensitiveName(%q) = false, want true", name)
			}
		})
	}

	for _, name := range []string{"model", "base_url", "token_budget"} {
		t.Run(name, func(t *testing.T) {
			if IsSensitiveName(name) {
				t.Fatalf("IsSensitiveName(%q) = true, want false", name)
			}
		})
	}
}
