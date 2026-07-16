package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
)

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

func NewClient(baseURL, secret string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" ||
		strings.TrimSpace(secret) == "" ||
		httpClient == nil {
		return nil, errors.New("runtime bridge URL, secret, and HTTP client are required")
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  secret,
		http:    httpClient,
	}, nil
}

func (c *Client) Authorize(
	ctx context.Context,
	targetOrganizationID int64,
	actorUserID int64,
) error {
	body, err := json.Marshal(map[string]int64{
		"target_platform_organization_id": targetOrganizationID,
		"actor_platform_user_id":          actorUserID,
	})
	if err != nil {
		return service.ErrRuntimeAuthorizationFailed
	}
	response, err := c.doPost(ctx, c.baseURL+"/authorize", body)
	if err != nil {
		return fmt.Errorf("%w: %v", service.ErrRuntimeAuthorizationFailed, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusForbidden {
		return service.ErrTargetOrganizationForbidden
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf(
			"%w: HTTP %d",
			service.ErrRuntimeAuthorizationFailed,
			response.StatusCode,
		)
	}
	return nil
}

func (c *Client) Install(
	ctx context.Context,
	request service.RuntimeInstallRequest,
) (service.RuntimeInstallResult, error) {
	body, err := json.Marshal(map[string]any{
		"installation_id":                 request.InstallationID,
		"platform_resource_type":          request.PlatformResourceType,
		"platform_resource_id":            request.PlatformResourceID,
		"source_release_id":               request.SourceReleaseID,
		"runtime_snapshot":                request.RuntimeSnapshot,
		"target_platform_organization_id": request.TargetOrganizationID,
		"actor_platform_user_id":          request.ActorUserID,
		"configuration":                   request.Configuration,
	})
	if err != nil {
		return service.RuntimeInstallResult{}, service.ErrRuntimeInstallationRejected
	}
	response, err := c.doPost(ctx, c.baseURL+"/apply", body)
	if err != nil {
		return service.RuntimeInstallResult{}, fmt.Errorf(
			"%w: %v",
			service.ErrRuntimeInstallationUnknown,
			err,
		)
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return service.RuntimeInstallResult{}, service.ErrRuntimeInstallationUnknown
	}
	if response.StatusCode != http.StatusOK {
		category := service.ErrRuntimeInstallationUnknown
		if response.StatusCode >= 400 && response.StatusCode < 500 {
			category = service.ErrRuntimeInstallationRejected
		}
		return service.RuntimeInstallResult{}, fmt.Errorf(
			"%w: HTTP %d",
			category,
			response.StatusCode,
		)
	}
	var result struct {
		RuntimeRef string          `json:"runtime_ref"`
		Result     json.RawMessage `json:"result"`
	}
	if json.Unmarshal(payload, &result) != nil || strings.TrimSpace(result.RuntimeRef) == "" {
		return service.RuntimeInstallResult{}, service.ErrRuntimeInstallationUnknown
	}
	return service.RuntimeInstallResult{
		RuntimeRef: result.RuntimeRef,
		Result:     result.Result,
	}, nil
}

func (c *Client) doPost(
	ctx context.Context,
	endpoint string,
	body []byte,
) (*http.Response, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", c.secret)
	return c.http.Do(request)
}
