"use client";

import { formatDistanceToNow } from "date-fns";
import { CircleCheck, CircleHelp, CircleX } from "lucide-react";
import { useTranslations } from "next-intl";
import type { RunnerData } from "@/lib/viewModels/runner";

interface RunnerTunnelStatusProps {
  runner: RunnerData;
}

export function RunnerTunnelStatus({ runner }: RunnerTunnelStatusProps) {
  const t = useTranslations();
  const status = tunnelStatusDisplay(runner.tunnel_state, t);

  return (
    <>
      <div>
        <dt className="text-sm text-muted-foreground">
          {t("runners.detail.outboundTunnel")}
        </dt>
        <dd className={`mt-1 flex items-center gap-1.5 text-sm font-medium ${status.className}`}>
          <status.Icon className="size-4" aria-hidden="true" />
          {status.label}
        </dd>
      </div>
      <div>
        <dt className="text-sm text-muted-foreground">
          {t("runners.detail.tunnelLastConfirmed")}
        </dt>
        <dd className="text-sm text-foreground">
          {runner.tunnel_last_seen_at
            ? formatDistanceToNow(new Date(runner.tunnel_last_seen_at), { addSuffix: true })
            : "-"}
        </dd>
      </div>
      {runner.tunnel_last_error && (
        <div>
          <dt className="text-sm text-muted-foreground">
            {t("runners.detail.tunnelError")}
          </dt>
          <dd className="font-mono text-sm text-foreground">{runner.tunnel_last_error}</dd>
        </div>
      )}
    </>
  );
}

function tunnelStatusDisplay(
  state: RunnerData["tunnel_state"],
  t: ReturnType<typeof useTranslations>,
) {
  if (state === "connected") {
    return { className: "text-success", Icon: CircleCheck, label: t("runners.detail.tunnelConnected") };
  }
  if (state === "disconnected") {
    return { className: "text-destructive", Icon: CircleX, label: t("runners.detail.tunnelDisconnected") };
  }
  return { className: "text-muted-foreground", Icon: CircleHelp, label: t("runners.detail.tunnelNotReported") };
}
