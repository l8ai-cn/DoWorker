import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, Loader2, Terminal as TerminalIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { RelayTerminalPanel } from "@/components/relay-terminal-panel";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { getMobileWorkerDescriptor, type MobileWorkerDescriptor } from "@/lib/mobile-pod-api";

export const Route = createFileRoute("/workers/$podKey/terminal")({
  head: ({ params }) => ({ meta: [{ title: pageTitle(`命令行 · ${params.podKey}`) }] }),
  component: WorkerTerminalPage,
});

function WorkerTerminalPage() {
  const { podKey } = Route.useParams();
  const [descriptor, setDescriptor] = useState<MobileWorkerDescriptor | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!readAuthToken()) return;
    void getMobileWorkerDescriptor(podKey)
      .then(setDescriptor)
      .catch((cause) => setError(cause instanceof Error ? cause.message : "无法加载 Worker"));
  }, [podKey]);

  if (!readAuthToken()) {
    return <LoginRequired podKey={podKey} />;
  }
  if (error || !descriptor) {
    return <WorkerLoadingState error={error} />;
  }
  if (!descriptor.consoleAvailable || descriptor.interactionMode !== "pty") {
    return (
      <WorkerModeMismatch
        podKey={podKey}
        message="此 Worker 没有可用的 PTY 命令行连接。"
      />
    );
  }
  return (
    <MobileFrame hideNav>
      <div className="flex h-[100dvh] flex-col">
        <WorkerTerminalHeader podKey={podKey} />
        <RelayTerminalPanel podKey={podKey} />
      </div>
    </MobileFrame>
  );
}

function LoginRequired({ podKey }: { podKey: string }) {
  return (
    <MobileFrame hideNav>
      <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
        <p className="text-sm text-muted-foreground">登录后可使用终端</p>
        <Link to="/login" search={{ workerPodKey: podKey }} className="text-xs text-primary">
          去登录
        </Link>
      </div>
    </MobileFrame>
  );
}

function WorkerLoadingState({ error }: { error: string | null }) {
  return (
    <MobileFrame hideNav>
      <div className="flex min-h-screen items-center justify-center gap-2 text-sm text-muted-foreground">
        {error ?? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" /> 正在连接 Worker…
          </>
        )}
      </div>
    </MobileFrame>
  );
}

function WorkerModeMismatch({ podKey, message }: { podKey: string; message: string }) {
  return (
    <MobileFrame hideNav>
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
        <p className="text-sm text-muted-foreground">{message}</p>
        <Link to="/workers/$podKey" params={{ podKey }} className="text-xs text-primary">
          返回 Worker
        </Link>
      </div>
    </MobileFrame>
  );
}

function WorkerTerminalHeader({ podKey }: { podKey: string }) {
  return (
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
        <p className="truncate font-mono text-[10px] text-muted-foreground">{podKey}</p>
      </div>
    </header>
  );
}
