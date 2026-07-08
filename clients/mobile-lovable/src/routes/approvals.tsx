import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, ShieldAlert } from "lucide-react";
import { MobileFrame } from "@/components/mobile-frame";
import { StatusPill } from "@/components/status-pill";
import { useSessionsList } from "@/hooks/useSessionsList";
import { liveSummaryToAgentSession } from "@/lib/live-session-mapper";
import { pageTitle, APP_NAME } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";

export const Route = createFileRoute("/approvals")({
  head: () => ({
    meta: [
      { title: pageTitle("待审批") },
      { name: "description", content: "所有等待你审批的 Agent 操作聚合。" },
    ],
  }),
  component: ApprovalsPage,
});

function ApprovalsPage() {
  const liveList = useSessionsList();
  const authed = Boolean(readAuthToken());
  const pending = authed
    ? liveList.items.filter((s) => s.pendingApprovals > 0).map(liveSummaryToAgentSession)
    : [];

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 flex items-center gap-2 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
          <Link to="/" className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div className="flex-1">
            <h1 className="text-[14px] font-semibold">待审批</h1>
            <p className="text-[10.5px] text-muted-foreground">{pending.length} 个操作等待处理</p>
          </div>
        </header>

        {!authed && (
          <div className="mx-5 mt-3 rounded-xl border border-border/60 bg-surface/40 px-3 py-2 text-center text-[11px] text-muted-foreground">
            <Link to="/login" className="text-primary">登录</Link> 后查看真实待审批项
          </div>
        )}

        <div className="flex-1 space-y-2 px-5 pb-24 pt-4">
          {pending.length === 0 && (
            <div className="flex flex-col items-center gap-2 rounded-2xl border border-dashed border-border/60 py-12 text-center">
              <ShieldAlert className="h-6 w-6 text-muted-foreground/50" />
              <p className="text-[12.5px] text-muted-foreground">当前没有待审批项</p>
            </div>
          )}
          {pending.map((s) => {
            const row = liveList.items.find((r) => r.id === s.id);
            return (
              <Link
                key={s.id}
                to="/sessions/$sessionId"
                params={{ sessionId: s.id }}
                className="block rounded-2xl border border-warning/40 bg-gradient-to-br from-warning/15 to-warning/5 p-4 transition active:scale-[0.99]"
              >
                <div className="flex items-center gap-2">
                  <StatusPill status={s.status} />
                  <span className="text-[11px] text-warning">{s.updatedAt}</span>
                  {row?.pendingApprovals ? (
                    <span className="ml-auto rounded-full bg-warning/20 px-2 py-0.5 text-[10px] font-medium text-warning">
                      {row.pendingApprovals} 项
                    </span>
                  ) : null}
                </div>
                <p className="mt-2 text-[13.5px] font-semibold leading-tight">{s.title}</p>
                <p className="mt-1 text-[11px] text-muted-foreground">
                  {s.projectId !== "live" ? s.projectId : APP_NAME} · {s.agent}
                </p>
                <p className="mt-2 line-clamp-2 rounded-lg bg-background/40 px-2 py-1.5 text-[11.5px] text-foreground/85">
                  {s.preview || "Agent 请求执行敏感操作，等待你的确认"}
                </p>
              </Link>
            );
          })}
        </div>
      </div>
    </MobileFrame>
  );
}
