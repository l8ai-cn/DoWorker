package workercreation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	resourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

const maxDraftFillProviderResponseBytes = 1 << 20

type ProviderDraftGenerator struct {
	client resourceservice.HTTPDoer
}

func NewProviderDraftGenerator(
	client resourceservice.HTTPDoer,
) *ProviderDraftGenerator {
	return &ProviderDraftGenerator{client: client}
}

func (generator *ProviderDraftGenerator) Generate(
	ctx context.Context,
	resource *resourceservice.ResolvedResource,
	systemPrompt, userPrompt string,
) ([]byte, error) {
	if generator == nil || generator.client == nil || resource == nil {
		return nil, specservice.ErrResolverUnavailable
	}
	if strings.TrimSpace(resource.Connection.BaseURL) == "" ||
		strings.TrimSpace(resource.Resource.ModelID) == "" {
		return nil, fmt.Errorf("worker draft fill model endpoint is incomplete")
	}
	if resource.Provider.Key != resource.Connection.ProviderKey {
		return nil, fmt.Errorf("worker draft fill provider does not match connection")
	}
	apiKey := strings.TrimSpace(resource.Credentials["api_key"])
	if apiKey == "" {
		return nil, resourceservice.ErrInvalidCredentials
	}
	switch resource.Provider.ProtocolAdapter {
	case "openai-compatible":
		return generator.generateOpenAICompatible(
			ctx,
			resource,
			apiKey,
			systemPrompt,
			userPrompt,
		)
	case "anthropic":
		return generator.generateAnthropic(
			ctx,
			resource,
			apiKey,
			systemPrompt,
			userPrompt,
		)
	case "gemini":
		return generator.generateGemini(
			ctx,
			resource,
			apiKey,
			systemPrompt,
			userPrompt,
		)
	default:
		return nil, fmt.Errorf(
			"worker draft fill protocol %q is unsupported",
			resource.Provider.ProtocolAdapter,
		)
	}
}

func (generator *ProviderDraftGenerator) postJSON(
	ctx context.Context,
	requestURL string,
	headers map[string]string,
	payload any,
) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode worker draft fill request: %w", err)
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, errors.New("create worker draft fill request failed")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := generator.client.Do(request)
	if err != nil {
		return nil, errors.New("worker draft fill provider request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"worker draft fill provider returned status %d",
			response.StatusCode,
		)
	}
	limited := io.LimitReader(
		response.Body,
		maxDraftFillProviderResponseBytes+1,
	)
	encoded, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read worker draft fill response: %w", err)
	}
	if len(encoded) > maxDraftFillProviderResponseBytes {
		return nil, fmt.Errorf("worker draft fill provider response is too large")
	}
	return encoded, nil
}

func joinDraftFillURL(baseURL string, pathSegments ...string) (string, error) {
	joined, err := url.JoinPath(
		strings.TrimRight(baseURL, "/"),
		pathSegments...,
	)
	if err != nil {
		return "", fmt.Errorf("build worker draft fill URL: %w", err)
	}
	return joined, nil
}

func decodeProviderResponse(body []byte, target any) error {
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode worker draft fill provider response: %w", err)
	}
	return nil
}
