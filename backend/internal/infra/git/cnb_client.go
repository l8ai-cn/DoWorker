package git

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type CNBProvider struct {
	apiBaseURL  string
	webBaseURL  string
	accessToken string
	httpClient  *http.Client
}

func NewCNBProvider(baseURL, accessToken string) (*CNBProvider, error) {
	if baseURL == "" {
		baseURL = "https://cnb.cool"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	apiBaseURL := baseURL
	webBaseURL := baseURL
	switch baseURL {
	case "https://cnb.cool", "http://cnb.cool":
		apiBaseURL = "https://api.cnb.cool"
	case "https://api.cnb.cool", "http://api.cnb.cool":
		webBaseURL = strings.Replace(baseURL, "://api.", "://", 1)
	default:
		apiBaseURL = strings.TrimSuffix(baseURL, "/")
	}

	return &CNBProvider{
		apiBaseURL:  apiBaseURL,
		webBaseURL:  webBaseURL,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}, nil
}

func (p *CNBProvider) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	reqURL := p.apiBaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		resp.Body.Close()
		return nil, ErrUnauthorized
	case http.StatusNotFound:
		resp.Body.Close()
		return nil, ErrNotFound
	case http.StatusTooManyRequests:
		resp.Body.Close()
		return nil, ErrRateLimited
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("CNB API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
