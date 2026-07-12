package config

import (
	"fmt"
	"strings"
)

// BaseURL returns the base URL with protocol (http:// or https://)
func (c *Config) BaseURL() string {
	protocol := "http"
	if c.UseHTTPS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.PrimaryDomain)
}

func (c *Config) WebSocketBaseURL() string {
	protocol := "ws"
	if c.UseHTTPS {
		protocol = "wss"
	}
	return fmt.Sprintf("%s://%s", protocol, c.PrimaryDomain)
}

func (c *Config) FrontendURL() string {
	return c.BaseURL()
}

func (c *Config) PublicWebBaseURL() string {
	return strings.TrimRight(c.PublicWebURL, "/")
}

func (c *Config) MobilePublicBaseURL() string {
	return strings.TrimRight(c.MobilePublicURL, "/")
}

func (c *Config) APIBaseURL() string {
	return c.BaseURL() + "/api"
}

func (c *Config) RelayURL() string {
	return c.WebSocketBaseURL() + "/relay"
}

// TunnelURL returns the WebSocket endpoint runners dial for the outbound HTTP
// data-plane tunnel.
func (c *Config) TunnelURL() string {
	return c.WebSocketBaseURL() + "/runner/tunnel"
}

func (c *Config) GitHubRedirectURL() string {
	return c.BaseURL() + "/api/v1/auth/oauth/github/callback"
}

func (c *Config) GoogleRedirectURL() string {
	return c.BaseURL() + "/api/v1/auth/oauth/google/callback"
}

func (c *Config) GitLabRedirectURL() string {
	return c.BaseURL() + "/api/v1/auth/oauth/gitlab/callback"
}

func (c *Config) GiteeRedirectURL() string {
	return c.BaseURL() + "/api/v1/auth/oauth/gitee/callback"
}

func (c *Config) AlipayNotifyURL() string {
	return c.BaseURL() + "/api/v1/webhooks/alipay"
}

func (c *Config) LemonSqueezyWebhookURL() string {
	return c.BaseURL() + "/api/v1/webhooks/lemonsqueezy"
}

func (c *Config) AlipayReturnURL() string {
	return c.BaseURL()
}

func (c *Config) WeChatNotifyURL() string {
	return c.BaseURL() + "/api/v1/webhooks/wechat"
}

func (c *Config) AdminFrontendURL() string {
	return c.BaseURL() + "/admin"
}
