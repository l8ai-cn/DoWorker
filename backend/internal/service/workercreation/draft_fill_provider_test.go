package workercreation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	resourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderDraftGeneratorSupportsWorkerModelProtocols(t *testing.T) {
	tests := []struct {
		name     string
		adapter  string
		alias    string
		path     string
		response string
		assert   func(*testing.T, *http.Request, map[string]any)
	}{
		{
			name: "openai compatible", adapter: "openai-compatible",
			alias:    "openai-worker",
			path:     "/chat/completions",
			response: `{"choices":[{"message":{"content":"{\"alias\":\"openai-worker\"}"}}]}`,
			assert: func(t *testing.T, request *http.Request, body map[string]any) {
				assert.Equal(t, "Bearer secret", request.Header.Get("Authorization"))
				assert.Equal(t, "model-v1", body["model"])
			},
		},
		{
			name: "anthropic", adapter: "anthropic",
			alias:    "anthropic-worker",
			path:     "/v1/messages",
			response: `{"content":[{"type":"text","text":"{\"alias\":\"anthropic-worker\"}"}]}`,
			assert: func(t *testing.T, request *http.Request, body map[string]any) {
				assert.Equal(t, "secret", request.Header.Get("x-api-key"))
				assert.Equal(t, "2023-06-01", request.Header.Get("anthropic-version"))
				assert.Equal(t, "model-v1", body["model"])
			},
		},
		{
			name: "gemini", adapter: "gemini",
			alias:    "gemini-worker",
			path:     "/v1beta/models/model-v1:generateContent",
			response: `{"candidates":[{"content":{"parts":[{"text":"{\"alias\":\"gemini-worker\"}"}]}}]}`,
			assert: func(t *testing.T, request *http.Request, body map[string]any) {
				assert.Equal(t, "secret", request.URL.Query().Get("key"))
				assert.Contains(t, body, "generationConfig")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				assert.Equal(t, test.path, request.URL.Path)
				var body map[string]any
				require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
				test.assert(t, request, body)
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(test.response))
			}))
			defer server.Close()
			resource := draftFillResource(test.adapter, server.URL)

			output, err := NewProviderDraftGenerator(server.Client()).Generate(
				context.Background(),
				resource,
				"system prompt",
				"user prompt",
			)

			require.NoError(t, err)
			assert.JSONEq(t, `{"alias":"`+test.alias+`"}`, string(output))
		})
	}
}

func TestProviderDraftGeneratorRejectsProviderFailuresWithoutLeakingBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"error":"provider-secret"}`))
	}))
	defer server.Close()

	_, err := NewProviderDraftGenerator(server.Client()).Generate(
		context.Background(),
		draftFillResource("openai-compatible", server.URL),
		"system prompt",
		"user prompt",
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "status 400")
	assert.NotContains(t, err.Error(), "provider-secret")
}

func TestProviderDraftGeneratorRejectsMalformedProviderResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		_, _ = writer.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	_, err := NewProviderDraftGenerator(server.Client()).Generate(
		context.Background(),
		draftFillResource("openai-compatible", server.URL),
		"system prompt",
		"user prompt",
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "content")
}

func TestProviderDraftGeneratorRedactsTransportErrors(t *testing.T) {
	_, err := NewProviderDraftGenerator(draftFillErrorDoer{
		err: errors.New("request failed for https://provider.example?key=secret"),
	}).Generate(
		context.Background(),
		draftFillResource("gemini", "https://provider.example"),
		"system prompt",
		"user prompt",
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "provider request failed")
	assert.NotContains(t, err.Error(), "secret")
	assert.NotContains(t, err.Error(), "provider.example")
}

type draftFillErrorDoer struct {
	err error
}

func (doer draftFillErrorDoer) Do(*http.Request) (*http.Response, error) {
	return nil, doer.err
}

func draftFillResource(
	adapter, baseURL string,
) *resourceservice.ResolvedResource {
	return &resourceservice.ResolvedResource{
		Provider: resourcedomain.ProviderDefinition{
			Key:             slugkit.MustNewForTest("test-provider"),
			ProtocolAdapter: adapter,
		},
		Connection: resourcedomain.Connection{
			ID:          201,
			ProviderKey: slugkit.MustNewForTest("test-provider"),
			BaseURL:     baseURL,
			Revision:    9,
		},
		Resource: resourcedomain.ModelResource{
			ID:                   101,
			ProviderConnectionID: 201,
			ModelID:              "model-v1",
			Revision:             7,
		},
		Credentials: map[string]string{"api_key": "secret"},
	}
}
