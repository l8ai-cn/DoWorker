"use client";

import { cn } from "@/lib/utils";
import { useTranslations } from "next-intl";
import type { ConnectionStatus } from "@/stores/relayConnection";

type SegmentStatus = "ok" | "connecting" | "warning" | "unknown";

interface RelayStatusOverlayProps {
  connectionStatus: ConnectionStatus;
  isRunnerDisconnected: boolean;
  className?: string;
}

function deriveSegments(
  connectionStatus: ConnectionStatus,
  isRunnerDisconnected: boolean,
): { webRelay: SegmentStatus; relayRunner: SegmentStatus } {
  const webRelay: SegmentStatus =
    connectionStatus === "connected" ? "ok"
    : connectionStatus === "connecting" ? "connecting"
    : "warning";

  const relayRunner: SegmentStatus =
    connectionStatus !== "connected" ? "unknown"
    : isRunnerDisconnected ? "warning"
    : "ok";

  return { webRelay, relayRunner };
}

function worstStatus(a: SegmentStatus, b: SegmentStatus): SegmentStatus {
  const priority: Record<SegmentStatus, number> = { warning: 3, connecting: 2, unknown: 1, ok: 0 };
  return priority[a] >= priority[b] ? a : b;
}

const dotColor: Record<SegmentStatus, string> = {
  ok: "bg-success",
  connecting: "bg-warning animate-pulse",
  warning: "bg-danger",
  unknown: "bg-muted-foreground",
};

const lineColor: Record<SegmentStatus, string> = {
  ok: "bg-success/60",
  connecting: "bg-warning/60",
  warning: "bg-danger/60",
  unknown: "bg-muted-foreground/40",
};

const badgeBg: Record<SegmentStatus, string> = {
  ok: "bg-success/15 border-success/20",
  connecting: "bg-warning/15 border-warning/20",
  warning: "bg-danger/15 border-danger/20",
  unknown: "bg-muted-foreground/15 border-muted-foreground/20",
};

const labelColor: Record<SegmentStatus, string> = {
  ok: "text-success",
  connecting: "text-warning",
  warning: "text-danger",
  unknown: "text-muted-foreground",
};

function webRelayTooltipKey(connectionStatus: ConnectionStatus): string {
  switch (connectionStatus) {
    case "connected": return "connected";
    case "connecting": return "connecting";
    case "disconnected": return "disconnected";
    case "error": return "error";
  }
}

function relayRunnerTooltipKey(connectionStatus: ConnectionStatus, disconnected: boolean): string {
  if (connectionStatus !== "connected") return "unknown";
  return disconnected ? "disconnected" : "connected";
}

export function RelayStatusOverlay({
  connectionStatus,
  isRunnerDisconnected,
  className,
}: RelayStatusOverlayProps) {
  const t = useTranslations("relayStatus");
  const { webRelay, relayRunner } = deriveSegments(connectionStatus, isRunnerDisconnected);
  const overall = worstStatus(webRelay, relayRunner);
  const webRelayTip = t(webRelayTooltipKey(connectionStatus));
  const relayRunnerTip = t(relayRunnerTooltipKey(connectionStatus, isRunnerDisconnected));

  return (
    <div
      className={cn(
        "absolute top-0 left-0 right-0 z-10 flex items-center justify-center pointer-events-none",
        className,
      )}
    >
      <div
        className={cn(
          "inline-flex items-center gap-1 px-2.5 py-0.5 rounded-b-md text-[11px] font-medium",
          "shadow-sm backdrop-blur-sm transition-colors duration-300 border-x border-b",
          badgeBg[overall],
        )}
      >
        <span className={labelColor[webRelay]}>{t("web")}</span>
        <SegmentDot status={webRelay} title={webRelayTip} />
        <span className={cn("h-px w-3 inline-block", lineColor[webRelay])} />
        <span className="text-muted-foreground">{t("relay")}</span>
        <span className={cn("h-px w-3 inline-block", lineColor[relayRunner])} />
        <SegmentDot status={relayRunner} title={relayRunnerTip} />
        <span className={labelColor[relayRunner]}>{t("runner")}</span>
      </div>
    </div>
  );
}

function SegmentDot({ status, title }: { status: SegmentStatus; title: string }) {
  return (
    <span
      className={cn("w-1.5 h-1.5 rounded-full inline-block flex-shrink-0", dotColor[status])}
      title={title}
      role="status"
      aria-label={title}
    />
  );
}

export default RelayStatusOverlay;
