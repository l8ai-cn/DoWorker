package imbridge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/imbridge"
)

func (b *Bridge) StartMonitor(ctx context.Context) {
	go b.runWeixinMonitor(ctx)
}

func (b *Bridge) runWeixinMonitor(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.pollActiveWeixinConnections(ctx)
		}
	}
}

func (b *Bridge) pollActiveWeixinConnections(ctx context.Context) {
	conns, err := b.repo.ListActiveByProvider(ctx, domain.ProviderWeixin)
	if err != nil {
		slog.WarnContext(ctx, "weixin monitor list failed", "error", err)
		return
	}
	wechatConns, err := b.repo.ListActiveByProvider(ctx, domain.ProviderWeChat)
	if err == nil {
		conns = append(conns, wechatConns...)
	}
	for _, conn := range conns {
		if err := b.pollWeixinConnection(ctx, conn); err != nil {
			slog.WarnContext(ctx, "weixin poll failed", "connection_id", conn.ID, "error", err)
		}
	}
}

func (b *Bridge) pollWeixinConnection(ctx context.Context, conn *domain.Connection) error {
	cfg, err := parseWeixinConfig(conn.Config)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.BotToken) == "" {
		return nil
	}
	p, err := GetProvider(b.registry, domain.ProviderWeixin)
	if err != nil {
		return err
	}
	wp, ok := p.(*WeixinProvider)
	if !ok {
		return fmt.Errorf("weixin provider unavailable")
	}
	updates, err := wp.ilink().getUpdates(ctx, cfg)
	if err != nil {
		b.markError(ctx, conn, err.Error())
		return err
	}
	if updates.ErrCode == -14 {
		b.markError(ctx, conn, "weixin session expired, please re-login")
		return nil
	}
	if updates.GetUpdatesBuf != "" && updates.GetUpdatesBuf != cfg.GetUpdatesBuf {
		merged, err := mergeWeixinConfig(conn.Config, weixinBridgeConfig{GetUpdatesBuf: updates.GetUpdatesBuf})
		if err != nil {
			return err
		}
		conn.Config = merged
		if err := b.repo.UpdateConnection(ctx, conn); err != nil {
			return err
		}
	}
	for _, msg := range updates.Messages {
		event := parseWeixinInbound(msg)
		if event == nil {
			continue
		}
		if err := b.DeliverInbound(ctx, conn, event); err != nil {
			b.markError(ctx, conn, err.Error())
		}
	}
	return nil
}
