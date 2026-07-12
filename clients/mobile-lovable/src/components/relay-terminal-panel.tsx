import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { Loader2, LockKeyhole, RefreshCw } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import {
  useMobileControlLease,
  type MobileRelayControl,
  type RelayLease,
} from "@/hooks/use-mobile-control-lease";
import { getMobileRelayManager } from "@/lib/mobile-relay-manager";
import { relayConnectionState, type RelayConnectionState } from "@/lib/relay-connection-state";
import { getSessionRelayConnection } from "@/lib/session-relay-api";
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
  const [relay, setRelay] = useState<MobileRelayControl | null>(null);
  const [podKey, setPodKey] = useState<string | null>(null);
  const [connection, setConnection] = useState<RelayConnectionState>("connecting");
  const [lease, setLease] = useState<RelayLease>(observerRelayLease);
  const [error, setError] = useState<string | null>(null);
  const control = useMobileControlLease(relay, podKey, lease);

  useEffect(() => {
    leaseRef.current = lease;
  }, [lease]);

  useEffect(() => {
    const node = containerRef.current;
    if (!node) return;
    let disposed = false;
    let subscriptionId = "";
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
      const relay = relayRef.current;
      const podKey = podKeyRef.current;
      if (relay && podKey) void relay.send_resize(podKey, terminal.cols, terminal.rows);
    };
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
        const [relay, info] = await Promise.all([
          getMobileRelayManager(),
          getSessionRelayConnection(sessionId),
        ]);
        if (disposed) return;
        relayRef.current = relay;
        podKeyRef.current = info.podKey;
        setRelay(relay);
        setPodKey(info.podKey);
        subscriptionId = `mobile-${sessionId}-${Math.random().toString(36).slice(2, 10)}`;
        await relay.on_status_change(info.podKey, (raw: unknown) => {
          if (disposed) return;
          const status =
            raw && typeof raw === "object" ? (raw as { status?: unknown }).status : undefined;
          setConnection(relayConnectionState(status));
          setLease(relayLeaseFromStatus(raw));
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
          sendSize();
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
      const relay = relayRef.current;
      const podKey = podKeyRef.current;
      if (relay && podKey && subscriptionId) void relay.unsubscribe(podKey, subscriptionId);
      relayRef.current = null;
      podKeyRef.current = null;
    };
  }, [sessionId]);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.options.disableStdin = lease.status !== "granted";
    }
  }, [lease.status]);

  const hasControl = lease.status === "granted";
  return (
    <div className="relative flex min-h-0 flex-1 flex-col bg-card">
      <div ref={containerRef} className="min-h-0 flex-1 overflow-hidden p-1" />
      <div className="safe-bottom flex min-h-12 items-center justify-between border-t border-border/60 px-3 py-2">
        <span className="text-xs text-muted-foreground">
          {hasControl ? "正在控制此 Worker" : "只读观察"}
        </span>
        {hasControl ? (
          <span className="text-xs font-medium text-success">输入已启用</span>
        ) : (
          <button
            type="button"
            onClick={() => void control.acquire()}
            disabled={connection !== "connected" || control.acquiring}
            className="flex min-h-9 items-center gap-1.5 rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground disabled:opacity-50"
          >
            {control.acquiring ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <LockKeyhole className="h-3.5 w-3.5" />
            )}
            接管输入
          </button>
        )}
      </div>
      {connection !== "connected" && (
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-background/85 px-6 text-center backdrop-blur-sm">
          {connection !== "failed" && <Loader2 className="h-5 w-5 animate-spin text-primary" />}
          {connection === "failed" && <RefreshCw className="h-5 w-5 text-destructive" />}
          <p className="text-sm text-muted-foreground">
            {connection === "connecting"
              ? "正在连接 Worker…"
              : connection === "reconnecting"
                ? "正在重新连接 Worker…"
                : (error ?? "终端连接失败")}
          </p>
        </div>
      )}
      {control.error && <p className="px-3 py-1 text-xs text-destructive">{control.error}</p>}
    </div>
  );
}
