import { Link, Outlet, createFileRoute, useLocation, useRouter } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { getMobileWorkerDescriptor, type MobileWorkerDescriptor } from "@/lib/mobile-pod-api";
import { resolveWorkerEntryRoute } from "@/lib/worker-entry-route";

export const Route = createFileRoute("/workers/$podKey")({
  head: ({ params }) => ({ meta: [{ title: pageTitle(`Worker · ${params.podKey}`) }] }),
  component: WorkerEntryPage,
});

function WorkerEntryPage() {
  const { podKey } = Route.useParams();
  const { pathname } = useLocation();
  const router = useRouter();
  const [descriptor, setDescriptor] = useState<MobileWorkerDescriptor | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!readAuthToken()) return;
    void getMobileWorkerDescriptor(podKey)
      .then(setDescriptor)
      .catch((cause) => {
        setError(cause instanceof Error ? cause.message : "无法打开 Worker");
      });
  }, [podKey]);

  useEffect(() => {
    if (!descriptor) return;
    const target = resolveWorkerEntryRoute(descriptor);
    if (target === "chat") {
      void router.navigate({
        to: "/workers/$podKey/chat",
        params: { podKey },
        replace: true,
      });
      return;
    }
    if (target === "terminal") {
      void router.navigate({
        to: "/workers/$podKey/terminal",
        params: { podKey },
        replace: true,
      });
    }
  }, [descriptor, podKey, router]);

  const entryRoute = descriptor ? resolveWorkerEntryRoute(descriptor) : undefined;

  if (pathname !== `/workers/${encodeURIComponent(podKey)}`) {
    return <Outlet />;
  }

  if (!readAuthToken()) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后可打开此 Worker</p>
          <Link
            to="/login"
            search={{ workerPodKey: podKey }}
            className="min-h-10 rounded-md bg-primary px-4 py-2 text-xs font-semibold text-primary-foreground"
          >
            去登录
          </Link>
        </div>
      </MobileFrame>
    );
  }

  if (error || entryRoute === null) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
          <p className="text-sm text-muted-foreground">
            {error ?? "Worker 当前不可连接，或未提供受支持的移动交互模式"}
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
