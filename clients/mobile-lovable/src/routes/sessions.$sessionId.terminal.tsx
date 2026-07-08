import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, Loader2, MessageSquare, RefreshCw, Terminal as TerminalIcon } from "lucide-react";
import { MobileFrame } from "@/components/mobile-frame";
import { TerminalAttachPanel } from "@/components/terminal-attach-panel";
import { useSessionTerminal } from "@/hooks/useSessionTerminal";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";

export const Route = createFileRoute("/sessions/$sessionId/terminal")({
  head: ({ params }) => ({
    meta: [{ title: pageTitle(`终端 · ${params.sessionId}`) }],
  }),
  component: SessionTerminalPage,
});

function SessionTerminalPage() {
  const { sessionId } = Route.useParams();
  const authed = Boolean(readAuthToken());
  const { terminal, loading, error, refresh } = useSessionTerminal(sessionId);

  if (!authed) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后可 attach 终端</p>
          <Link to="/login" className="text-xs text-primary">去登录</Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <MobileFrame hideNav>
      <div className="flex h-[100dvh] flex-col">
        <header className="safe-top flex shrink-0 items-center gap-2 border-b border-border/60 bg-background/90 px-3 pb-2 pt-2 backdrop-blur-xl">
          <Link
            to="/sessions/$sessionId"
            params={{ sessionId }}
            className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div className="min-w-0 flex-1">
            <p className="flex items-center gap-1.5 text-[13px] font-semibold">
              <TerminalIcon className="h-3.5 w-3.5 text-primary" />
              CLI 终端
            </p>
            <p className="truncate font-mono text-[10px] text-muted-foreground">{sessionId}</p>
          </div>
          <Link
            to="/sessions/$sessionId"
            params={{ sessionId }}
            className="flex items-center gap-1 rounded-full bg-surface px-2.5 py-1 text-[10px] font-medium text-muted-foreground ring-1 ring-border/60"
          >
            <MessageSquare className="h-3 w-3" />
            聊天
          </Link>
          <button
            type="button"
            onClick={() => void refresh()}
            className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface"
            aria-label="刷新终端列表"
          >
            <RefreshCw className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </header>

        {loading && (
          <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
            <Loader2 className="mr-2 h-4 w-4 animate-spin" /> 查找终端…
          </div>
        )}

        {!loading && error && (
          <div className="flex flex-1 flex-col items-center justify-center gap-2 px-6 text-center text-sm text-destructive">
            <p>{error}</p>
            <button type="button" onClick={() => void refresh()} className="text-xs text-primary">
              重试
            </button>
          </div>
        )}

        {!loading && !error && !terminal && (
          <div className="flex flex-1 flex-col items-center justify-center gap-2 px-6 text-center text-sm text-muted-foreground">
            <p>会话尚未暴露终端资源</p>
            <p className="text-[11px]">Agent 启动后（codex-native-ui / claude-native-ui）会自动创建终端</p>
            <button type="button" onClick={() => void refresh()} className="mt-2 text-xs text-primary">
              刷新
            </button>
          </div>
        )}

        {terminal && (
          <TerminalAttachPanel sessionId={sessionId} terminalId={terminal.id} />
        )}
      </div>
    </MobileFrame>
  );
}
