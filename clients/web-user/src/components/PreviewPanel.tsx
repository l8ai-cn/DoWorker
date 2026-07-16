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
  podKey: string;
  className?: string;
}

export function buildPreviewWindowUrl(
  info: { session_url: string },
  appOrigin = window.location.origin,
): string {
  const shellUrl = new URL("/preview-window.html", appOrigin);
  shellUrl.hash = encodeURIComponent(info.session_url);
  return shellUrl.href;
}

export function PreviewPanel({ podKey, className }: PreviewPanelProps) {
  const query = usePodPreview(podKey);
  const [reloadNonce, setReloadNonce] = useState(0);

  const openInNewWindow = () => {
    if (!query.data) return;
    window.open(
      buildPreviewWindowUrl(query.data),
      "_blank",
      "noopener,noreferrer",
    );
  };

  const refresh = () => {
    setReloadNonce((n) => n + 1);
    void query.refetch();
  };

  const showLoading = query.isPending || (query.isFetching && !query.data);

  return (
    <div className={cn("flex min-h-0 flex-1 flex-col overflow-hidden bg-card", className)}>
      <div className="flex shrink-0 items-center justify-between gap-2 border-b border-border px-3 py-2">
        <span className="text-xs font-medium text-muted-foreground">应用预览</span>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label="刷新预览"
            onClick={refresh}
            disabled={query.isFetching}
          >
            <RefreshCwIcon className={cn("size-4", query.isFetching && "animate-spin")} />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label="在新窗口打开预览"
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
            <span>正在加载预览</span>
          </div>
        ) : query.isError ? (
          <PreviewErrorState error={query.error} onRetry={refresh} />
        ) : query.data ? (
          <iframe
            key={getPreviewFrameKey(query.data, reloadNonce)}
            title={`Pod ${podKey} preview`}
            src={buildPreviewSrc(query.data)}
            sandbox="allow-scripts allow-same-origin allow-forms allow-downloads"
            referrerPolicy="no-referrer"
            allow="fullscreen 'self'"
            allowFullScreen
            className="h-full w-full border-0"
          />
        ) : null}
      </div>
    </div>
  );
}

export function getPreviewFrameKey(
  data: { session_url: string; expires_at: string },
  nonce: number,
): string {
  return `${data.session_url}:${data.expires_at}:${nonce}`;
}

function PreviewErrorState({ error, onRetry }: { error: Error | null; onRetry: () => void }) {
  const status = error instanceof PodPreviewError ? error.status : undefined;
  const { Icon, title, detail } = describeError(status);
  return (
    <div className="flex h-full flex-col items-center justify-center gap-2 p-8 text-center text-sm text-muted-foreground">
      <Icon className="size-6 opacity-60" />
      <span className="font-medium text-foreground">{title}</span>
      <span>{detail}</span>
      <Button type="button" variant="outline" size="sm" onClick={onRetry} className="mt-2">
        重试
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
        title: "无权访问",
        detail: "你没有查看这个工作区预览的权限。",
      };
    case 404:
      return {
        Icon: CloudOffIcon,
        title: "预览不可用",
        detail: "这个工作区尚未启用预览服务。",
      };
    case 409:
      return {
        Icon: CloudOffIcon,
        title: "工作区未运行",
        detail: "启动工作区后才能查看预览。",
      };
    case 503:
      return {
        Icon: CloudOffIcon,
        title: "预览已离线",
        detail: "工作区的预览通道当前未连接。",
      };
    default:
      return {
        Icon: CloudOffIcon,
        title: "无法加载预览",
        detail: "加载预览时发生错误。",
      };
  }
}
