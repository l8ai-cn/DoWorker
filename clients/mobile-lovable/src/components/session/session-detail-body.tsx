import { Link } from "@tanstack/react-router";
import {
  ArrowLeft,
  Clock,
  Coins,
  FileEdit,
  Hammer,
  Loader2,
  MoreHorizontal,
} from "lucide-react";
import { useEffect, useRef } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { StatusPill } from "@/components/status-pill";
import { SessionViewToggle } from "@/components/session-view-toggle";
import { SessionBottomBar } from "@/components/session/session-bottom-bar";
import { EventCard } from "@/components/session/session-event-card";
import { formatTokens, Metric } from "@/components/session/session-metrics-chips";
import { useSessionVisibleEvents } from "@/hooks/useSessionVisibleEvents";
import type { AgentEvent, AgentSession, Project } from "@/lib/mock-agents";
import { pushNotification } from "@/lib/notifications";
import { useDecisions } from "@/lib/session-approval-decisions";

export function SessionDetailBody({
  session,
  project,
  isLive = false,
}: {
  session: AgentSession;
  project: Project | undefined;
  isLive?: boolean;
}) {
  const { visibleEvents, visibleCount, isStreaming, currentPhase, phaseTotal } =
    useSessionVisibleEvents(session, isLive);

  const decisions = useDecisions();
  const pendingApproval = visibleEvents.find(
    (e: AgentEvent) =>
      e.type === "permission_request" && e.status === "pending" && !decisions[e.id],
  );

  const bottomRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (isStreaming) bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [visibleCount, isStreaming]);

  const notifiedRef = useRef<Set<string>>(new Set());
  useEffect(() => {
    if (!isStreaming) return;
    const last = visibleEvents[visibleEvents.length - 1];
    if (!last || notifiedRef.current.has(last.id)) return;
    notifiedRef.current.add(last.id);
    const href = `/sessions/${session.id}`;
    if (last.type === "permission_request" && last.status === "pending") {
      pushNotification({
        kind: "approval",
        title: "待审批：" + (last.title ?? "工具调用"),
        body: session.title,
        sessionId: session.id,
        href,
      });
    } else if (last.type === "ask_user") {
      pushNotification({
        kind: "ask_user",
        title: "需要你的输入：" + (last.askUser?.title ?? last.title ?? ""),
        body: session.title,
        sessionId: session.id,
        href,
      });
    } else if (last.type === "error") {
      pushNotification({
        kind: "error",
        title: "执行出错：" + (last.title ?? ""),
        body: last.detail,
        sessionId: session.id,
        href,
      });
    } else if (last.type === "phase") {
      pushNotification(
        {
          kind: "info",
          title: `阶段 ${last.phaseIndex}/${last.phaseTotal}：${last.title ?? ""}`,
          body: last.phaseSummary,
          sessionId: session.id,
          href,
        },
        { toast: false },
      );
    }
    if (visibleCount === session.events.length) {
      pushNotification({
        kind: "success",
        title: "任务完成：" + session.title,
        sessionId: session.id,
        href,
      });
    }
  }, [visibleCount, isStreaming, visibleEvents, session]);

  return (
    <MobileFrame hideNav>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 border-b border-border/60 bg-background/90 px-3 pb-2.5 pt-2.5 backdrop-blur-xl">
          <div className="flex items-center gap-2">
            <Link
              to={project ? "/projects/$projectId" : "/"}
              params={project ? { projectId: project.id } : undefined}
              className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full hover:bg-surface"
            >
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <div className="min-w-0 flex-1">
              <h1 className="line-clamp-1 text-[15px] font-semibold leading-snug">{session.title}</h1>
              <div className="mt-0.5 flex items-center gap-1.5 overflow-hidden">
                <StatusPill status={session.status} className="shrink-0 py-px text-[10px]" />
                <span className="truncate text-[11px] text-muted-foreground">{session.agent}</span>
                {session.branch && (
                  <>
                    <span className="text-muted-foreground/40">·</span>
                    <span className="truncate font-mono text-[10px] text-muted-foreground">@{session.branch}</span>
                  </>
                )}
              </div>
            </div>
            {isLive && <SessionViewToggle sessionId={session.id} mode="chat" />}
            <button className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full hover:bg-surface">
              <MoreHorizontal className="h-4 w-4 text-muted-foreground" />
            </button>
          </div>
        </header>

        {session.metrics && (
          <div className="grid grid-cols-4 gap-px border-b border-border/60 bg-border/40 text-center">
            <Metric icon={Clock} label="用时" value={session.metrics.elapsed} />
            <Metric icon={Hammer} label="工具" value={String(session.metrics.toolCalls)} />
            <Metric icon={FileEdit} label="文件" value={String(session.metrics.filesChanged)} />
            <Metric icon={Coins} label="Token" value={formatTokens(session.metrics.tokensIn + session.metrics.tokensOut)} />
          </div>
        )}

        {currentPhase && phaseTotal ? (
          <div className="sticky top-[76px] z-20 border-b border-border/60 bg-background/90 px-4 py-2 backdrop-blur-xl">
            <div className="flex items-center gap-2">
              <span className="text-base leading-none">{currentPhase.phaseEmoji ?? "▶"}</span>
              <div className="min-w-0 flex-1">
                <p className="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-primary">
                  Phase {currentPhase.phaseIndex}/{phaseTotal}
                  <span className="text-muted-foreground/70">·</span>
                  <span className="truncate normal-case tracking-normal text-foreground/80">{currentPhase.title}</span>
                </p>
                <div className="mt-1 flex items-center gap-2">
                  <div className="h-1 flex-1 overflow-hidden rounded-full bg-surface">
                    <div
                      className="h-full rounded-full bg-primary transition-all"
                      style={{ width: `${Math.min(100, ((currentPhase.phaseIndex ?? 0) / phaseTotal) * 100)}%` }}
                    />
                  </div>
                  <span className="font-mono text-[9.5px] text-muted-foreground tabular-nums">
                    {visibleCount}/{session.events.length}
                  </span>
                </div>
              </div>
            </div>
          </div>
        ) : null}

        <div className="flex-1 space-y-3 px-3 pb-36 pt-4">
          {visibleEvents.map((ev: AgentEvent) => (
            <EventCard key={ev.id} event={ev} />
          ))}
          {isStreaming && (
            <div className="flex items-center gap-2 px-2 pt-1">
              <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />
              <span className="shimmer-text text-[12px]">
                {visibleCount < session.events.length ? "Agent 正在生成下一步..." : "Agent 正在思考..."}
              </span>
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        <SessionBottomBar pendingApproval={pendingApproval} />
      </div>
    </MobileFrame>
  );
}
