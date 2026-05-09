"use client";

import { useRouter } from "next/navigation";
import { Laptop, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { useCurrentOrg } from "@/stores/auth";
import { useRunners, getRunnerStatusInfo } from "@/stores/runner";
import { Button } from "@/components/ui/button";
import { STEP_LABELS, useLocalRunnerOnboarding } from "@/hooks/useLocalRunnerOnboarding";

const SECTION_HEADER = "px-3 pt-3 pb-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground";

/**
 * Sidebar footer section for the user's own machine. Shows either an
 * onboarding card (when this Mac isn't registered) or a row that mirrors
 * normal runners but is keyed off the local-runner config — clicking jumps
 * to the matching runner detail.
 *
 * Renders nothing on web (no electron adapter → no local runner service).
 */
export function ThisMacSection() {
  const { unsupported, localNodeId, phase, onRegister } = useLocalRunnerOnboarding();
  const router = useRouter();
  const currentOrg = useCurrentOrg();
  const runners = useRunners();

  if (unsupported) return null;
  if (phase.kind === "loading") {
    return (
      <div className="border-t border-border">
        <div className={SECTION_HEADER}>This Mac</div>
        <div className="flex items-center gap-2 px-3 pb-3 text-xs text-muted-foreground">
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
          <span>Checking…</span>
        </div>
      </div>
    );
  }

  const matchingRunner = localNodeId
    ? runners.find((r) => r.node_id === localNodeId) ?? null
    : null;
  const isRegistered = phase.kind === "idle" && phase.status === "running" && !!localNodeId;

  return (
    <div className="border-t border-border">
      <div className={SECTION_HEADER}>This Mac</div>
      {isRegistered && matchingRunner ? (
        <RegisteredRow
          runner={matchingRunner}
          onClick={() => {
            if (currentOrg) {
              router.push(`/${currentOrg.slug}/infra?tab=runners&id=${matchingRunner.id}`);
            }
          }}
        />
      ) : isRegistered && !matchingRunner ? (
        // Registered locally but backend list hasn't picked it up yet (heartbeat in flight).
        <div
          data-testid="this-mac-syncing"
          className="px-3 pb-3 text-xs text-muted-foreground"
        >
          Active locally · syncing with server…
        </div>
      ) : (
        <OnboardingBlock
          phase={phase}
          onRegister={onRegister}
        />
      )}
    </div>
  );
}

function RegisteredRow({
  runner,
  onClick,
}: {
  runner: ReturnType<typeof useRunners>[number];
  onClick: () => void;
}) {
  const statusInfo = getRunnerStatusInfo(runner.status);
  return (
    <div
      className={cn(
        "group flex cursor-pointer items-center gap-2 px-3 py-2 hover:bg-muted/50",
      )}
      onClick={onClick}
    >
      <Laptop className="h-3.5 w-3.5 flex-shrink-0 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className={cn("h-2 w-2 flex-shrink-0 rounded-full", statusInfo.dotColor)} />
          <p className="truncate text-sm font-medium">{runner.node_id}</p>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <span>{runner.current_pods}/{runner.max_concurrent_pods} pods</span>
          {runner.host_info?.os && (
            <>
              <span>·</span>
              <span>{runner.host_info.os}</span>
            </>
          )}
          <span>·</span>
          <span className="text-green-600 dark:text-green-400">active</span>
        </div>
      </div>
    </div>
  );
}

function OnboardingBlock({
  phase,
  onRegister,
}: {
  phase: ReturnType<typeof useLocalRunnerOnboarding>["phase"];
  onRegister: () => void;
}) {
  const busy = phase.kind === "installing";
  const stepLabel = phase.kind === "installing" ? STEP_LABELS[phase.step] : null;
  const errorLine = phase.kind === "error"
    ? phase.step
      ? `Failed at ${STEP_LABELS[phase.step]}`
      : "Failed"
    : null;

  return (
    <div className="px-3 pb-3">
      <p className="mb-2 text-xs text-muted-foreground">
        Run pods locally with no cloud relay round-trip.
      </p>
      <Button
        data-testid="this-mac-register-btn"
        size="sm"
        variant={phase.kind === "error" ? "outline" : "default"}
        className="w-full h-8 text-xs"
        onClick={onRegister}
        disabled={busy}
      >
        {busy ? "Working…" : phase.kind === "error" ? "Retry" : "Register This Mac"}
      </Button>
      {stepLabel && (
        <p className="mt-2 truncate text-[11px] text-muted-foreground" title={stepLabel}>
          {stepLabel}…
        </p>
      )}
      {errorLine && phase.kind === "error" && (
        <p
          className="mt-2 truncate text-[11px] text-destructive"
          title={`${errorLine}: ${phase.message}`}
        >
          {errorLine}
        </p>
      )}
    </div>
  );
}
