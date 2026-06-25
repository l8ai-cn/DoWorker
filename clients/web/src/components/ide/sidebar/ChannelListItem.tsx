"use client";

import { cn } from "@/lib/utils";
import type { Channel, ChannelLastMessage } from "@/stores/channel";
import { formatRelativeShort } from "@/lib/format-relative-time";
import { Lock } from "lucide-react";

interface ChannelListItemProps {
  channel: Channel;
  isSelected: boolean;
  unreadCount?: number;
  lastMessage?: ChannelLastMessage | null;
  onClick: () => void;
}

/**
 * Channel row in the sidebar list: hash + name + last message preview, short
 * time, and unread dot. Private channels use the lock icon in place of #.
 */
export function ChannelListItem({
  channel,
  isSelected,
  unreadCount = 0,
  lastMessage,
  onClick,
}: ChannelListItemProps) {
  const hasUnread = unreadCount > 0 && !isSelected;
  const isPrivate = channel.visibility === "private";
  const preview = lastMessage
    ? lastMessage.sender_name
      ? `${lastMessage.sender_name}: ${lastMessage.content_preview}`
      : lastMessage.content_preview
    : channel.description ?? "";
  const time = formatRelativeShort(lastMessage?.timestamp ?? channel.updated_at);

  return (
    <button
      type="button"
      onClick={onClick}
      data-testid="channel-list-item"
      data-channel-id={String(channel.id)}
      className={cn(
        "nav-row pressable w-full gap-2.5 !py-2",
        isSelected ? "nav-row-active" : "nav-row-idle",
      )}
    >
      <span
        className={cn(
          "shrink-0 font-mono text-[14px]",
          isSelected ? "font-semibold text-foreground" : "text-muted-foreground/70",
        )}
      >
        {isPrivate ? <Lock className="h-3.5 w-3.5" /> : "#"}
      </span>

      <span className="min-w-0 flex-1 flex flex-col gap-0.5">
        <span className="flex items-center justify-between gap-2">
          <span
            className={cn(
              "truncate text-[13px]",
              isSelected ? "font-semibold text-foreground" : "text-foreground",
            )}
          >
            {channel.name}
          </span>
          {time && (
            <span className="shrink-0 text-[10px] text-muted-foreground/70">{time}</span>
          )}
        </span>
        {preview && (
          <span className="truncate text-[11px] text-muted-foreground/70">
            {preview}
          </span>
        )}
      </span>

      <span className="flex w-2 shrink-0 justify-center">
        {hasUnread && (
          <span className="h-1.5 w-1.5 rounded-full bg-destructive" aria-label="unread" />
        )}
      </span>
    </button>
  );
}

export default ChannelListItem;
