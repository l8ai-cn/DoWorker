import { useSyncExternalStore } from "react";
import { toast } from "sonner";
import type { LiveSessionSummary } from "./sessions-api";

export type NotificationKind =
  | "approval"
  | "ask_user"
  | "error"
  | "success"
  | "info"
  | "goal";

export interface AppNotification {
  id: string;
  kind: NotificationKind;
  title: string;
  body?: string;
  ts: number; // epoch ms
  read: boolean;
  sessionId?: string;
  href?: string;
}

const kindMeta: Record<NotificationKind, { emoji: string; ring: string }> = {
  approval: { emoji: "🔐", ring: "ring-warning/40" },
  ask_user: { emoji: "❓", ring: "ring-info/40" },
  error:    { emoji: "⛔", ring: "ring-destructive/40" },
  success:  { emoji: "✅", ring: "ring-success/40" },
  info:     { emoji: "💬", ring: "ring-border" },
  goal:     { emoji: "🎯", ring: "ring-accent/40" },
};

export const notificationMeta = kindMeta;

// ---------- store ----------
let items: AppNotification[] = [];
const listeners = new Set<() => void>();
let liveSynced = false;

function emit() {
  listeners.forEach((l) => l());
}

export function syncNotificationsFromSessions(sessions: LiveSessionSummary[]): void {
  if (sessions.length === 0 && liveSynced) return;
  liveSynced = true;
  const now = Date.now();
  const next: AppNotification[] = [];
  sessions.forEach((s, i) => {
    const title = s.title ?? s.agentName ?? s.id;
    if (s.pendingApprovals > 0) {
      next.push({
        id: `n-appr-${s.id}`,
        kind: "approval",
        title: `待审批：${title}`,
        body: s.workspace ?? "Agent 请求执行敏感操作",
        ts: now - (i + 1) * 60_000,
        read: false,
        sessionId: s.id,
        href: `/sessions/${s.id}`,
      });
    }
    if (s.status === "failed") {
      next.push({
        id: `n-err-${s.id}`,
        kind: "error",
        title: `任务失败：${title}`,
        ts: now - (i + 2) * 60_000,
        read: false,
        sessionId: s.id,
        href: `/sessions/${s.id}`,
      });
    }
  });
  if (next.length === 0 && items.length === 0) {
    next.push({
      id: "n-welcome",
      kind: "info",
      title: "已连接 Do Worker",
      body: "会话审批与状态变更会实时汇总到这里。",
      ts: now,
      read: true,
    });
  }
  items = next.sort((a, b) => b.ts - a.t);
  emit();
}

export function resetNotificationsForLogout(): void {
  items = [];
  liveSynced = false;
  emit();
}

export function subscribeNotifications(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function getNotifications(): AppNotification[] {
  return items;
}

export function getUnreadCount(): number {
  return items.reduce((n, x) => n + (x.read ? 0 : 1), 0);
}

export function markRead(id: string) {
  let changed = false;
  items = items.map((n) => {
    if (n.id === id && !n.read) {
      changed = true;
      return { ...n, read: true };
    }
    return n;
  });
  if (changed) emit();
}

export function markAllRead() {
  if (items.every((n) => n.read)) return;
  items = items.map((n) => ({ ...n, read: true }));
  emit();
}

export function clearAll() {
  if (items.length === 0) return;
  items = [];
  emit();
}

export function pushNotification(
  n: Omit<AppNotification, "id" | "ts" | "read"> & { id?: string; ts?: number; read?: boolean },
  opts: { toast?: boolean } = { toast: true },
) {
  const item: AppNotification = {
    id: n.id ?? `n-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    ts: n.ts ?? Date.now(),
    read: n.read ?? false,
    kind: n.kind,
    title: n.title,
    body: n.body,
    sessionId: n.sessionId,
    href: n.href,
  };
  items = [item, ...items].slice(0, 200);
  emit();
  if (opts.toast !== false) fireToast(item);
  return item;
}

function fireToast(n: AppNotification) {
  const meta = kindMeta[n.kind];
  const opts = { description: n.body, duration: n.kind === "error" ? 6000 : 4000 };
  const label = `${meta.emoji} ${n.title}`;
  switch (n.kind) {
    case "error":   toast.error(label, opts); break;
    case "success": toast.success(label, opts); break;
    case "approval":
    case "ask_user":
    case "goal":    toast.warning(label, opts); break;
    default:        toast(label, opts);
  }
}

// ---------- react hooks ----------
export function useNotifications(): AppNotification[] {
  return useSyncExternalStore(subscribeNotifications, getNotifications, getNotifications);
}

export function useUnreadCount(): number {
  return useSyncExternalStore(subscribeNotifications, getUnreadCount, getUnreadCount);
}

// ---------- pretty time ----------
export function formatRelative(ts: number, now = Date.now()): string {
  const s = Math.max(1, Math.round((now - ts) / 1000));
  if (s < 60) return `${s} 秒前`;
  const m = Math.round(s / 60);
  if (m < 60) return `${m} 分钟前`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h} 小时前`;
  const d = Math.round(h / 24);
  return `${d} 天前`;
}
