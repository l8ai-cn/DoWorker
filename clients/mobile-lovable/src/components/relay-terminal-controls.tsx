import { Loader2, LockKeyhole, RefreshCw } from "lucide-react";
import type { RelayConnectionState } from "@/lib/relay-connection-state";

interface RelayTerminalControlBarProps {
  hasControl: boolean;
  connected: boolean;
  acquiring: boolean;
  onAcquire: () => void;
}

export function RelayTerminalControlBar({
  hasControl,
  connected,
  acquiring,
  onAcquire,
}: RelayTerminalControlBarProps) {
  return (
    <div className="safe-bottom flex min-h-12 items-center justify-between border-t border-border/60 px-3 py-2">
      <span className="text-xs text-muted-foreground">
        {hasControl ? "正在控制此 Worker" : "只读观察"}
      </span>
      {hasControl ? (
        <span className="text-xs font-medium text-success">输入已启用</span>
      ) : (
        <button
          type="button"
          onClick={onAcquire}
          disabled={!connected || acquiring}
          className="flex min-h-9 items-center gap-1.5 rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground disabled:opacity-50"
        >
          {acquiring ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <LockKeyhole className="h-3.5 w-3.5" />
          )}
          接管输入
        </button>
      )}
    </div>
  );
}

interface RelayTerminalOverlayProps {
  connection: RelayConnectionState;
  error: string | null;
  onReconnect: () => void;
}

export function RelayTerminalOverlay({
  connection,
  error,
  onReconnect,
}: RelayTerminalOverlayProps) {
  if (connection === "connected") return null;
  return (
    <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-background/85 px-6 text-center backdrop-blur-sm">
      {connection !== "failed" && <Loader2 className="h-5 w-5 animate-spin text-primary" />}
      {connection === "failed" && <RefreshCw className="h-5 w-5 text-destructive" />}
      <p className="text-sm text-muted-foreground">
        {connection === "connecting"
          ? "正在连接 Worker…"
          : connection === "reconnecting"
            ? "正在重新连接 Worker…"
            : (error ?? "终端连接失败")}
      </p>
      {connection === "failed" && (
        <button
          type="button"
          onClick={onReconnect}
          className="min-h-9 rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground"
        >
          重新连接
        </button>
      )}
    </div>
  );
}
