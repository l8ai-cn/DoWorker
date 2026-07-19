import type { ReactNode } from "react";
import type { Host } from "@/hooks/useHosts";

export function harnessUnconfiguredOnHost(
  harness: string | null | undefined,
  host: Host | undefined | null,
): boolean {
  return harnessUnavailableReasonOnHost(harness, host) !== null;
}

export function harnessUnavailableReasonOnHost(
  harness: string | null | undefined,
  host: Host | undefined | null,
): string | null {
  if (!harness || !host?.configured_harnesses) return null;
  const availability = host.configured_harnesses[harness];
  if (availability === false) return isCodexHarness(harness) ? "binary-missing" : "unconfigured";
  if (
    isCodexHarness(harness) &&
    (availability === "binary-missing" || availability === "needs-auth")
  ) {
    return availability;
  }
  return null;
}

export function harnessWarningBadgeText(reason: string | null): string {
  if (reason === "binary-missing") return "binary missing";
  if (reason === "needs-auth") return "needs auth";
  return "needs setup";
}

export function harnessWarningMessageText(
  agentName: string | undefined,
  hostName: string | undefined,
  reason: string | null,
): string {
  if (reason === "needs-auth") {
    return `${agentName} needs Codex authentication on ${hostName} — run codex login on that machine.`;
  }
  if (reason === "binary-missing") {
    return `${agentName} is missing the Codex binary on ${hostName} — run runner register on that machine.`;
  }
  return `${agentName} isn't configured on ${hostName} — run runner register on that machine.`;
}

export function HarnessWarningMessage({
  agentName,
  hostName,
  reason,
}: {
  agentName: string | undefined;
  hostName: string | undefined;
  reason: string | null;
}): ReactNode {
  if (reason === "needs-auth") {
    return (
      <>
        {agentName} needs Codex authentication on {hostName} — run <code>codex login</code> on that
        machine.
      </>
    );
  }
  if (reason === "binary-missing") {
    return (
      <>
        {agentName} is missing the Codex binary on {hostName} — run <code>runner register</code> on
        that machine.
      </>
    );
  }
  return (
    <>
      {agentName} isn&apos;t configured on {hostName} — run <code>runner register</code> on that
      machine.
    </>
  );
}

function isCodexHarness(harness: string): boolean {
  return harness === "codex" || harness === "codex-native" || harness === "native-codex";
}
