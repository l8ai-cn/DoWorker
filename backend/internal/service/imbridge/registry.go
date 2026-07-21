package imbridge

import (
	"fmt"
	"net/http"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/imbridge"
)

func NewRegistry(httpClient *http.Client) map[string]Provider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return map[string]Provider{
		domain.ProviderFeishu:   &FeishuProvider{HTTP: httpClient},
		domain.ProviderDingTalk: &DingTalkProvider{HTTP: httpClient},
		domain.ProviderWeCom:    &WeComProvider{HTTP: httpClient},
		domain.ProviderSlack:    &SlackProvider{HTTP: httpClient},
		domain.ProviderWeixin:   NewWeixinProvider(httpClient),
	}
}

func GetProvider(registry map[string]Provider, providerType string) (Provider, error) {
	if providerType == domain.ProviderWeChat {
		providerType = domain.ProviderWeixin
	}
	p, ok := registry[providerType]
	if !ok {
		return nil, fmt.Errorf("unsupported im provider: %s", providerType)
	}
	return p, nil
}

func ListProviderMeta(registry map[string]Provider) []map[string]string {
	out := make([]map[string]string, 0, len(registry))
	for _, t := range domain.SupportedProviders {
		if p, ok := registry[t]; ok {
			out = append(out, map[string]string{
				"type":         p.Type(),
				"display_name": p.DisplayName(),
			})
		}
	}
	return out
}
