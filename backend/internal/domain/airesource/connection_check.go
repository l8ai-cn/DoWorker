package airesource

type ConnectionAuthStrategy string

const (
	ConnectionAuthBearer      ConnectionAuthStrategy = "bearer"
	ConnectionAuthHeader      ConnectionAuthStrategy = "header"
	ConnectionAuthQuery       ConnectionAuthStrategy = "query"
	ConnectionAuthUnsupported ConnectionAuthStrategy = "unsupported"
)

type StaticHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ConnectionCheck struct {
	Method        string                 `json:"method,omitempty"`
	Path          string                 `json:"path,omitempty"`
	AuthStrategy  ConnectionAuthStrategy `json:"auth_strategy"`
	CredentialKey string                 `json:"credential_key,omitempty"`
	AuthName      string                 `json:"auth_name,omitempty"`
	StaticHeaders []StaticHeader         `json:"static_headers,omitempty"`
}

func connectionCheck(providerKey string) ConnectionCheck {
	switch providerKey {
	case "openai", "minimax", "custom-openai-compatible":
		return bearerCheck("/models", "api_key")
	case "openrouter":
		return bearerCheck("/key", "api_key")
	case "deepseek", "xai", "mistral":
		return bearerCheck("/models", "api_key")
	case "doubao":
		return bearerCheck("/contents/generations/tasks", "api_key")
	case "sub2api-seedance":
		return bearerCheck("/contents/generations/tasks", "api_key")
	case "anthropic":
		return headerCheck("/v1/models", "api_key", "x-api-key", StaticHeader{Name: "anthropic-version", Value: "2023-06-01"})
	case "gemini":
		return ConnectionCheck{Method: "GET", Path: "/v1beta/models", AuthStrategy: ConnectionAuthQuery, CredentialKey: "api_key", AuthName: "key"}
	case "azure-openai":
		return headerCheck("/openai/v1/models", "api_key", "api-key")
	case "stability-ai":
		return bearerCheck("/v1/user/account", "api_key")
	case "black-forest-labs":
		return headerCheck("/credits", "api_key", "x-key")
	case "elevenlabs":
		return headerCheck("/models", "api_key", "xi-api-key")
	case "runway":
		return bearerCheck("/organization", "api_key", StaticHeader{Name: "X-Runway-Version", Value: "2024-11-06"})
	case "luma":
		return bearerCheck("/generations", "api_key")
	case "replicate":
		return bearerCheck("/models", "api_token")
	case "ideogram":
		return headerCheck("/models", "api_key", "Api-Key")
	case "dashscope", "zhipu", "moonshot",
		"azure-speech", "kling", "hailuo", "fal", "custom-anthropic-compatible":
		return unsupportedCheck()
	default:
		panic("provider connection check policy is not explicit: " + providerKey)
	}
}

func bearerCheck(path, credential string, headers ...StaticHeader) ConnectionCheck {
	return ConnectionCheck{Method: "GET", Path: path, AuthStrategy: ConnectionAuthBearer, CredentialKey: credential, StaticHeaders: headers}
}

func headerCheck(path, credential, name string, headers ...StaticHeader) ConnectionCheck {
	return ConnectionCheck{Method: "GET", Path: path, AuthStrategy: ConnectionAuthHeader, CredentialKey: credential, AuthName: name, StaticHeaders: headers}
}

func unsupportedCheck() ConnectionCheck {
	return ConnectionCheck{AuthStrategy: ConnectionAuthUnsupported}
}
