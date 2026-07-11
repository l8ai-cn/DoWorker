"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { relayPool, type ControlLeaseStatus } from "@/stores/relayConnection";
import { useTerminalStatus } from "@/hooks/useTerminalStatus";

const RENEW_AHEAD_MS = 10_000;
const MIN_RENEW_DELAY_MS = 5_000;

export interface WorkerControlLease {
  status: ControlLeaseStatus;
  connected: boolean;
  acquiring: boolean;
  error: string | null;
  acquire: () => Promise<void>;
}

export function useWorkerControlLease(
  podKey: string,
  clientLabel: string,
): WorkerControlLease {
  const relay = useTerminalStatus(podKey);
  const [acquiring, setAcquiring] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const leaseIdRef = useRef<string | undefined>(undefined);
  const releasedLeaseIdRef = useRef<string | undefined>(undefined);
  const status = relay.controlLease.status;
  const connected = relay.status === "connected";

  if (
    relay.controlLease.leaseId &&
    relay.controlLease.leaseId !== releasedLeaseIdRef.current
  ) {
    leaseIdRef.current = relay.controlLease.leaseId;
  }

  const acquire = useCallback(async () => {
    setAcquiring(true);
    setError(null);
    try {
      await relayPool.acquireControl(podKey, clientLabel);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setAcquiring(false);
    }
  }, [clientLabel, podKey]);

  useEffect(() => {
    const leaseId = relay.controlLease.leaseId;
    const expiresAt = relay.controlLease.expiresAt;
    if (status !== "granted" || !leaseId || !expiresAt) return;
    const delay = Math.max(MIN_RENEW_DELAY_MS, expiresAt - Date.now() - RENEW_AHEAD_MS);
    const timer = window.setTimeout(() => {
      void relayPool.renewControl(podKey, leaseId).catch((cause) => {
        setError(cause instanceof Error ? cause.message : String(cause));
      });
    }, delay);
    return () => window.clearTimeout(timer);
  }, [podKey, relay.controlLease.expiresAt, relay.controlLease.leaseId, status]);

  useEffect(() => {
    const release = () => {
      const leaseId = leaseIdRef.current;
      if (!leaseId) return;
      leaseIdRef.current = undefined;
      releasedLeaseIdRef.current = leaseId;
      void relayPool.releaseControl(podKey, leaseId);
    };
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
  }, [podKey]);

  return { status, connected, acquiring, error, acquire };
}
