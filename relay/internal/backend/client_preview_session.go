package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrPreviewBootstrapConsumed    = errors.New("preview bootstrap consumed")
	ErrPreviewBootstrapUnavailable = errors.New("preview bootstrap unavailable")
	ErrPreviewSessionUnauthorized  = errors.New("preview session unauthorized")
	ErrPreviewSessionUnavailable   = errors.New("preview session unavailable")
)

type PreviewSessionRegistration struct {
	ID        string
	PodKey    string
	UserID    int64
	OrgID     int64
	ExpiresAt time.Time
}

type PreviewSessionIdentity struct {
	ID     string
	PodKey string
	UserID int64
	OrgID  int64
}

type PreviewBootstrapRedeemRequest struct {
	BootstrapID string    `json:"bootstrap_id"`
	SessionID   string    `json:"session_id"`
	PodKey      string    `json:"pod_key"`
	UserID      int64     `json:"user_id"`
	OrgID       int64     `json:"org_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (c *Client) RedeemPreviewBootstrap(ctx context.Context, bootstrapID string, session PreviewSessionRegistration) error {
	status, err := c.postPreviewSession(ctx, "/api/internal/relays/preview-bootstrap/redeem", PreviewBootstrapRedeemRequest{
		BootstrapID: bootstrapID,
		SessionID:   session.ID,
		PodKey:      session.PodKey,
		UserID:      session.UserID,
		OrgID:       session.OrgID,
		ExpiresAt:   session.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPreviewBootstrapUnavailable, err)
	}
	switch status {
	case http.StatusNoContent:
		return nil
	case http.StatusConflict:
		return ErrPreviewBootstrapConsumed
	default:
		return ErrPreviewBootstrapUnavailable
	}
}

func (c *Client) AuthorizePreviewSession(ctx context.Context, identity PreviewSessionIdentity) error {
	status, err := c.postPreviewSession(ctx, "/api/internal/relays/preview-sessions/authorize", struct {
		SessionID string `json:"session_id"`
		PodKey    string `json:"pod_key"`
		UserID    int64  `json:"user_id"`
		OrgID     int64  `json:"org_id"`
	}{
		SessionID: identity.ID,
		PodKey:    identity.PodKey,
		UserID:    identity.UserID,
		OrgID:     identity.OrgID,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPreviewSessionUnavailable, err)
	}
	switch status {
	case http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return ErrPreviewSessionUnauthorized
	default:
		return ErrPreviewSessionUnavailable
	}
}

func (c *Client) postPreviewSession(ctx context.Context, path string, body any) (int, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", c.internalAPISecret)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer drainBody(response.Body)
	return response.StatusCode, nil
}
