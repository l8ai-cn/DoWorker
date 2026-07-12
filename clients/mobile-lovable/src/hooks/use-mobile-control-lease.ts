import { useCallback, useEffect, useRef, useState } from "react";

const RENEW_AHEAD_MS = 10_000;
const MIN_RENEW_DELAY_MS = 5_000;

export interface RelayLease {
  status: "observer" | "acquiring" | "granted";
  leaseId?: string;
  expiresAt?: number;
}

export interface MobileRelayControl {
  acquire_control(podKey: string, clientLabel: string): Promise<void>;
  renew_control(podKey: string, leaseId: string): Promise<void>;
  release_control(podKey: string, leaseId: string): Promise<void>;
}

export function useMobileControlLease(
  relay: MobileRelayControl | null,
  podKey: string | null,
  lease: RelayLease,
) {
  const [acquiring, setAcquiring] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const leaseIdRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    leaseIdRef.current = lease.status === "granted" ? lease.leaseId : undefined;
  }, [lease.leaseId, lease.status]);

  const release = useCallback(() => {
    const leaseId = leaseIdRef.current;
    if (!relay || !podKey || !leaseId) return;
    leaseIdRef.current = undefined;
    void relay.release_control(podKey, leaseId);
  }, [podKey, relay]);

  const acquire = useCallback(async () => {
    if (!relay || !podKey) return;
    setAcquiring(true);
    setError(null);
    try {
      await relay.acquire_control(podKey, "mobile");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "无法接管输入");
    } finally {
      setAcquiring(false);
    }
  }, [podKey, relay]);

  useEffect(() => {
    if (!relay || !podKey || lease.status !== "granted" || !lease.leaseId || !lease.expiresAt) {
      return;
    }
    const delay = Math.max(MIN_RENEW_DELAY_MS, lease.expiresAt - Date.now() - RENEW_AHEAD_MS);
    const timer = window.setTimeout(() => {
      void relay.renew_control(podKey, lease.leaseId!).catch((cause) => {
        setError(cause instanceof Error ? cause.message : "控制续租失败");
      });
    }, delay);
    return () => window.clearTimeout(timer);
  }, [lease.expiresAt, lease.leaseId, lease.status, podKey, relay]);

  useEffect(() => {
    const onVisibilityChange = () => {
      if (document.visibilityState === "hidden") release();
    };
    document.addEventListener("visibilitychange", onVisibilityChange);
    window.addEventListener("pagehide", release);
    return () => {
      document.removeEventListener("visibilitychange", onVisibilityChange);
      window.removeEventListener("pagehide", release);
      release();
    };
  }, [release]);

  return { acquiring, error, acquire, release };
}
