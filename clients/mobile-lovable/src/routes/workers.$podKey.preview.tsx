import { Link, createFileRoute } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { replaceWithMobilePodPreview } from "@/lib/mobile-pod-preview";

export const Route = createFileRoute("/workers/$podKey/preview")({
  head: ({ params }) => ({ meta: [{ title: pageTitle(`Preview · ${params.podKey}`) }] }),
  component: PodPreviewPage,
});

function PodPreviewPage() {
  const { podKey } = Route.useParams();
  const started = useRef(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!readAuthToken() || started.current) return;
    started.current = true;
    void replaceWithMobilePodPreview(podKey, window.location.replace.bind(window.location)).catch(
      (cause) => setError(cause instanceof Error ? cause.message : "无法打开 Preview"),
    );
  }, [podKey]);

  if (!readAuthToken()) {
    return (
      <MobileFrame hideNav>
        <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm text-muted-foreground">登录后可打开 Preview</p>
          <Link
            to="/login"
            search={{ workerPodKey: podKey, workerTarget: "preview" }}
            className="min-h-10 rounded-md bg-primary px-4 py-2 text-xs font-semibold text-primary-foreground"
          >
            去登录
          </Link>
        </div>
      </MobileFrame>
    );
  }

  return (
    <MobileFrame hideNav>
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
        {error ? (
          <>
            <p className="text-sm text-muted-foreground">{error}</p>
            <Link to="/workers/$podKey" params={{ podKey }} className="text-xs text-primary">
              打开 Worker 控制台
            </Link>
          </>
        ) : (
          <>
            <Loader2 className="h-5 w-5 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">正在打开 Preview…</p>
          </>
        )}
      </div>
    </MobileFrame>
  );
}
