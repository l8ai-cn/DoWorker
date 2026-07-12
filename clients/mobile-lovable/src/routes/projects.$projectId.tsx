import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, GitBranch, Loader2, Server } from "lucide-react";
import { useMemo } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { StatusPill } from "@/components/status-pill";
import { useSessionsList } from "@/hooks/useSessionsList";
import { liveSummaryToAgentSession } from "@/lib/live-session-mapper";
import { localProjectMeta } from "@/lib/projects-local";
import { projectNameFromId } from "@/lib/project-label";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import type { AgentSession } from "@/lib/session-types";

export const Route = createFileRoute("/projects/$projectId")({
  head: ({ params }) => ({
    meta: [{ title: pageTitle(projectNameFromId(params.projectId)) }],
  }),
  component: ProjectDetail,
});

function ProjectDetail() {
  const { projectId } = Route.useParams();
  const name = projectNameFromId(projectId);
  const meta = localProjectMeta(name);
  const liveList = useSessionsList();
  const authed = Boolean(readAuthToken());

  const sessions = useMemo(() => {
    if (!authed) return [];
    return liveList.items
      .filter((s) => s.project === name)
      .map(liveSummaryToAgentSession);
  }, [authed, liveList.items, name]);

  const grouped = {
    active: sessions.filter((s) => s.status === "running" || s.status === "waiting_approval"),
    idle: sessions.filter((s) => s.status === "idle" || s.status === "completed"),
    failed: sessions.filter((s) => s.status === "failed"),
  };

  if (!authed) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后查看项目会话</p>
          <Link to="/login" className="text-sm text-primary">去登录</Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
          <div className="flex items-center gap-2">
            <Link to="/" className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface">
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <div className="min-w-0 flex-1">
              <p className="text-[10.5px] text-muted-foreground">项目</p>
              <h1 className="line-clamp-1 text-[14px] font-semibold leading-tight">{name}</h1>
            </div>
            <Link
              to="/new"
              search={{ project: projectId }}
              className="rounded-full bg-primary px-3 py-1 text-[11px] font-semibold text-primary-foreground"
            >
              新任务
            </Link>
          </div>
          {(meta?.repo || meta?.host) && (
            <div className="mt-3 grid grid-cols-2 gap-2 text-[11px]">
              {meta.repo && (
                <div className="flex items-center gap-1.5 rounded-lg bg-surface px-2.5 py-1.5 ring-1 ring-border/40">
                  <GitBranch className="h-3 w-3 text-muted-foreground" />
                  <span className="truncate font-mono text-[11px]">{meta.repo}</span>
                </div>
              )}
              {meta.host && (
                <div className="flex items-center gap-1.5 rounded-lg bg-surface px-2.5 py-1.5 ring-1 ring-border/40">
                  <Server className="h-3 w-3 text-success" />
                  <span className="truncate font-mono text-[11px]">{meta.host}</span>
                </div>
              )}
            </div>
          )}
        </header>

        <div className="flex-1 space-y-6 px-5 pb-32 pt-5">
          {liveList.loading ? (
            <div className="flex items-center gap-2 py-8 text-[12px] text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> 加载会话…
            </div>
          ) : (
            <>
              {grouped.active.length > 0 && <SessionGroup title="进行中" items={grouped.active} />}
              {grouped.idle.length > 0 && <SessionGroup title="空闲 / 已完成" items={grouped.idle} />}
              {grouped.failed.length > 0 && <SessionGroup title="失败" items={grouped.failed} />}
              {sessions.length === 0 && (
                <div className="rounded-2xl border border-dashed border-border/70 p-8 text-center">
                  <p className="text-[13px] text-muted-foreground">这个项目下还没有会话</p>
                  <Link to="/new" search={{ project: projectId }} className="mt-2 inline-block text-[12px] text-primary">
                    创建第一个任务 →
                  </Link>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </MobileFrame>
  );
}

function SessionGroup({ title, items }: { title: string; items: AgentSession[] }) {
  return (
    <section>
      <h2 className="mb-2 text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
        {title} · {items.length}
      </h2>
      <div className="space-y-2">
        {items.map((s) => (
          <Link
            key={s.id}
            to="/sessions/$sessionId"
            params={{ sessionId: s.id }}
            className="block rounded-xl border border-border/50 bg-card p-3.5 transition hover:border-primary/40 active:scale-[0.99]"
          >
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <StatusPill status={s.status} />
                  <span className="font-mono text-[10px] text-muted-foreground">{s.agent}</span>
                </div>
                <h3 className="mt-1.5 line-clamp-2 text-[13.5px] font-medium leading-snug">{s.title}</h3>
              </div>
              <span className="shrink-0 text-[10.5px] text-muted-foreground">{s.updatedAt}</span>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
