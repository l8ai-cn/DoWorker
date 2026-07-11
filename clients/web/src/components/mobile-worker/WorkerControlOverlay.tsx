"use client";

import { useEffect } from "react";
import { Eye, Loader2, LockKeyhole } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { useWorkerControlLease } from "@/hooks/useWorkerControlLease";
import { cn } from "@/lib/utils";

interface WorkerControlOverlayProps {
  podKey: string;
  clientLabel: string;
  preserveHeader?: boolean;
}

export function WorkerControlOverlay({
  podKey,
  clientLabel,
  preserveHeader = false,
}: WorkerControlOverlayProps) {
  const t = useTranslations("mobile.control");
  const lease = useWorkerControlLease(podKey, clientLabel);
  const controlsInput = lease.status === "granted";

  useEffect(() => {
    if (!controlsInput && document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  }, [controlsInput]);

  if (controlsInput) return null;

  const busy = lease.status === "busy";
  return (
    <div
      className={cn(
        "absolute inset-x-0 bottom-0 z-20 flex items-center justify-center bg-background/80 p-4 backdrop-blur-[1px]",
        preserveHeader ? "top-8" : "top-0",
      )}
    >
      <div className="w-full max-w-sm space-y-3 rounded-md border bg-background p-4 text-center shadow-sm">
        {busy ? (
          <LockKeyhole className="mx-auto h-6 w-6 text-warning" />
        ) : (
          <Eye className="mx-auto h-6 w-6 text-muted-foreground" />
        )}
        <div className="space-y-1">
          <p className="text-sm font-medium text-foreground">
            {busy ? t("busy") : t("observer")}
          </p>
          <p className="text-xs text-muted-foreground">
            {lease.connected ? t("takeHint") : t("waiting")}
          </p>
          {lease.error && <p className="text-xs text-danger">{lease.error}</p>}
        </div>
        <Button
          type="button"
          className="h-11 w-full"
          disabled={!lease.connected || lease.acquiring}
          onClick={() => void lease.acquire()}
        >
          {lease.acquiring && <Loader2 className="h-4 w-4 animate-spin" />}
          {t("takeControl")}
        </Button>
      </div>
    </div>
  );
}
