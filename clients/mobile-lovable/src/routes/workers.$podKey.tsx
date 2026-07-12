import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { getSessionByPodKey, type LiveSessionSummary } from "@/lib/sessions-api";

export const Route = createFileRoute("/workers/$podKey")({
  head: ({ params }) => ({ meta: [{ title: pageTitle(`Worker · ${params.podKey}`) }] }),
  component: WorkerEntryPage,
});

function WorkerEntryPage() {
  const { podKey } = Route.useParams();
  const router = useRouter();
  const [session, setSession] = useState<LiveSessionSummary | null | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!readAuthToken()) return;
    void getSessionByPodKey(podKey)
      .then(setSession)
      .catch((cause) => {
        setError(cause instanceof Error ? cause.message : "无法打开 Worker");
      });
  }, [podKey]);

  useEffect(() => {
    if (!session) return;
    if (session.interactionMode === "acp") {
      void router.navigate({
        to: "/sessions/$sessionId",
        params: { sessionId: session.id },
        replace: true,
      });
      return;
    }
    if (session.interactionMode === "pty") {
      void router.navigate({
        to: "/sessions/$sessionId/terminal",
        params: { sessionId: session.id },
        replace: true,
      });
    }
  }, [router, session]);

  const missingInteractionMode =
    session !== undefined &&
    session !== null &&
    session.interactionMode !== "acp" &&
    session.interactionMode !== "pty";

  if (!readAuthToken()) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后可打开此 Worker</p>
          <Link
            to="/login"
            className="min-h-10 rounded-md bg-primary px-4 py-2 text-xs font-semibold text-primary-foreground"
          >
            去登录
          </Link>
        </div>
      </MobileFrame>
    );
  }

  if (error || session === null || missingInteractionMode) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
          <p className="text-sm text-muted-foreground">
            {error ??
              (missingInteractionMode
                ? "Worker 缺少交互模式，无法安全打开"
                : "此 Worker 没有关联的移动会话")}
          </p>
          <Link to="/" className="text-xs text-primary">
            返回会话列表
          </Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <MobileFrame hideNav>
      <div className="flex min-h-screen items-center justify-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" /> 正在打开 Worker…
      </div>
    </MobileFrame>
  );
}
