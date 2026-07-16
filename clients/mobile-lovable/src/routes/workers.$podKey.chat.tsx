import { Link, createFileRoute } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileAcpPanel } from "@/components/mobile-acp-panel";
import { MobileFrame } from "@/components/mobile-frame";
import { MobileWorkerConsoleHeader } from "@/components/mobile-worker-console-header";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { getMobileWorkerDescriptor, type MobileWorkerDescriptor } from "@/lib/mobile-pod-api";

export const Route = createFileRoute("/workers/$podKey/chat")({
  head: ({ params }) => ({ meta: [{ title: pageTitle(`对话 · ${params.podKey}`) }] }),
  component: WorkerChatPage,
});

function WorkerChatPage() {
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
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后可打开对话</p>
          <Link to="/login" search={{ workerPodKey: podKey }} className="text-xs text-primary">
            去登录
          </Link>
        </div>
      </MobileFrame>
    );
  }
  if (error || !descriptor) {
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
  if (!descriptor.consoleAvailable || descriptor.interactionMode !== "acp") {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm text-muted-foreground">此 Worker 没有可用的 ACP 对话连接。</p>
          <Link to="/workers/$podKey" params={{ podKey }} className="text-xs text-primary">
            返回 Worker
          </Link>
        </div>
      </MobileFrame>
    );
  }
  return (
    <MobileFrame hideNav>
      <div className="flex h-[100dvh] flex-col">
        <MobileWorkerConsoleHeader
          mode="acp"
          podKey={podKey}
          previewAvailable={descriptor.previewAvailable}
        />
        <MobileAcpPanel podKey={podKey} />
      </div>
    </MobileFrame>
  );
}
