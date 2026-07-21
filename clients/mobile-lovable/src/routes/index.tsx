import { Link, createFileRoute } from "@tanstack/react-router";
import { FolderPlus, GitBranch, Loader2, Search, Server, Zap } from "lucide-react";
import { useMemo, useState } from "react";
import { AgentCloudLogo } from "@/components/agent-cloud-logo";
import { NotificationCenter } from "@/components/notification-center";
import { MobileFrame } from "@/components/mobile-frame";
import { StatusPill } from "@/components/status-pill";
import { useLiveProjects } from "@/hooks/useLiveProjects";
import { useSessionsList } from "@/hooks/useSessionsList";
import { liveSummaryToAgentSession, statusRank } from "@/lib/live-session-mapper";
import { localProjectMeta } from "@/lib/projects-local";
import { projectIdFromName } from "@/lib/project-label";
import { useIsAuthed } from "@/hooks/useIsAuthed";
import { APP_NAME, pageTitle } from "@/lib/app-brand";
import type { AgentSession } from "@/lib/session-types";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: pageTitle("移动端") },
      {
        name: "description",
        content: "在手机上通过 ACP 协议连接 Codex / Claude Code / Gemini CLI，多项目多会话统一管理。",
      },
    ],
  }),
  component: Home,
});

function Home() {
  const liveList = useSessionsList();
  const liveProjects = useLiveProjects();
  const authed = useIsAuthed();
  const [q, setQ] = useState("");

  const allSessions = useMemo(
    () => (authed ? liveList.items.map(liveSummaryToAgentSession) : []),
    [authed, liveList.items],
  );

  const needle = q.trim().toLowerCase();
  const filteredSessions = needle
    ? allSessions.filter(
        (s) =>
          s.title.toLowerCase().includes(needle) ||
          s.agent.toLowerCase().includes(needle) ||
          s.id.toLowerCase().includes(needle),
      )
    : allSessions;

  const pendingApprovals = authed ? liveList.pendingCount : 0;
  const running = allSessions.filter((s) => s.status === "running").length;
  const recentTasks = [...filteredSessions].sort((a, b) => statusRank(a.status) - statusRank(b.status)).slice(0, 8);
  const firstPending = allSessions.find((s) => s.status === "waiting_approval");

  const projectNames = authed ? liveProjects.names : [];

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 border-b border-border/60 bg-background/85 px-5 pb-3 pt-4 backdrop-blur-xl">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className="relative flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl bg-primary/10 p-1.5 ring-1 ring-primary/30">
                <AgentCloudLogo className="h-full w-full" />
                {authed && <span className="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-success ring-2 ring-background" />}
              </div>
              <div>
                <h1 className="text-[15px] font-semibold leading-tight tracking-tight">{APP_NAME}</h1>
                <p className="text-[11px] text-muted-foreground">
                  {authed ? `${running} 运行 · ${pendingApprovals} 待审批` : "未连接"}
                </p>
              </div>
            </div>
            <NotificationCenter />
          </div>

          {!authed && (
            <Link
              to="/login"
              className="mt-3 block rounded-xl border border-primary/30 bg-primary/10 px-3 py-2 text-center text-[12px] font-medium text-primary"
            >
              登录 {APP_NAME} 以连接真实 Agent
            </Link>
          )}

          <div className="mt-3 flex items-center gap-2 rounded-xl bg-surface px-3 py-2 ring-1 ring-border/40">
            <Search className="h-3.5 w-3.5 text-muted-foreground" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="搜索项目 / 会话 / Agent..."
              className="flex-1 bg-transparent text-[13px] outline-none placeholder:text-muted-foreground"
            />
          </div>
        </header>

        {firstPending && (
          <Link
            to="/sessions/$sessionId"
            params={{ sessionId: firstPending.id }}
            className="mx-5 mt-4 block rounded-2xl border border-warning/40 bg-warning/10 p-3 stream-in"
          >
            <p className="text-[11px] font-semibold text-warning">⚠ {pendingApprovals} 个操作待你审批</p>
            <p className="mt-0.5 truncate text-[12px] text-foreground/80">{firstPending.title}</p>
          </Link>
        )}

        {authed && (
          <section className="px-5 pt-5">
            <div className="mb-2.5 flex items-center justify-between">
              <h2 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                项目 · {projectNames.length}
              </h2>
              <Link
                to="/projects/new"
                className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] text-muted-foreground hover:text-foreground"
              >
                <FolderPlus className="h-3 w-3" />
                新建
              </Link>
            </div>
            {liveProjects.loading ? (
              <div className="flex items-center gap-2 py-6 text-[12px] text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" /> 加载项目…
              </div>
            ) : projectNames.length === 0 ? (
              <p className="rounded-2xl border border-dashed border-border/60 py-8 text-center text-[12px] text-muted-foreground">
                暂无项目 · 创建任务时可指定项目
              </p>
            ) : (
              <div className="space-y-2">
                {projectNames.map((name) => (
                  <LiveProjectCard key={name} name={name} sessions={allSessions} />
                ))}
              </div>
            )}
          </section>
        )}

        <section className="flex-1 px-5 pb-24 pt-5">
          <div className="mb-2.5 flex items-center justify-between">
            <h2 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
              {authed ? "最近任务" : "会话"}
            </h2>
            <span className="text-[11px] text-muted-foreground">{filteredSessions.length} 条</span>
          </div>
          {!authed ? (
            <p className="rounded-2xl border border-dashed border-border/60 py-10 text-center text-[12px] text-muted-foreground">
              登录后查看真实 Agent 会话
            </p>
          ) : liveList.loading ? (
            <div className="flex items-center gap-2 py-8 text-[12px] text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> 同步会话…
            </div>
          ) : (
            <div className="space-y-2">
              {recentTasks.map((s) => (
                <TaskRow key={s.id} session={s} />
              ))}
              {recentTasks.length === 0 && (
                <Link
                  to="/new"
                  className="block rounded-2xl border border-dashed border-primary/40 bg-primary/5 py-10 text-center text-[12px] text-primary"
                >
                  创建第一个任务 →
                </Link>
              )}
            </div>
          )}
        </section>
      </div>
    </MobileFrame>
  );
}

function LiveProjectCard({ name, sessions }: { name: string; sessions: AgentSession[] }) {
  const meta = localProjectMeta(name);
  const pid = projectIdFromName(name);
  const ps = sessions.filter((s) => s.projectId === pid);
  const running = ps.filter((s) => s.status === "running").length;
  const approvals = ps.filter((s) => s.status === "waiting_approval").length;
  const preview = ps.slice(0, 2);
  const color = meta?.color ?? "primary";
  const accent = {
    primary: { bg: "bg-primary/15", text: "text-primary" },
    accent: { bg: "bg-accent/15", text: "text-accent" },
    info: { bg: "bg-info/15", text: "text-info" },
  }[color] ?? { bg: "bg-muted", text: "text-muted-foreground" };

  return (
    <Link
      to="/projects/$projectId"
      params={{ projectId: pid }}
      className="block rounded-2xl border border-border/50 bg-card p-4 transition hover:border-primary/40 active:scale-[0.99]"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className={cn("flex h-10 w-10 items-center justify-center rounded-xl ring-1 ring-white/5", accent.bg)}>
            <GitBranch className={cn("h-4 w-4", accent.text)} />
          </div>
          <div>
            <h3 className="text-[14.5px] font-semibold leading-tight">{name}</h3>
            {meta?.repo && (
              <p className="mt-0.5 flex items-center gap-1 font-mono text-[11px] text-muted-foreground">{meta.repo}</p>
            )}
          </div>
        </div>
      </div>
      <div className="mt-3 flex items-center gap-2 text-[11px]">
        {meta?.host && (
          <span className="flex items-center gap-1 rounded-full bg-success/10 px-2 py-0.5 text-success">
            <Server className="h-3 w-3" />
            {meta.host}
          </span>
        )}
        {running > 0 && <span className="rounded-full bg-primary/15 px-2 py-0.5 text-primary">{running} 运行</span>}
        {approvals > 0 && <span className="rounded-full bg-warning/15 px-2 py-0.5 text-warning">{approvals} 待批</span>}
        <span className="ml-auto text-muted-foreground">{ps.length} 会话</span>
      </div>
      {preview.length > 0 && (
        <div className="mt-3 space-y-1.5 border-t border-border/50 pt-3">
          {preview.map((s) => (
            <div key={s.id} className="flex items-center gap-2">
              <StatusPill status={s.status} className="shrink-0" />
              <span className="truncate text-[11.5px] text-foreground/80">{s.title}</span>
            </div>
          ))}
        </div>
      )}
    </Link>
  );
}

function TaskRow({ session }: { session: AgentSession }) {
  return (
    <Link
      to="/sessions/$sessionId"
      params={{ sessionId: session.id }}
      className="flex items-center gap-3 rounded-2xl border border-border/50 bg-card p-3 transition hover:border-primary/40 active:scale-[0.99]"
    >
      <StatusPill status={session.status} className="shrink-0" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-[13px] font-medium leading-tight">{session.title}</p>
        <p className="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className="truncate">{session.agent}</span>
          <span>·</span>
          <span className="shrink-0">{session.updatedAt}</span>
        </p>
      </div>
      {(session.status === "running" || session.status === "waiting_approval") && (
        <Zap className="h-3.5 w-3.5 shrink-0 text-primary" />
      )}
    </Link>
  );
}
