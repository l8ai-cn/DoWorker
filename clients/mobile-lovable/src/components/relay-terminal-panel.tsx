import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
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
import { getMobileRelayManager } from "@/lib/mobile-relay-manager";
import { resetMobileRelayConnection } from "@/lib/mobile-relay-reset";
import { relayConnectionState, type RelayConnectionState } from "@/lib/relay-connection-state";
import { getSessionRelayConnection } from "@/lib/session-relay-api";
import { sendGrantedTerminalResize } from "@/lib/relay-terminal-resize";
import {
  observerRelayLease,
  relayLeaseFromStatus,
  relayOutput,
  relayTerminalTheme,
} from "@/lib/relay-terminal-state";

export function RelayTerminalPanel({ sessionId }: { sessionId: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<Terminal | null>(null);
  const relayRef = useRef<Awaited<ReturnType<typeof getMobileRelayManager>> | null>(null);
  const podKeyRef = useRef<string | null>(null);
  const leaseRef = useRef<RelayLease>(observerRelayLease);
  const sendSizeRef = useRef<() => void>(() => {});
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
    const node = containerRef.current;
    if (!node) return;
    let disposed = false;
    const terminal = new Terminal({
      cursorBlink: true,
      disableStdin: true,
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
      fontSize: 13,
      minimumContrastRatio: 4.5,
      scrollback: 8000,
      theme: relayTerminalTheme(document.documentElement.classList.contains("dark")),
    });
    const fit = new FitAddon();
    terminal.loadAddon(fit);
    terminal.open(node);
    terminalRef.current = terminal;
    try {
      fit.fit();
    } catch {
      terminal.resize(80, 24);
    }

    const sendSize = () => {
      try {
        fit.fit();
      } catch {
        return;
      }
      sendGrantedTerminalResize(
        relayRef.current,
        podKeyRef.current,
        leaseRef.current,
        terminal.cols,
        terminal.rows,
      );
    };
    sendSizeRef.current = sendSize;
    const resizeObserver = new ResizeObserver(sendSize);
    resizeObserver.observe(node);
    const input = terminal.onData((data) => {
      const relay = relayRef.current;
      const podKey = podKeyRef.current;
      if (relay && podKey && leaseRef.current.status === "granted") {
        void relay.send(podKey, data);
      }
    });

    const connect = async () => {
      try {
        const relay = await getMobileRelayManager();
        const previousPodKey = podKeyRef.current;
        if (previousPodKey) await resetMobileRelayConnection(relay, previousPodKey);
        if (disposed) return;
        leaseRef.current = observerRelayLease;
        setLease(observerRelayLease);
        setConnection("connecting");
        const info = await getSessionRelayConnection(sessionId);
        if (disposed) return;
        relayRef.current = relay;
        podKeyRef.current = info.podKey;
        setRelay(relay);
        setPodKey(info.podKey);
        const subscriptionId = `mobile-${sessionId}-${Math.random().toString(36).slice(2, 10)}`;
        await relay.on_status_change(info.podKey, (raw: unknown) => {
          if (disposed) return;
          const status =
            raw && typeof raw === "object" ? (raw as { status?: unknown }).status : undefined;
          const nextConnection = relayConnectionState(status);
          setConnection(nextConnection);
          setLease(relayLeaseFromStatus(raw));
          if (nextConnection === "failed") setError("终端连接已失效，请重新连接");
        });
        await relay.subscribe(
          info.podKey,
          subscriptionId,
          info.relayUrl,
          info.token,
          (data: unknown) => {
            const output = relayOutput(data);
            if (output) terminal.write(output);
          },
        );
        if (!disposed) {
          terminal.focus();
        }
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
      resizeObserver.disconnect();
      input.dispose();
      terminal.dispose();
      terminalRef.current = null;
      sendSizeRef.current = () => {};
      const relay = relayRef.current;
      const podKey = podKeyRef.current;
      const lease = leaseRef.current;
      if (relay && podKey && lease.status === "granted" && lease.leaseId) {
        void relay.release_control(podKey, lease.leaseId);
      }
      if (relay && podKey) void resetMobileRelayConnection(relay, podKey);
    };
  }, [connectionAttempt, sessionId]);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.options.disableStdin = lease.status !== "granted";
    }
    if (lease.status === "granted") sendSizeRef.current();
  }, [lease.status]);

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
