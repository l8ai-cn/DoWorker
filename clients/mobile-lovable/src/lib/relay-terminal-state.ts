import type { RelayLease } from "@/hooks/use-mobile-control-lease";

export const observerRelayLease: RelayLease = { status: "observer" };

export function relayTerminalTheme(isDark: boolean) {
  return isDark
    ? { background: "#131517", foreground: "#e4e4e7", cursor: "#22d3ee" }
    : { background: "#ffffff", foreground: "#18181b", cursor: "#0891b2" };
}

export function relayLeaseFromStatus(raw: unknown): RelayLease {
  if (!raw || typeof raw !== "object") return observerRelayLease;
  const value = raw as Record<string, unknown>;
  const status = value.controlLeaseStatus;
  if (status !== "acquiring" && status !== "granted") return observerRelayLease;
  return {
    status,
    leaseId: typeof value.controlLeaseId === "string" ? value.controlLeaseId : undefined,
    expiresAt:
      typeof value.controlLeaseExpiresAt === "number" ? value.controlLeaseExpiresAt : undefined,
  };
}

export function relayOutput(data: unknown): Uint8Array | string | null {
  if (typeof data === "string" || data instanceof Uint8Array) return data;
  if (data instanceof ArrayBuffer) return new Uint8Array(data);
  return null;
}
