import { Link, createFileRoute } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { MobileWorkerConsoleHeader } from "@/components/mobile-worker-console-header";
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
        <MobileWorkerConsoleHeader
          mode="pty"
          podKey={podKey}
          previewAvailable={descriptor.previewAvailable}
        />
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
