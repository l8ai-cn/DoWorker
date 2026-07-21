package airesource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
)

type HTTPDoer interface {
	Do(request *http.Request) (*http.Response, error)
}

type HTTPConnectionProber struct{ client HTTPDoer }

func NewHTTPConnectionProber(client HTTPDoer) (*HTTPConnectionProber, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP client is required")
	}
	return &HTTPConnectionProber{client: client}, nil
}

func (prober *HTTPConnectionProber) Probe(ctx context.Context, input ProbeInput) error {
	if input.Provider.ConnectionCheck.AuthStrategy == domain.ConnectionAuthUnsupported {
		return ErrProbeUnsupported
	}
	requestURL, err := probeURL(input.BaseURL, input.Provider.ConnectionCheck.Path)
	if err != nil {
		return ErrInvalidEndpoint
	}
	request, err := http.NewRequestWithContext(ctx, input.Provider.ConnectionCheck.Method, requestURL, nil)
	if err != nil {
		return ErrValidation
	}
	if err := prober.applyAuthentication(request, input.Provider.ConnectionCheck, input.Credentials); err != nil {
		return err
	}
	response, err := prober.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: provider request failed", ErrValidation)
	}
	defer response.Body.Close()
	_, _ = io.CopyN(io.Discard, response.Body, 4096)
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return ErrInvalidCredentials
	}
	if response.StatusCode == http.StatusNotFound {
		return ErrProviderEndpointUnavailable
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%w: provider status %d", ErrValidation, response.StatusCode)
	}
	return nil
}

func probeURL(baseURL, checkPath string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", ErrInvalidEndpoint
	}
	joined, err := url.JoinPath(strings.TrimRight(baseURL, "/"), checkPath)
	if err != nil {
		return "", err
	}
	return joined, nil
}

func (prober *HTTPConnectionProber) applyAuthentication(request *http.Request, check domain.ConnectionCheck, credentials map[string]string) error {
	credential := strings.TrimSpace(credentials[check.CredentialKey])
	if credential == "" {
		return ErrInvalidCredentials
	}
	for _, header := range check.StaticHeaders {
		request.Header.Set(header.Name, header.Value)
	}
	switch check.AuthStrategy {
	case domain.ConnectionAuthBearer:
		request.Header.Set("Authorization", "Bearer "+credential)
	case domain.ConnectionAuthHeader:
		if check.AuthName == "" {
			return ErrValidation
		}
		request.Header.Set(check.AuthName, credential)
	case domain.ConnectionAuthQuery:
		if check.AuthName == "" {
			return ErrValidation
		}
		query := request.URL.Query()
		query.Set(check.AuthName, credential)
		request.URL.RawQuery = query.Encode()
	default:
		return ErrValidation
	}
	return nil
}
