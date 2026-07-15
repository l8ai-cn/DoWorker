import { useEffect, useRef, useState } from "react";
import {
  RelayTerminalControlBar,
  RelayTerminalOverlay,
} from "@/components/relay-terminal-controls";
import {
  useMobileControlLease,
  type MobileRelayControl,
  type RelayLease,
} from "@/hooks/use-mobile-control-lease";
import { useRelayTerminalViewport } from "@/hooks/use-relay-terminal-viewport";
import { getMobileRelayManager } from "@/lib/mobile-relay-manager";
import { resetMobileRelayConnection } from "@/lib/mobile-relay-reset";
import { getMobilePodConnection } from "@/lib/mobile-pod-api";
import { relayConnectionState, type RelayConnectionState } from "@/lib/relay-connection-state";
import { getSessionRelayConnection } from "@/lib/session-relay-api";
import {
  observerRelayLease,
  relayLeaseFromStatus,
  relayOutput,
} from "@/lib/relay-terminal-state";

type RelayTerminalPanelProps =
  | { podKey: string; sessionId?: never }
  | { sessionId: string; podKey?: never };

export function RelayTerminalPanel(props: RelayTerminalPanelProps) {
  const directPodKey = "podKey" in props ? props.podKey : undefined;
  const sessionId = "sessionId" in props ? props.sessionId : undefined;
  const containerRef = useRef<HTMLDivElement>(null);
  const relayRef = useRef<Awaited<ReturnType<typeof getMobileRelayManager>> | null>(null);
  const podKeyRef = useRef<string | null>(null);
  const leaseRef = useRef<RelayLease>(observerRelayLease);
  const { sendSizeRef, terminalRef } = useRelayTerminalViewport({
    containerRef,
    leaseRef,
    podKeyRef,
    relayRef,
  });
  const listenerId = directPodKey
    ? `mobile-terminal-pod-${directPodKey}`
    : `mobile-terminal-session-${sessionId}`;
  const [relay, setRelay] = useState<MobileRelayControl | null>(null);
  const [podKey, setPodKey] = useState<string | null>(null);
  const [connection, setConnection] = useState<RelayConnectionState>("connecting");
  const [lease, setLease] = useState<RelayLease>(observerRelayLease);
  const [error, setError] = useState<string | null>(null);
  const [connectionAttempt, setConnectionAttempt] = useState(0);
  const control = useMobileControlLease(relay, podKey, lease);

  useEffect(() => {
    leaseRef.current = lease;
  }, [lease]);

  useEffect(() => {
    let disposed = false;

    const connect = async () => {
      try {
        const relay = await getMobileRelayManager();
        const previousPodKey = podKeyRef.current;
        if (previousPodKey) await resetMobileRelayConnection(relay, previousPodKey);
        if (disposed) return;
        leaseRef.current = observerRelayLease;
        setLease(observerRelayLease);
        setConnection("connecting");
        const info = directPodKey
          ? await getMobilePodConnection(directPodKey)
          : await getSessionRelayConnection(sessionId!);
        if (disposed) return;
        relayRef.current = relay;
        podKeyRef.current = info.podKey;
        setRelay(relay);
        setPodKey(info.podKey);
        const subscriptionId = `mobile-${info.podKey}-${Math.random().toString(36).slice(2, 10)}`;
        await relay.set_status_listener(info.podKey, listenerId, (raw: unknown) => {
          if (disposed) return;
          const status =
            raw && typeof raw === "object" ? (raw as { status?: unknown }).status : undefined;
          const nextConnection = relayConnectionState(status);
          setConnection(nextConnection);
          setLease(relayLeaseFromStatus(raw));
          if (nextConnection === "connected") terminalRef.current?.focus();
          if (nextConnection === "failed") setError("终端连接已失效，请重新连接");
        });
        if (disposed) {
          relay.remove_status_listener(info.podKey, listenerId);
          return;
        }
        await relay.subscribe(
          info.podKey,
          subscriptionId,
          info.relayUrl,
          info.token,
          (data: unknown) => {
            const output = relayOutput(data);
            if (output) terminalRef.current?.write(output);
          },
        );
        if (disposed) void resetMobileRelayConnection(relay, info.podKey);
      } catch (cause) {
        if (!disposed) {
          setConnection("failed");
          setError(cause instanceof Error ? cause.message : "终端连接失败");
        }
      }
    };
    void connect();

    return () => {
      disposed = true;
      const relay = relayRef.current;
      const podKey = podKeyRef.current;
      const lease = leaseRef.current;
      if (relay && podKey) {
        relay.remove_status_listener(podKey, listenerId);
      }
      if (relay && podKey && lease.status === "granted" && lease.leaseId) {
        void relay.release_control(podKey, lease.leaseId);
      }
      if (relay && podKey) void resetMobileRelayConnection(relay, podKey);
    };
  }, [connectionAttempt, directPodKey, sessionId, terminalRef]);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.options.disableStdin = lease.status !== "granted";
    }
    if (lease.status === "granted") sendSizeRef.current();
  }, [lease.status, sendSizeRef, terminalRef]);

  useEffect(() => {
    const refreshAfterForeground = () => {
      if (document.visibilityState === "visible") {
        setError(null);
        setConnectionAttempt((attempt) => attempt + 1);
      }
    };
    document.addEventListener("visibilitychange", refreshAfterForeground);
    return () => document.removeEventListener("visibilitychange", refreshAfterForeground);
  }, []);

  const reconnect = () => {
    setError(null);
    setConnection("connecting");
    setConnectionAttempt((attempt) => attempt + 1);
  };

  const hasControl = lease.status === "granted";
  return (
    <div className="relative flex min-h-0 flex-1 flex-col bg-card">
      <div ref={containerRef} className="min-h-0 flex-1 overflow-hidden p-1" />
      <RelayTerminalControlBar
        hasControl={hasControl}
        connected={connection === "connected"}
        acquiring={control.acquiring}
        onAcquire={() => void control.acquire()}
      />
      <RelayTerminalOverlay connection={connection} error={error} onReconnect={reconnect} />
      {control.error && <p className="px-3 py-1 text-xs text-destructive">{control.error}</p>}
    </div>
  );
}
