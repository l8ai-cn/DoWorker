package imbridge

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	channelDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/channel"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/imbridge"
	channelSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/channel"
)

type contextKey struct{}

var skipOutboundKey = contextKey{}

func WithSkipOutbound(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipOutboundKey, true)
}

func skipOutbound(ctx context.Context) bool {
	v, _ := ctx.Value(skipOutboundKey).(bool)
	return v
}

type ChannelBridge interface {
	SendMessageAsUser(ctx context.Context, channelID, userID int64, content channelDomain.MessageContent) (*channelDomain.Message, error)
	GetChannel(ctx context.Context, channelID int64) (*channelDomain.Channel, error)
	CreateChannel(ctx context.Context, req *channelSvc.CreateChannelRequest) (*channelDomain.Channel, error)
}

type Bridge struct {
	*Service
	channels     ChannelBridge
	weixinLogin  *weixinLoginStore
}

func NewBridge(svc *Service, channels ChannelBridge) *Bridge {
	return &Bridge{Service: svc, channels: channels, weixinLogin: newWeixinLoginStore()}
}

func (b *Bridge) HandleWebhookDeliver(ctx context.Context, provider string, connectionID int64, token string, hdr http.Header, body []byte) (interface{}, error) {
	conn, err := b.connectionForWebhook(ctx, provider, token, connectionID)
	if err != nil {
		return nil, err
	}
	if conn.Status != domain.StatusActive {
		return nil, ErrConnectionPaused
	}
	p, err := GetProvider(b.registry, provider)
	if err != nil {
		return nil, err
	}
	if err := p.VerifyWebhook(ctx, conn.Config, hdr, body); err != nil {
		b.markError(ctx, conn, err.Error())
		return nil, err
	}
	event, err := p.ParseInbound(ctx, conn.Config, hdr, body)
	if err != nil {
		b.markError(ctx, conn, err.Error())
		return nil, err
	}
	if event.Challenge != "" {
		switch provider {
		case domain.ProviderFeishu, domain.ProviderSlack:
			return map[string]string{"challenge": event.Challenge}, nil
		default:
			return event.Challenge, nil
		}
	}
	if strings.TrimSpace(event.Text) == "" {
		return map[string]string{"status": "ignored"}, nil
	}
	if err := b.DeliverInbound(ctx, conn, event); err != nil {
		b.markError(ctx, conn, err.Error())
		return nil, err
	}
	return map[string]string{"status": "delivered"}, nil
}

func (b *Bridge) DeliverInbound(ctx context.Context, conn *domain.Connection, event *InboundEvent) error {
	if conn.Status != domain.StatusActive {
		return ErrConnectionPaused
	}
	if strings.TrimSpace(event.Text) == "" {
		return nil
	}
	channelID, err := b.resolveChannel(ctx, conn, event.ExternalThreadID, event.ContextToken)
	if err != nil {
		return err
	}
	label := event.SenderName
	if label == "" {
		label = conn.Provider
	}
	content := textContent(fmt.Sprintf("[%s] %s", label, event.Text))
	_, err = b.channels.SendMessageAsUser(WithSkipOutbound(ctx), channelID, conn.CreatedByUserID, content)
	return err
}

func (b *Bridge) resolveChannel(ctx context.Context, conn *domain.Connection, threadID, contextToken string) (int64, error) {
	if mapping, err := b.repo.GetThreadMapping(ctx, conn.ID, threadID); err != nil {
		return 0, err
	} else if mapping != nil {
		if contextToken != "" && (mapping.ContextToken == nil || *mapping.ContextToken != contextToken) {
			mapping.ContextToken = strPtrIf(contextToken)
			_ = b.repo.UpsertThreadMapping(ctx, mapping)
		}
		return mapping.ChannelID, nil
	}
	if conn.ChannelID != nil && *conn.ChannelID > 0 {
		return *conn.ChannelID, nil
	}
	name := fmt.Sprintf("im-%s-%s", conn.Provider, sanitizeName(threadID))
	ch, err := b.channels.CreateChannel(ctx, &channelSvc.CreateChannelRequest{
		OrganizationID:  conn.OrganizationID,
		Name:            name,
		Description:     strPtr(fmt.Sprintf("Auto-created IM bridge (%s)", conn.Provider)),
		CreatedByUserID: &conn.CreatedByUserID,
		Visibility:      channelDomain.VisibilityPrivate,
	})
	if err != nil {
		return 0, err
	}
	if err := b.repo.UpsertThreadMapping(ctx, &domain.ThreadMapping{
		ConnectionID:     conn.ID,
		ExternalThreadID: threadID,
		ChannelID:        ch.ID,
		ContextToken:     strPtrIf(contextToken),
	}); err != nil {
		return 0, err
	}
	return ch.ID, nil
}

func (b *Bridge) OutboundHook() channelSvc.PostSendHook {
	return func(ctx context.Context, mc *channelSvc.MessageContext) error {
		if skipOutbound(ctx) || mc == nil || mc.Channel == nil || mc.Message == nil {
			return nil
		}
		body := strings.TrimSpace(mc.Message.Body)
		if body == "" {
			return nil
		}
		conns, err := b.repo.ListConnections(ctx, mc.Channel.OrganizationID)
		if err != nil {
			return err
		}
		for _, conn := range conns {
			if conn.Status != domain.StatusActive {
				continue
			}
			if conn.ChannelID != nil && *conn.ChannelID != mc.Channel.ID {
				continue
			}
			mapping, err := b.repo.GetThreadMappingByChannel(ctx, conn.ID, mc.Channel.ID)
			if err != nil {
				return err
			}
			if conn.ChannelID == nil && mapping == nil {
				continue
			}
			threadID := fmt.Sprintf("channel:%d", mc.Channel.ID)
			contextToken := ""
			if mapping != nil {
				threadID = mapping.ExternalThreadID
				if mapping.ContextToken != nil {
					contextToken = *mapping.ContextToken
				}
			} else if conn.ChannelID != nil {
				// Fixed channel binding without prior inbound message — provider may use default target.
				threadID = ""
			}
			p, err := GetProvider(b.registry, conn.Provider)
			if err != nil {
				continue
			}
			if err := p.SendOutbound(ctx, conn.Config, OutboundMessage{
				ExternalThreadID: threadID,
				Text:             body,
				SenderLabel:      "Agent Cloud",
				ContextToken:     contextToken,
			}); err != nil {
				b.markError(ctx, conn, err.Error())
				continue
			}
		}
		return nil
	}
}

func textContent(text string) channelDomain.MessageContent {
	return channelDomain.MessageContent{
		SchemaVersion: 1,
		Kind:          "text",
		Blocks: []channelDomain.Block{{
			Type: "paragraph",
			Elements: []channelDomain.InlineElement{{
				Type: channelDomain.InlineText,
				Text: text,
			}},
		}},
	}
}

func strPtr(s string) *string { return &s }

func strPtrIf(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func sanitizeName(s string) string {
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, s)
	if len(s) > 40 {
		s = s[:40]
	}
	return strings.Trim(s, "-")
}
