import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, Loader2, Zap } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { pageTitle } from "@/lib/app-brand";
import { getLiveExpert, type LiveExpert } from "@/lib/experts-api";
import { readAuthToken } from "@/lib/auth-store";

export const Route = createFileRoute("/experts/$expertId")({
  head: ({ params }) => ({
    meta: [{ title: pageTitle(params.expertId) }],
  }),
  component: ExpertDetail,
});

function ExpertDetail() {
  const { expertId } = Route.useParams();
  const authed = Boolean(readAuthToken());
  const [expert, setExpert] = useState<LiveExpert | null | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!authed) {
      setExpert(null);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const row = await getLiveExpert(expertId);
        if (!cancelled) setExpert(row);
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "加载失败");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [authed, expertId]);

  if (!authed) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 px-8 text-center">
          <p className="text-[13px] text-muted-foreground">请先登录查看专家详情</p>
          <Link to="/login" className="text-[13px] text-primary">去登录</Link>
        </div>
      </MobileFrame>
    );
  }

  if (expert === undefined) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen items-center justify-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" /> 加载专家…
        </div>
      </MobileFrame>
    );
  }

  if (!expert) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 px-8 text-center">
          <p className="text-[13px] text-muted-foreground">{error ?? "找不到这位专家"}</p>
          <Link to="/experts" className="text-[13px] text-primary">返回专家库</Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 flex items-center gap-2 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
          <Link to="/experts" className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <h1 className="flex-1 truncate text-[14px] font-semibold">{expert.name}</h1>
        </header>

        <div className="flex-1 space-y-5 px-5 pb-28 pt-5">
          <div className="flex items-start gap-3">
            <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl bg-primary/15 text-3xl">
              🤖
            </div>
            <div className="min-w-0 flex-1">
              <h2 className="truncate text-[17px] font-bold tracking-tight">{expert.name}</h2>
              <p className="mt-0.5 text-[11.5px] text-muted-foreground">
                {expert.agent_slug} · {expert.interaction_mode.toUpperCase()}
              </p>
              <p className="mt-1 text-[11px] text-muted-foreground">已运行 {expert.run_count} 次</p>
            </div>
          </div>

          {expert.description && (
            <p className="text-[13px] leading-relaxed text-foreground/85">{expert.description}</p>
          )}

          {expert.prompt && (
            <section>
              <h3 className="mb-2 text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
                系统提示
              </h3>
              <pre className="whitespace-pre-wrap rounded-xl bg-surface p-3 text-[11.5px] text-foreground/80 ring-1 ring-border/40">
                {expert.prompt}
              </pre>
            </section>
          )}
        </div>

        <div className="safe-bottom sticky bottom-0 border-t border-border/60 bg-background/95 px-5 pt-3 backdrop-blur-xl">
          <Link
            to="/new"
            search={{ expert: expert.slug, prompt: expert.prompt ?? undefined }}
            className="flex w-full items-center justify-center gap-2 rounded-full bg-primary py-3.5 text-[14px] font-semibold text-primary-foreground glow-primary transition active:scale-[0.98]"
          >
            <Zap className="h-4 w-4" />
            下发任务给此专家
          </Link>
        </div>
      </div>
    </MobileFrame>
  );
}
