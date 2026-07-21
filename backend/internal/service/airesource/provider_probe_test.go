package airesource

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureDoer struct {
	request *http.Request
	status  int
	body    string
	err     error
}

func (d *captureDoer) Do(request *http.Request) (*http.Response, error) {
	d.request = request
	if d.err != nil {
		return nil, d.err
	}
	return &http.Response{StatusCode: d.status, Body: io.NopCloser(strings.NewReader(d.body)), Header: make(http.Header)}, nil
}

func TestHTTPProberBuildsRegistryOwnedAuthentication(t *testing.T) {
	tests := []struct {
		provider      string
		credentials   map[string]string
		assertRequest func(*testing.T, *http.Request)
	}{
		{"openai", map[string]string{"api_key": "openai-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer openai-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/v1/models", r.URL.Path)
		}},
		{"openrouter", map[string]string{"api_key": "openrouter-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer openrouter-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/api/v1/key", r.URL.Path)
		}},
		{"deepseek", map[string]string{"api_key": "deepseek-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer deepseek-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/models", r.URL.Path)
		}},
		{"xai", map[string]string{"api_key": "xai-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer xai-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/v1/models", r.URL.Path)
		}},
		{"mistral", map[string]string{"api_key": "mistral-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer mistral-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/v1/models", r.URL.Path)
		}},
		{"anthropic", map[string]string{"api_key": "anthropic-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "anthropic-secret", r.Header.Get("x-api-key"))
			assert.NotEmpty(t, r.Header.Get("anthropic-version"))
		}},
		{"gemini", map[string]string{"api_key": "gemini-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "gemini-secret", r.URL.Query().Get("key"))
			assert.Empty(t, r.Header.Get("Authorization"))
		}},
		{"azure-openai", map[string]string{"api_key": "azure-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "azure-secret", r.Header.Get("api-key"))
			assert.Equal(t, "/openai/v1/models", r.URL.Path)
		}},
		{"stability-ai", map[string]string{"api_key": "stability-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer stability-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/v1/user/account", r.URL.Path)
		}},
		{"black-forest-labs", map[string]string{"api_key": "bfl-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "bfl-secret", r.Header.Get("x-key"))
			assert.Equal(t, "/v1/credits", r.URL.Path)
		}},
		{"elevenlabs", map[string]string{"api_key": "eleven-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "eleven-secret", r.Header.Get("xi-api-key"))
			assert.Equal(t, "/v1/models", r.URL.Path)
		}},
		{"runway", map[string]string{"api_key": "runway-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer runway-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "2024-11-06", r.Header.Get("X-Runway-Version"))
			assert.Equal(t, "/v1/organization", r.URL.Path)
		}},
		{"luma", map[string]string{"api_key": "luma-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer luma-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/dream-machine/v1/generations", r.URL.Path)
		}},
		{"ideogram", map[string]string{"api_key": "ideogram-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "ideogram-secret", r.Header.Get("Api-Key"))
			assert.Equal(t, "/models", r.URL.Path)
		}},
		{"replicate", map[string]string{"api_token": "replicate-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer replicate-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/v1/models", r.URL.Path)
		}},
		{"doubao", map[string]string{"api_key": "ark-secret"}, func(t *testing.T, r *http.Request) {
			assert.Equal(t, "Bearer ark-secret", r.Header.Get("Authorization"))
			assert.Equal(t, "/api/v3/contents/generations/tasks", r.URL.Path)
		}},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			definition, ok := domain.Provider(test.provider)
			require.True(t, ok)
			baseURL := definition.DefaultBaseURL
			if test.provider == "azure-openai" {
				baseURL = "https://resource.openai.azure.com"
			}
			doer := &captureDoer{status: http.StatusOK}
			prober, err := NewHTTPConnectionProber(doer)
			require.NoError(t, err)
			err = prober.Probe(context.Background(), ProbeInput{Provider: definition, BaseURL: baseURL, Credentials: test.credentials})
			require.NoError(t, err)
			test.assertRequest(t, doer.request)
		})
	}
}

func TestHTTPProberDoesNotCallNetworkForUnsupportedProviders(t *testing.T) {
	keys := []string{"fal", "kling", "hailuo", "dashscope", "azure-speech", "custom-anthropic-compatible"}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			definition, ok := domain.Provider(key)
			require.True(t, ok)
			doer := &captureDoer{status: http.StatusOK}
			prober, err := NewHTTPConnectionProber(doer)
			require.NoError(t, err)
			err = prober.Probe(context.Background(), ProbeInput{Provider: definition, BaseURL: definition.DefaultBaseURL, Credentials: map[string]string{"api_key": "secret", "access_key": "access", "secret_key": "secret", "subscription_key": "secret", "region": "region"}})
			assert.ErrorIs(t, err, ErrProbeUnsupported)
			assert.Nil(t, doer.request)
		})
	}
}

func TestHTTPProberMapsStatusAndTransportErrors(t *testing.T) {
	definition, _ := domain.Provider("openai")
	tests := []struct {
		name     string
		doer     *captureDoer
		expected error
	}{
		{"unauthorized", &captureDoer{status: http.StatusUnauthorized}, ErrInvalidCredentials},
		{"forbidden", &captureDoer{status: http.StatusForbidden}, ErrInvalidCredentials},
		{"missing endpoint", &captureDoer{status: http.StatusNotFound}, ErrProviderEndpointUnavailable},
		{"provider failure", &captureDoer{status: http.StatusBadGateway, body: strings.Repeat("secret", 10_000)}, ErrValidation},
		{"transport", &captureDoer{err: errInjected}, ErrValidation},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prober, err := NewHTTPConnectionProber(test.doer)
			require.NoError(t, err)
			err = prober.Probe(context.Background(), ProbeInput{Provider: definition, BaseURL: definition.DefaultBaseURL, Credentials: map[string]string{"api_key": "secret"}})
			assert.ErrorIs(t, err, test.expected)
			assert.NotContains(t, err.Error(), "secret")
		})
	}
}

func TestHTTPProberRejectsMissingPolicyCredential(t *testing.T) {
	definition, _ := domain.Provider("openai")
	prober, err := NewHTTPConnectionProber(&captureDoer{status: 200})
	require.NoError(t, err)
	err = prober.Probe(context.Background(), ProbeInput{Provider: definition, BaseURL: definition.DefaultBaseURL, Credentials: map[string]string{}})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestSafeHTTPClientRejectsAllRedirects(t *testing.T) {
	client := NewSafeHTTPClient(NewEndpointPolicy(false, staticResolver{}), nil)
	request, _ := http.NewRequest(http.MethodGet, "https://other.example", nil)
	err := client.CheckRedirect(request, []*http.Request{{}})
	assert.ErrorIs(t, err, ErrInvalidEndpoint)
}

func TestHTTPProberConstructorRejectsNilDependencies(t *testing.T) {
	prober, err := NewHTTPConnectionProber(nil)
	assert.Nil(t, prober)
	assert.Error(t, err)
	assert.True(t, errors.Is(ErrInvalidCredentials, ErrInvalidCredentials))
}
