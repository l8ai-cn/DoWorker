"use client";

import { useEffect } from "react";
import { Eye, Loader2, LockKeyhole } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import type { WorkerControlLease } from "@/hooks/useWorkerControlLease";
import { cn } from "@/lib/utils";

interface WorkerControlOverlayProps {
  lease: WorkerControlLease;
  blocking?: boolean;
  preserveHeader?: boolean;
}

export function WorkerControlOverlay({
  lease,
  blocking = true,
  preserveHeader = false,
}: WorkerControlOverlayProps) {
  const t = useTranslations("mobile.control");
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
        "absolute z-20 flex p-3",
        blocking
          ? "inset-x-0 bottom-0 items-center justify-center bg-background/80 backdrop-blur-[1px]"
          : "inset-x-0 bottom-0 pointer-events-none justify-center max-sm:inset-x-auto max-sm:bottom-3 max-sm:right-3",
        blocking && (preserveHeader ? "top-8" : "top-0"),
      )}
    >
      <div
        className={cn(
          "w-full rounded-md border bg-background shadow-sm",
          blocking
            ? "max-w-sm space-y-3 p-4 text-center"
            : "pointer-events-auto flex max-w-xl items-center gap-3 p-3 max-sm:w-auto max-sm:gap-2 max-sm:p-2",
        )}
      >
        {busy ? (
          <LockKeyhole className={cn("h-5 w-5 shrink-0 text-warning", blocking && "mx-auto")} />
        ) : (
          <Eye className={cn("h-5 w-5 shrink-0 text-muted-foreground", blocking && "mx-auto")} />
        )}
        <div className={cn("space-y-1", !blocking && "min-w-0 flex-1 max-sm:hidden")}>
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
          className={cn("h-11", blocking ? "w-full" : "shrink-0")}
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
