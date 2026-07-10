package airesource

import "github.com/anthropics/agentsmesh/backend/pkg/slugkit"

type CredentialField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Secret   bool   `json:"secret"`
	Required bool   `json:"required"`
}

type ProviderDefinition struct {
	Key                    slugkit.Slug      `json:"key"`
	DisplayName            string            `json:"display_name"`
	Modalities             []Modality        `json:"modalities"`
	CredentialFields       []CredentialField `json:"credential_fields"`
	DefaultBaseURL         string            `json:"default_base_url"`
	ProtocolAdapter        string            `json:"protocol_adapter"`
	SupportsCustomEndpoint bool              `json:"supports_custom_endpoint"`
	SupportsModelDiscovery bool              `json:"supports_model_discovery"`
	ConnectionCheck        ConnectionCheck   `json:"connection_check"`
}

var providerRegistry = mustProviderRegistry([]ProviderDefinition{
	provider("openai", "OpenAI", "https://api.openai.com/v1", "openai-compatible", false, true,
		modalities(ModalityChat, ModalityImage, ModalityAudio, ModalityVideo, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("anthropic", "Anthropic", "https://api.anthropic.com", "anthropic", false, true,
		modalities(ModalityChat, ModalityMultimodal), apiKey()),
	provider("gemini", "Google Gemini", "https://generativelanguage.googleapis.com", "gemini", false, true,
		modalities(ModalityChat, ModalityImage, ModalityAudio, ModalityVideo, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("azure-openai", "Azure OpenAI", "", "azure-openai", true, false,
		modalities(ModalityChat, ModalityImage, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("openrouter", "OpenRouter", "https://openrouter.ai/api/v1", "openai-compatible", false, true,
		modalities(ModalityChat, ModalityMultimodal), apiKey()),
	provider("minimax", "MiniMax", "https://api.minimax.io/v1", "minimax", false, false,
		modalities(ModalityChat, ModalityAudio, ModalityVideo, ModalityMultimodal), apiKey()),
	provider("dashscope", "Alibaba DashScope", "https://dashscope.aliyuncs.com/compatible-mode/v1", "openai-compatible", false, false,
		modalities(ModalityChat, ModalityImage, ModalityAudio, ModalityVideo, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("doubao", "Volcengine Doubao", "https://ark.cn-beijing.volces.com/api/v3", "openai-compatible", false, false,
		modalities(ModalityChat, ModalityImage, ModalityAudio, ModalityVideo, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("deepseek", "DeepSeek", "https://api.deepseek.com", "openai-compatible", false, true,
		modalities(ModalityChat), apiKey()),
	provider("zhipu", "Zhipu AI", "https://open.bigmodel.cn/api/paas/v4", "openai-compatible", false, false,
		modalities(ModalityChat, ModalityMultimodal), apiKey()),
	provider("moonshot", "Moonshot AI", "https://api.moonshot.cn/v1", "openai-compatible", false, true,
		modalities(ModalityChat, ModalityMultimodal), apiKey()),
	provider("xai", "xAI", "https://api.x.ai/v1", "openai-compatible", false, true,
		modalities(ModalityChat, ModalityImage, ModalityMultimodal), apiKey()),
	provider("mistral", "Mistral AI", "https://api.mistral.ai/v1", "openai-compatible", false, true,
		modalities(ModalityChat, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("stability-ai", "Stability AI", "https://api.stability.ai", "stability-ai", false, false,
		modalities(ModalityImage), apiKey()),
	provider("black-forest-labs", "Black Forest Labs", "https://api.bfl.ai/v1", "black-forest-labs", false, false,
		modalities(ModalityImage), apiKey()),
	provider("elevenlabs", "ElevenLabs", "https://api.elevenlabs.io/v1", "elevenlabs", false, true,
		modalities(ModalityAudio), apiKey()),
	provider("azure-speech", "Azure Speech", "", "azure-speech", true, false,
		modalities(ModalityAudio), subscriptionKey(), region()),
	provider("runway", "Runway", "https://api.dev.runwayml.com/v1", "runway", false, false,
		modalities(ModalityVideo), apiKey()),
	provider("kling", "Kling", "https://api-singapore.klingai.com", "kling", false, false,
		modalities(ModalityImage, ModalityVideo), accessKey(), secretKey()),
	provider("hailuo", "Hailuo", "https://api.minimax.io", "hailuo", false, false,
		modalities(ModalityVideo), apiKey()),
	provider("luma", "Luma", "https://api.lumalabs.ai/dream-machine/v1", "luma", false, false,
		modalities(ModalityImage, ModalityVideo), apiKey()),
	provider("replicate", "Replicate", "https://api.replicate.com/v1", "replicate", false, true,
		modalities(ModalityImage, ModalityAudio, ModalityVideo, ModalityMultimodal), apiToken()),
	provider("fal", "fal.ai", "https://queue.fal.run", "fal", false, true,
		modalities(ModalityImage, ModalityAudio, ModalityVideo), apiKey()),
	provider("ideogram", "Ideogram", "https://api.ideogram.ai", "ideogram", false, false,
		modalities(ModalityImage), apiKey()),
	provider("custom-openai-compatible", "Custom OpenAI-compatible", "", "openai-compatible", true, true,
		modalities(ModalityChat, ModalityImage, ModalityAudio, ModalityVideo, ModalityEmbedding, ModalityMultimodal), apiKey()),
	provider("custom-anthropic-compatible", "Custom Anthropic-compatible", "", "anthropic", true, true,
		modalities(ModalityChat, ModalityMultimodal), apiKey()),
})

func Provider(key string) (ProviderDefinition, bool) {
	for _, definition := range providerRegistry {
		if definition.Key.String() == key {
			return cloneProvider(definition), true
		}
	}
	return ProviderDefinition{}, false
}

func Providers() []ProviderDefinition {
	definitions := make([]ProviderDefinition, len(providerRegistry))
	for i, definition := range providerRegistry {
		definitions[i] = cloneProvider(definition)
	}
	return definitions
}

func provider(
	key, name, baseURL, adapter string,
	customEndpoint, modelDiscovery bool,
	supported []Modality,
	credentials ...CredentialField,
) ProviderDefinition {
	definition := ProviderDefinition{
		Key: slugkit.Slug(key), DisplayName: name, Modalities: supported,
		CredentialFields: credentials, DefaultBaseURL: baseURL, ProtocolAdapter: adapter,
		SupportsCustomEndpoint: customEndpoint, SupportsModelDiscovery: modelDiscovery,
		ConnectionCheck: connectionCheck(key),
	}
	if err := ValidateProviderDefinition(definition); err != nil {
		panic(err)
	}
	return definition
}

func cloneProvider(definition ProviderDefinition) ProviderDefinition {
	definition.Modalities = append([]Modality(nil), definition.Modalities...)
	definition.CredentialFields = append([]CredentialField(nil), definition.CredentialFields...)
	definition.ConnectionCheck.StaticHeaders = append([]StaticHeader(nil), definition.ConnectionCheck.StaticHeaders...)
	return definition
}

func modalities(values ...Modality) []Modality { return values }

func apiKey() CredentialField    { return credential("api_key", "API key") }
func apiToken() CredentialField  { return credential("api_token", "API token") }
func accessKey() CredentialField { return credential("access_key", "Access key") }
func secretKey() CredentialField { return credential("secret_key", "Secret key") }
func subscriptionKey() CredentialField {
	return credential("subscription_key", "Subscription key")
}
func region() CredentialField {
	return CredentialField{Key: "region", Label: "Region", Required: true}
}

func credential(key, label string) CredentialField {
	return CredentialField{Key: key, Label: label, Secret: true, Required: true}
}
