package workercreation

import (
	"context"
	"fmt"
	"strings"

	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

type openAIDraftFillMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIDraftFillRequest struct {
	Model       string                   `json:"model"`
	Messages    []openAIDraftFillMessage `json:"messages"`
	Temperature float64                  `json:"temperature"`
}

type openAIDraftFillResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (generator *ProviderDraftGenerator) generateOpenAICompatible(
	ctx context.Context,
	resource *resourceservice.ResolvedResource,
	apiKey, systemPrompt, userPrompt string,
) ([]byte, error) {
	requestURL, err := joinDraftFillURL(
		resource.Connection.BaseURL,
		"chat",
		"completions",
	)
	if err != nil {
		return nil, err
	}
	body, err := generator.postJSON(
		ctx,
		requestURL,
		map[string]string{"Authorization": "Bearer " + apiKey},
		openAIDraftFillRequest{
			Model: resource.Resource.ModelID,
			Messages: []openAIDraftFillMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			Temperature: 0,
		},
	)
	if err != nil {
		return nil, err
	}
	var response openAIDraftFillResponse
	if err := decodeProviderResponse(body, &response); err != nil {
		return nil, err
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("worker draft fill provider response has no content")
	}
	content := strings.TrimSpace(response.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("worker draft fill provider response has no content")
	}
	return []byte(content), nil
}
