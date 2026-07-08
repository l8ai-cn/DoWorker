import { useState } from "react";
import { Link } from "@tanstack/react-router";
import { Bell, CheckCheck, Trash2 } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  useNotifications,
  useUnreadCount,
  markRead,
  markAllRead,
  clearAll,
  notificationMeta,
  formatRelative,
  type AppNotification,
} from "@/lib/notifications";

export function NotificationCenter({ triggerClassName = "" }: { triggerClassName?: string }) {
  const items = useNotifications();
  const unread = useUnreadCount();
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={`通知 · ${unread} 未读`}
          className={
            "relative flex h-9 w-9 items-center justify-center rounded-full bg-surface hover:bg-surface-2 " +
            triggerClassName
          }
        >
          <Bell className="h-4 w-4 text-muted-foreground" />
          {unread > 0 && (
            <span className="absolute right-1.5 top-1.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-warning px-1 text-[9px] font-bold text-primary-foreground">
              {unread > 99 ? "99+" : unread}
            </span>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        sideOffset={8}
        className="w-[340px] p-0 overflow-hidden"
      >
        <div className="flex items-center justify-between border-b border-border/60 px-3 py-2">
          <div className="flex items-center gap-2">
            <span className="text-[13px] font-semibold">通知</span>
            {unread > 0 && (
              <span className="rounded-full bg-warning/15 px-1.5 py-0.5 text-[10px] font-semibold text-warning">
                {unread} 未读
              </span>
            )}
          </div>
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={markAllRead}
              disabled={unread === 0}
              className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-[11px] text-muted-foreground hover:bg-surface disabled:opacity-40"
              title="全部标为已读"
            >
              <CheckCheck className="h-3 w-3" /> 已读
            </button>
            <button
              type="button"
              onClick={clearAll}
              disabled={items.length === 0}
              className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-[11px] text-muted-foreground hover:bg-surface disabled:opacity-40"
              title="清空"
            >
              <Trash2 className="h-3 w-3" /> 清空
            </button>
          </div>
        </div>
        <div className="max-h-[420px] overflow-y-auto">
          {items.length === 0 ? (
            <div className="px-4 py-10 text-center text-[12px] text-muted-foreground">
              目前没有通知
            </div>
          ) : (
            <ul className="divide-y divide-border/40">
              {items.map((n) => (
                <NotificationItem
                  key={n.id}
                  n={n}
                  onNavigate={() => {
                    markRead(n.id);
                    setOpen(false);
                  }}
                />
              ))}
            </ul>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}

function NotificationItem({ n, onNavigate }: { n: AppNotification; onNavigate: () => void }) {
  const meta = notificationMeta[n.kind];
  const content = (
    <div className="flex gap-2.5 px-3 py-2.5">
      <div
        className={
          "mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-surface ring-1 " +
          meta.ring
        }
      >
        <span className="text-[13px] leading-none">{meta.emoji}</span>
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline justify-between gap-2">
          <p className={"truncate text-[12.5px] " + (n.read ? "text-foreground/80" : "font-semibold text-foreground")}>
            {n.title}
          </p>
          <span className="shrink-0 text-[10px] text-muted-foreground">{formatRelative(n.ts)}</span>
        </div>
        {n.body && (
          <p className="mt-0.5 line-clamp-2 text-[11.5px] text-muted-foreground">{n.body}</p>
        )}
      </div>
      {!n.read && <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" />}
    </div>
  );

  if (n.href) {
    return (
      <li>
        <Link
          to={n.href as string}
          onClick={onNavigate}
          className="block hover:bg-surface"
        >
          {content}
        </Link>
      </li>
    );
  }

  return (
    <li>
      <button
        type="button"
        onClick={() => markRead(n.id)}
        className="block w-full text-left hover:bg-surface"
      >
        {content}
      </button>
    </li>
  );
}
