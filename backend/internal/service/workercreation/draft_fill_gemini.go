package workercreation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	resourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
)

type geminiDraftFillPart struct {
	Text string `json:"text"`
}

type geminiDraftFillContent struct {
	Role  string                `json:"role,omitempty"`
	Parts []geminiDraftFillPart `json:"parts"`
}

type geminiDraftFillRequest struct {
	SystemInstruction geminiDraftFillContent   `json:"systemInstruction"`
	Contents          []geminiDraftFillContent `json:"contents"`
	GenerationConfig  struct {
		Temperature      float64 `json:"temperature"`
		ResponseMimeType string  `json:"responseMimeType"`
	} `json:"generationConfig"`
}

type geminiDraftFillResponse struct {
	Candidates []struct {
		Content geminiDraftFillContent `json:"content"`
	} `json:"candidates"`
}

func (generator *ProviderDraftGenerator) generateGemini(
	ctx context.Context,
	resource *resourceservice.ResolvedResource,
	apiKey, systemPrompt, userPrompt string,
) ([]byte, error) {
	modelID := strings.TrimPrefix(resource.Resource.ModelID, "models/")
	if modelID == "" || strings.Contains(modelID, "/") {
		return nil, fmt.Errorf("worker draft fill Gemini model ID is invalid")
	}
	requestURL, err := joinDraftFillURL(
		resource.Connection.BaseURL,
		"v1beta",
		"models",
		url.PathEscape(modelID)+":generateContent",
	)
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(requestURL)
	if err != nil {
		return nil, fmt.Errorf("build worker draft fill Gemini URL: %w", err)
	}
	query := parsed.Query()
	query.Set("key", apiKey)
	parsed.RawQuery = query.Encode()
	request := geminiDraftFillRequest{
		SystemInstruction: geminiDraftFillContent{
			Parts: []geminiDraftFillPart{{Text: systemPrompt}},
		},
		Contents: []geminiDraftFillContent{{
			Role:  "user",
			Parts: []geminiDraftFillPart{{Text: userPrompt}},
		}},
	}
	request.GenerationConfig.Temperature = 0
	request.GenerationConfig.ResponseMimeType = "application/json"
	body, err := generator.postJSON(ctx, parsed.String(), nil, request)
	if err != nil {
		return nil, err
	}
	var response geminiDraftFillResponse
	if err := decodeProviderResponse(body, &response); err != nil {
		return nil, err
	}
	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("worker draft fill provider response has no content")
	}
	var content strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		content.WriteString(part.Text)
	}
	result := strings.TrimSpace(content.String())
	if result == "" {
		return nil, fmt.Errorf("worker draft fill provider response has no content")
	}
	return []byte(result), nil
}
