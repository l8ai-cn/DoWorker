package imbridge

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/imbridge"
)

type weixinQRSession struct {
	SessionID      string
	ConnectionID   int64
	OrganizationID int64
	QRCodeValue    string
	QRCodeURL      string
	BaseURL        string
	RefreshCount   int
	DeadlineUnix   int64
}

type weixinLoginStore struct {
	mu       sync.Mutex
	sessions map[string]*weixinQRSession
}

func newWeixinLoginStore() *weixinLoginStore {
	return &weixinLoginStore{sessions: make(map[string]*weixinQRSession)}
}

const (
	weixinQRTimeoutSeconds = 480
	maxWeixinQRRefreshes   = 3
)

func (b *Bridge) StartWeixinQRLogin(ctx context.Context, orgID, connectionID int64) (map[string]any, error) {
	conn, err := b.GetConnection(ctx, orgID, connectionID)
	if err != nil {
		return nil, err
	}
	if conn.Provider != domain.ProviderWeixin && conn.Provider != domain.ProviderWeChat {
		return nil, fmt.Errorf("connection is not a weixin provider")
	}
	p, err := GetProvider(b.registry, domain.ProviderWeixin)
	if err != nil {
		return nil, err
	}
	wp, ok := p.(*WeixinProvider)
	if !ok {
		return nil, fmt.Errorf("weixin provider unavailable")
	}
	qr, err := wp.ilink().fetchBotQR(ctx, "3")
	if err != nil {
		return nil, err
	}
	sessionID := fmt.Sprintf("wxqr-%d-%d", connectionID, time.Now().UnixNano())
	expiresAt := time.Now().Add(weixinQRTimeoutSeconds * time.Second).Unix()
	b.weixinLogin.mu.Lock()
	b.weixinLogin.sessions[sessionID] = &weixinQRSession{
		SessionID:      sessionID,
		ConnectionID:   connectionID,
		OrganizationID: orgID,
		QRCodeValue:    qr.QRCodeValue,
		QRCodeURL:      qr.QRCodeURL,
		BaseURL:        defaultIlinkBaseURL,
		DeadlineUnix:   expiresAt,
	}
	b.weixinLogin.mu.Unlock()
	return map[string]any{
		"session_id":         sessionID,
		"status":             "wait",
		"qrcode":             qr.QRCodeValue,
		"qrcode_url":         qr.QRCodeURL,
		"expires_at":         expiresAt,
		"poll_interval_ms":   1000,
		"provider":           "ilink",
		"connection_id":      connectionID,
	}, nil
}

func (b *Bridge) PollWeixinQRLogin(ctx context.Context, orgID int64, sessionID string) (map[string]any, error) {
	b.weixinLogin.mu.Lock()
	session, ok := b.weixinLogin.sessions[sessionID]
	if !ok {
		b.weixinLogin.mu.Unlock()
		return nil, ErrNotFound
	}
	if session.OrganizationID != orgID {
		b.weixinLogin.mu.Unlock()
		return nil, ErrNotFound
	}
	snapshot := *session
	b.weixinLogin.mu.Unlock()

	if time.Now().Unix() >= snapshot.DeadlineUnix {
		b.weixinLogin.mu.Lock()
		delete(b.weixinLogin.sessions, sessionID)
		b.weixinLogin.mu.Unlock()
		return map[string]any{
			"session_id": sessionID,
			"status":     "timed_out",
			"message":    "二维码登录已超时，请重新开始",
		}, nil
	}

	p, err := GetProvider(b.registry, domain.ProviderWeixin)
	if err != nil {
		return nil, err
	}
	wp := p.(*WeixinProvider)
	payload, err := wp.ilink().pollQRStatus(ctx, snapshot.BaseURL, snapshot.QRCodeValue)
	if err != nil {
		return nil, err
	}
	status, _ := payload["status"].(string)
	status = strings.TrimSpace(status)

	switch status {
	case "wait":
		return map[string]any{
			"session_id":  sessionID,
			"status":      "wait",
			"qrcode":      snapshot.QRCodeValue,
			"qrcode_url":  snapshot.QRCodeURL,
			"expires_at":  snapshot.DeadlineUnix,
		}, nil
	case "scaned", "scaned_but_redirect":
		if host, ok := payload["redirect_host"].(string); ok && strings.TrimSpace(host) != "" {
			b.weixinLogin.mu.Lock()
			if s, ok := b.weixinLogin.sessions[sessionID]; ok {
				s.BaseURL = "https://" + strings.TrimSpace(host)
			}
			b.weixinLogin.mu.Unlock()
		}
		return map[string]any{
			"session_id": sessionID,
			"status":     "scanned",
			"message":    "已扫码，请在微信中确认登录",
			"qrcode_url": snapshot.QRCodeURL,
			"expires_at": snapshot.DeadlineUnix,
		}, nil
	case "expired":
		return b.refreshWeixinQR(ctx, orgID, sessionID, &snapshot, wp)
	case "confirmed":
		return b.completeWeixinQRLogin(ctx, orgID, sessionID, &snapshot, payload)
	default:
		return map[string]any{
			"session_id": sessionID,
			"status":     status,
			"message":    "iLink 返回了未识别的状态",
		}, nil
	}
}

func (b *Bridge) refreshWeixinQR(ctx context.Context, orgID int64, sessionID string, snapshot *weixinQRSession, wp *WeixinProvider) (map[string]any, error) {
	if snapshot.RefreshCount >= maxWeixinQRRefreshes {
		b.weixinLogin.mu.Lock()
		delete(b.weixinLogin.sessions, sessionID)
		b.weixinLogin.mu.Unlock()
		return map[string]any{
			"session_id": sessionID,
			"status":     "failed",
			"message":    "二维码多次过期，请重新开始登录",
		}, nil
	}
	qr, err := wp.ilink().fetchBotQR(ctx, "3")
	if err != nil {
		return nil, err
	}
	b.weixinLogin.mu.Lock()
	if s, ok := b.weixinLogin.sessions[sessionID]; ok {
		s.QRCodeValue = qr.QRCodeValue
		s.QRCodeURL = qr.QRCodeURL
		s.BaseURL = defaultIlinkBaseURL
		s.RefreshCount++
	}
	b.weixinLogin.mu.Unlock()
	return map[string]any{
		"session_id": sessionID,
		"status":     "wait",
		"message":    "二维码已刷新，请重新扫码",
		"qrcode":     qr.QRCodeValue,
		"qrcode_url": qr.QRCodeURL,
		"expires_at": snapshot.DeadlineUnix,
		"refreshed":  true,
	}, nil
}

func (b *Bridge) completeWeixinQRLogin(ctx context.Context, orgID int64, sessionID string, snapshot *weixinQRSession, payload map[string]any) (map[string]any, error) {
	accountID, _ := payload["ilink_bot_id"].(string)
	token, _ := payload["bot_token"].(string)
	baseURL, _ := payload["baseurl"].(string)
	userID, _ := payload["ilink_user_id"].(string)
	accountID = strings.TrimSpace(accountID)
	token = strings.TrimSpace(token)
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultIlinkBaseURL
	}
	if accountID == "" || token == "" {
		return nil, fmt.Errorf("ilink confirmed login but credential payload was incomplete")
	}
	conn, err := b.GetConnection(ctx, orgID, snapshot.ConnectionID)
	if err != nil {
		return nil, err
	}
	merged, err := mergeWeixinConfig(conn.Config, weixinBridgeConfig{
		AccountID: accountID,
		BotToken:  token,
		BaseURL:   baseURL,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}
	conn.Config = merged
	if err := b.repo.UpdateConnection(ctx, conn); err != nil {
		return nil, err
	}
	b.weixinLogin.mu.Lock()
	delete(b.weixinLogin.sessions, sessionID)
	b.weixinLogin.mu.Unlock()
	slog.InfoContext(ctx, "weixin ilink login confirmed", "connection_id", conn.ID, "account_id", accountID)
	return map[string]any{
		"session_id":    sessionID,
		"status":        "confirmed",
		"connection_id": conn.ID,
		"account_id":    accountID,
	}, nil
}

func (b *Bridge) GetWeixinQRImage(sessionID string) (mediaType string, data []byte, err error) {
	b.weixinLogin.mu.Lock()
	session, ok := b.weixinLogin.sessions[sessionID]
	b.weixinLogin.mu.Unlock()
	if !ok {
		return "", nil, ErrNotFound
	}
	src := strings.TrimSpace(session.QRCodeURL)
	if src == "" {
		return "", nil, fmt.Errorf("qr image unavailable")
	}
	if strings.HasPrefix(src, "data:image/") {
		parts := strings.SplitN(src, ",", 2)
		if len(parts) != 2 {
			return "", nil, fmt.Errorf("invalid data url")
		}
		header := strings.TrimPrefix(parts[0], "data:")
		mediaType = strings.TrimSuffix(header, ";base64")
		data, err = decodeBase64(parts[1])
		return mediaType, data, err
	}
	return "", nil, fmt.Errorf("remote qr image fetch not implemented")
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
