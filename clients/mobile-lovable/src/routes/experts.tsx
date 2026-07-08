import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, Loader2, Search, Server } from "lucide-react";
import { useMemo, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { useLiveExperts } from "@/hooks/useLiveExperts";
import { APP_NAME, pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";

export const Route = createFileRoute("/experts")({
  head: () => ({
    meta: [
      { title: pageTitle("专家库") },
      { name: "description", content: "组织内可复用的 Expert Agent，选中即可下发任务。" },
    ],
  }),
  component: ExpertsPage,
});

function ExpertsPage() {
  const [q, setQ] = useState("");
  const liveExperts = useLiveExperts();
  const authed = Boolean(readAuthToken());

  const filtered = useMemo(() => {
    const items = liveExperts.items;
    if (!q.trim()) return items;
    const needle = q.toLowerCase();
    return items.filter(
      (e) =>
        e.name.toLowerCase().includes(needle) ||
        (e.description ?? "").toLowerCase().includes(needle) ||
        e.agent_slug.toLowerCase().includes(needle),
    );
  }, [liveExperts.items, q]);

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 border-b border-border/60 bg-background/85 px-5 pb-3 pt-4 backdrop-blur-xl">
          <div className="flex items-center gap-2">
            <Link to="/" className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface">
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <h1 className="flex-1 text-[15px] font-semibold tracking-tight">专家库</h1>
            <span className="text-[11px] text-muted-foreground">
              {authed ? `${liveExperts.items.length} 位专家` : "—"}
            </span>
          </div>
          <div className="mt-3 flex items-center gap-2 rounded-xl bg-surface px-3 py-2.5 ring-1 ring-border/40">
            <Search className="h-3.5 w-3.5 text-muted-foreground" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="搜索专家名称、描述..."
              className="flex-1 bg-transparent text-[13px] outline-none placeholder:text-muted-foreground"
            />
          </div>
        </header>

        <div className="flex-1 space-y-4 px-5 pb-24 pt-5">
          {!authed && (
            <div className="rounded-2xl border border-border/60 bg-surface/40 px-4 py-6 text-center text-[12px] text-muted-foreground">
              <Link to="/login" className="font-medium text-primary">登录 {APP_NAME}</Link> 后查看组织专家
            </div>
          )}

          {authed && liveExperts.loading && (
            <div className="flex items-center gap-2 py-8 text-[12px] text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> 加载专家…
            </div>
          )}

          {authed && !liveExperts.loading && (
            <section>
              <div className="mb-3 flex items-center gap-1.5">
                <Server className="h-3.5 w-3.5 text-primary" />
                <h2 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                  组织专家 · {filtered.length}
                </h2>
              </div>
              <div className="space-y-2">
                {filtered.map((e) => (
                  <Link
                    key={e.slug}
                    to="/experts/$expertId"
                    params={{ expertId: e.slug }}
                    className="flex items-center gap-3 rounded-2xl border border-border/50 bg-card p-3 transition hover:border-primary/40 active:scale-[0.99]"
                  >
                    <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary/15 text-lg">
                      🤖
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[13.5px] font-semibold">{e.name}</p>
                      <p className="truncate text-[11px] text-muted-foreground">
                        {e.agent_slug} · {e.interaction_mode.toUpperCase()}
                      </p>
                      {e.description && (
                        <p className="mt-0.5 line-clamp-1 text-[11px] text-foreground/70">{e.description}</p>
                      )}
                    </div>
                    <span className="text-[10px] text-muted-foreground">{e.run_count} 次</span>
                  </Link>
                ))}
                {filtered.length === 0 && (
                  <p className="rounded-2xl border border-dashed border-border/60 py-10 text-center text-[12px] text-muted-foreground">
                    暂无组织专家 · 可在 Admin 或 API 中创建
                  </p>
                )}
              </div>
            </section>
          )}
        </div>
      </div>
    </MobileFrame>
  );
}
