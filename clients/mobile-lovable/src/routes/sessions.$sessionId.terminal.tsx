import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, Loader2, Terminal as TerminalIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { RelayTerminalPanel } from "@/components/relay-terminal-panel";
import { SessionViewToggle } from "@/components/session-view-toggle";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { getSession, type LiveSessionSummary } from "@/lib/sessions-api";

export const Route = createFileRoute("/sessions/$sessionId/terminal")({
  head: ({ params }) => ({ meta: [{ title: pageTitle(`终端 · ${params.sessionId}`) }] }),
  component: SessionTerminalPage,
});

function SessionTerminalPage() {
  const { sessionId } = Route.useParams();
  const [session, setSession] = useState<LiveSessionSummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!readAuthToken()) return;
    void getSession(sessionId)
      .then(setSession)
      .catch((cause) => {
        setError(cause instanceof Error ? cause.message : "加载会话失败");
      });
  }, [sessionId]);

  if (!readAuthToken()) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后可使用终端</p>
          <Link to="/login" className="text-xs text-primary">
            去登录
          </Link>
        </div>
      </MobileFrame>
    );
  }

  if (error || !session) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen items-center justify-center gap-2 text-sm text-muted-foreground">
          {error ?? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" /> 正在加载会话…
            </>
          )}
        </div>
      </MobileFrame>
    );
  }

  if (session.interactionMode !== "pty") {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm font-medium">此 Worker 以对话模式启动</p>
          <p className="text-xs text-muted-foreground">ACP Worker 不提供 PTY 命令行。</p>
          <Link
            to="/sessions/$sessionId"
            params={{ sessionId }}
            className="min-h-10 rounded-md bg-primary px-4 py-2 text-xs font-semibold text-primary-foreground"
          >
            打开对话
          </Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <MobileFrame hideNav>
      <div className="flex h-[100dvh] flex-col">
        <header className="safe-top flex shrink-0 items-center gap-2 border-b border-border/60 bg-background/90 px-3 pb-2 pt-2 backdrop-blur-xl">
          <Link
            to="/"
            className="flex h-9 w-9 items-center justify-center rounded-md hover:bg-surface"
            aria-label="返回会话列表"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div className="min-w-0 flex-1">
            <p className="flex items-center gap-1.5 text-sm font-semibold">
              <TerminalIcon className="h-4 w-4 text-primary" /> 命令行
            </p>
            <p className="truncate font-mono text-[10px] text-muted-foreground">{sessionId}</p>
          </div>
          <SessionViewToggle
            sessionId={sessionId}
            mode="terminal"
            interactionMode={session.interactionMode}
          />
        </header>
        <RelayTerminalPanel sessionId={sessionId} />
      </div>
    </MobileFrame>
  );
}
