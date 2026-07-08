import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { SessionDetailBody } from "@/components/session/session-detail-body";
import { useLiveSession } from "@/hooks/useLiveSession";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { getProject, getSession, type AgentSession, type Project } from "@/lib/mock-agents";
import { SessionActionProvider } from "@/lib/session-action-context";

interface SessionLoaderData {
  session: AgentSession | null;
  project: Project | undefined;
}

export const Route = createFileRoute("/sessions/$sessionId")({
  loader: ({ params }): SessionLoaderData => {
    const session = getSession(params.sessionId);
    return { session: session ?? null, project: session ? getProject(session.projectId) : undefined };
  },
  head: ({ loaderData, params }) => ({
    meta: [
      { title: pageTitle(loaderData?.session?.title ?? params.sessionId) },
      {
        name: "description",
        content: loaderData?.session
          ? `实时查看 ${loaderData.session.agent ?? "agent"} 在 ${loaderData.project?.repo ?? ""} 的执行过程。`
          : "Agent 会话详情",
      },
    ],
  }),
  component: SessionDetail,
  errorComponent: ({ reset }) => {
    const router = useRouter();
    return (
      <MobileFrame>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm text-muted-foreground">加载会话失败</p>
          <button
            onClick={() => { router.invalidate(); reset(); }}
            className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-primary-foreground"
          >重试</button>
        </div>
      </MobileFrame>
    );
  },
  notFoundComponent: () => (
    <MobileFrame>
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
        <p className="text-sm">找不到这个会话</p>
        <Link to="/" className="text-xs text-primary">← 返回首页</Link>
      </div>
    </MobileFrame>
  ),
});

function SessionDetail() {
  const { sessionId } = Route.useParams();
  const live = useLiveSession(sessionId);
  const mockLoader = Route.useLoaderData();
  const [authed, setAuthed] = useState<boolean | null>(null);
  useEffect(() => {
    setAuthed(Boolean(readAuthToken()));
  }, []);

  if (authed === null) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
          <Loader2 className="mr-2 h-4 w-4 animate-spin" /> 加载…
        </div>
      </MobileFrame>
    );
  }
  const session = authed ? live.session : mockLoader.session ?? undefined;
  const project = authed ? undefined : mockLoader.project;

  if (authed && live.loading) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
          <Loader2 className="mr-2 h-4 w-4 animate-spin" /> 连接会话…
        </div>
      </MobileFrame>
    );
  }

  if (!session) {
    return (
      <MobileFrame>
        <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
          <p className="text-sm text-muted-foreground">{live.error ?? "会话不可用"}</p>
          <Link to="/login" className="text-xs text-primary">去登录</Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <SessionActionProvider
      value={
        authed
          ? { onSend: live.send, onApprove: live.approve, onStop: live.stop }
          : {}
      }
    >
      <SessionDetailBody session={session} project={project} isLive={authed} />
    </SessionActionProvider>
  );
}
