package channel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	channelDomain "github.com/anthropics/agentsmesh/backend/internal/domain/channel"
)

const podMentionTextLen = 8

type PodPromptRouter interface {
	RoutePrompt(podKey string, prompt string) error
}

type OfflinePromptQueuer interface {
	QueueChannelPrompt(ctx context.Context, podKey, prompt string, messageID int64) error
}

type SystemMessageWriter interface {
	CreateMessage(ctx context.Context, msg *channelDomain.Message) error
}

func NewPodPromptHook(router PodPromptRouter, msgWriter SystemMessageWriter, queuer OfflinePromptQueuer) PostSendHook {
	return func(ctx context.Context, mc *MessageContext) error {
		if router == nil || mc.Mentions == nil || len(mc.Mentions.PodKeys) == 0 {
			return nil
		}

		prompt := buildPodPrompt(mc.Message.Body, mc.Channel.Name, mc.Channel.ID, mc.Mentions.PodKeys)

		for _, podKey := range mc.Mentions.PodKeys {
			if mc.Message.SenderPod != nil && *mc.Message.SenderPod == podKey {
				continue
			}

			if err := router.RoutePrompt(podKey, prompt); err != nil {
				if queuer != nil {
					if qErr := queuer.QueueChannelPrompt(ctx, podKey, prompt, mc.Message.ID); qErr == nil {
						writeQueuedNotice(ctx, msgWriter, mc.Message.ChannelID, podKey)
						continue
					}
				}
				slog.WarnContext(ctx, "pod unreachable for prompt",
					"pod_key", podKey,
					"channel", mc.Channel.Name,
					"error", err,
				)
				writeOfflineNotice(ctx, msgWriter, mc.Message.ChannelID, podKey)
				continue
			}
		}

		return nil
	}
}

func writeQueuedNotice(ctx context.Context, w SystemMessageWriter, channelID int64, podKey string) {
	if w == nil {
		return
	}
	msg := &channelDomain.Message{
		ChannelID:   channelID,
		MessageType: channelDomain.MessageTypeSystem,
		Body:        fmt.Sprintf("@%s is offline, message queued for delivery.", podKey),
	}
	if err := w.CreateMessage(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "failed to write pod-queued system message", "error", err)
	}
}

func writeOfflineNotice(ctx context.Context, w SystemMessageWriter, channelID int64, podKey string) {
	if w == nil {
		return
	}
	msg := &channelDomain.Message{
		ChannelID:   channelID,
		MessageType: channelDomain.MessageTypeSystem,
		Body:        fmt.Sprintf("@%s is offline and cannot receive this message.", podKey),
	}
	if err := w.CreateMessage(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "failed to write pod-offline system message", "error", err)
	}
}

func stripPodMentions(content string, podKeys []string) string {
	result := content
	for _, key := range podKeys {
		mention := key
		if len(mention) > podMentionTextLen {
			mention = mention[:podMentionTextLen]
		}
		result = strings.ReplaceAll(result, "@"+mention+" ", "")
		result = strings.ReplaceAll(result, "@"+mention, "")
	}
	return strings.TrimSpace(result)
}

func buildPodPrompt(content, channelName string, channelID int64, podKeys []string) string {
	rawPrompt := stripPodMentions(content, podKeys)
	rawPrompt = ptyPromptFlattener.Replace(rawPrompt)
	return fmt.Sprintf("Message from channel(#%s, channel_id=%d): %s. If you finish it, please reply to this channel using send_channel_message(channel_id=%d).", channelName, channelID, rawPrompt, channelID)
}

var ptyPromptFlattener = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")
