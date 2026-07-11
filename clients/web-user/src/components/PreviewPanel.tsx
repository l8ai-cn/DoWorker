import { useState } from "react";
import {
  CloudOffIcon,
  ExternalLinkIcon,
  Loader2Icon,
  RefreshCwIcon,
  ShieldAlertIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { buildPreviewSrc, PodPreviewError, usePodPreview } from "@/hooks/usePodPreview";

export interface PreviewPanelProps {
  /** The pod whose loopback HTTP/WS services this panel previews. */
  podKey: string;
  className?: string;
}

/**
 * Embeds a pod's HTTP preview (dev server, static site, etc.) via the
 * Gateway's `/preview/{podKey}/*` HTTP data plane. The iframe uses the
 * session URL so preview navigation can keep cookie-based auth separate from
 * query-string credentials.
 */
export function PreviewPanel({ podKey, className }: PreviewPanelProps) {
  const query = usePodPreview(podKey);
  const [reloadNonce, setReloadNonce] = useState(0);

  const openInNewWindow = () => {
    if (!query.data) return;
    window.open(buildPreviewSrc(query.data), "_blank", "noopener,noreferrer");
  };

  const refresh = () => {
    setReloadNonce((n) => n + 1);
    void query.refetch();
  };

  const showLoading = query.isPending || (query.isFetching && !query.data);

  return (
    <div className={cn("flex min-h-0 flex-1 flex-col overflow-hidden bg-card", className)}>
      <div className="flex shrink-0 items-center justify-between gap-2 border-b border-border px-3 py-2">
        <span className="text-xs font-medium text-muted-foreground">Preview</span>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label="Refresh preview"
            onClick={refresh}
            disabled={query.isFetching}
          >
            <RefreshCwIcon className={cn("size-4", query.isFetching && "animate-spin")} />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label="Open preview in new window"
            onClick={openInNewWindow}
            disabled={!query.data}
          >
            <ExternalLinkIcon className="size-4" />
          </Button>
        </div>
      </div>
      <div className="relative min-h-0 flex-1">
        {showLoading ? (
          <div className="flex h-full flex-col items-center justify-center gap-2 text-sm text-muted-foreground">
            <Loader2Icon className="size-5 animate-spin" />
            <span>Loading preview…</span>
          </div>
        ) : query.isError ? (
          <PreviewErrorState error={query.error} onRetry={refresh} />
        ) : query.data ? (
          <iframe
            key={getPreviewFrameKey(query.data, reloadNonce)}
            title={`Pod ${podKey} preview`}
            src={buildPreviewSrc(query.data)}
            sandbox="allow-scripts allow-same-origin allow-forms"
            className="h-full w-full border-0"
          />
        ) : null}
      </div>
    </div>
  );
}

export function getPreviewFrameKey(data: { session_url: string; expires_at: string }, nonce: number): string {
  return `${data.session_url}:${data.expires_at}:${nonce}`;
}

function PreviewErrorState({
  error,
  onRetry,
}: {
  error: Error | null;
  onRetry: () => void;
}) {
  const status = error instanceof PodPreviewError ? error.status : undefined;
  const { Icon, title, detail } = describeError(status);
  return (
    <div className="flex h-full flex-col items-center justify-center gap-2 p-8 text-center text-sm text-muted-foreground">
      <Icon className="size-6 opacity-60" />
      <span className="font-medium text-foreground">{title}</span>
      <span>{detail}</span>
      <Button type="button" variant="outline" size="sm" onClick={onRetry} className="mt-2">
        Retry
      </Button>
    </div>
  );
}

function describeError(status: number | undefined): {
  Icon: typeof CloudOffIcon;
  title: string;
  detail: string;
} {
  switch (status) {
    case 403:
      return {
        Icon: ShieldAlertIcon,
        title: "Access denied",
        detail: "You don't have permission to preview this pod.",
      };
    case 404:
      return {
        Icon: CloudOffIcon,
        title: "Preview unavailable",
        detail: "Preview is not enabled for this pod.",
      };
    case 409:
      return {
        Icon: CloudOffIcon,
        title: "Pod not active",
        detail: "Start the pod to preview it.",
      };
    case 503:
      return {
        Icon: CloudOffIcon,
        title: "Preview offline",
        detail: "The pod's tunnel isn't connected right now.",
      };
    default:
      return {
        Icon: CloudOffIcon,
        title: "Couldn't load preview",
        detail: "Something went wrong loading the preview.",
      };
  }
}
