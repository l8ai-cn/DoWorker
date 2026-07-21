package channel

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
)

type MemberIDProvider interface {
	GetMemberUserIDs(ctx context.Context, channelID int64) ([]int64, error)
}

func NewEventPublishHook(eb *eventbus.EventBus, userNames UserNameResolver, members MemberIDProvider) PostSendHook {
	return func(ctx context.Context, mc *MessageContext) error {
		if eb == nil {
			return nil
		}

		msgData := &eventsv1.ChannelMessageEventData{
			Id:           mc.Message.ID,
			ChannelId:    mc.Message.ChannelID,
			SenderPod:    mc.Message.SenderPod,
			SenderUserId: mc.Message.SenderUserID,
			SenderName:   resolveSenderName(ctx, mc, userNames),
			MessageType:  mc.Message.MessageType,
			Body:         mc.Message.Body,
			ContentJson:  marshalJSONField(mc.Message.Content),
			MentionsJson: marshalJSONField(mc.Message.Mentions),
			ReplyTo:      mc.Message.ReplyTo,
			CreatedAt:    mc.Message.CreatedAt.Format(time.RFC3339),
		}

		if mc.Message.SenderPodInfo != nil {
			info := &eventsv1.SenderPodInfoEventData{PodKey: mc.Message.SenderPodInfo.PodKey}
			if mc.Message.SenderPodInfo.Alias != nil {
				info.Alias = mc.Message.SenderPodInfo.Alias
			}
			if mc.Message.SenderPodInfo.Agent != nil {
				info.Agent = &eventsv1.SenderPodAgentEventData{Name: mc.Message.SenderPodInfo.Agent.Name}
			}
			msgData.SenderPodInfo = info
		}

		data, err := eventbus.MarshalEventData(msgData)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal channel message event", "error", err)
			return err
		}

		var targetUserIDs []int64
		if members != nil {
			targetUserIDs, _ = members.GetMemberUserIDs(ctx, mc.Channel.ID)
		}

		eb.Publish(ctx, &eventbus.Event{
			Type:           eventbus.EventChannelMessage,
			Category:       eventbus.CategoryEntity,
			OrganizationID: mc.Channel.OrganizationID,
			EntityType:     "channel",
			EntityID:       strconv.FormatInt(mc.Message.ChannelID, 10),
			Data:           data,
			TargetUserIDs:  targetUserIDs,
		})

		return nil
	}
}

// marshalJSONField serialises arbitrary JSONMap fields (Content/Mentions
// ASTs whose per-node-type schemas are renderer-owned) to the JSON string
// the proto carries opaquely. Empty in / empty out.
func marshalJSONField(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
