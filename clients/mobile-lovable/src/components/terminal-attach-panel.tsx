import { Loader2, RefreshCw } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { buildTerminalAttachUrl } from "@/lib/terminals-api";
import {
  isUnexpectedTerminalClose,
  TERMINAL_RECONNECT_BACKOFF_MS,
} from "@/lib/terminal-close-policy";
import { TerminalSession, type ConnectionState } from "@/lib/terminal-session";

interface TerminalAttachPanelProps {
  sessionId: string;
  terminalId: string;
  readOnly?: boolean;
}

export function TerminalAttachPanel({
  sessionId,
  terminalId,
  readOnly = false,
}: TerminalAttachPanelProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const sessionRef = useRef<TerminalSession | null>(null);
  const attemptRef = useRef(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [state, setState] = useState<ConnectionState>({ kind: "connecting" });
  const [isDark, setIsDark] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);

  useEffect(() => {
    setIsDark(document.documentElement.classList.contains("dark"));
  }, []);

  useEffect(() => {
    const node = containerRef.current;
    if (!node) return;

    attemptRef.current = 0;

    const clearReconnectTimer = () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
    };

    const connect = () => {
      sessionRef.current?.dispose();
      setState({ kind: "connecting" });
      const url = buildTerminalAttachUrl(sessionId, terminalId, readOnly);
      sessionRef.current = new TerminalSession(
        node,
        url,
        (next) => {
          setState(next);
          if (next.kind === "closed" && isUnexpectedTerminalClose(next.code)) {
            const idx = Math.min(attemptRef.current, TERMINAL_RECONNECT_BACKOFF_MS.length - 1);
            const delay = TERMINAL_RECONNECT_BACKOFF_MS[idx];
            attemptRef.current += 1;
            setReconnecting(true);
            clearReconnectTimer();
            reconnectTimerRef.current = setTimeout(() => {
              setReconnecting(false);
              connect();
            }, delay);
          } else if (next.kind === "connected") {
            attemptRef.current = 0;
            setReconnecting(false);
          }
        },
        isDark,
      );
    };

    connect();

    return () => {
      clearReconnectTimer();
      sessionRef.current?.dispose();
      sessionRef.current = null;
    };
  }, [sessionId, terminalId, readOnly, isDark]);

  const manualReconnect = () => {
    attemptRef.current = 0;
    const node = containerRef.current;
    if (!node) return;
    sessionRef.current?.dispose();
    setReconnecting(false);
    setState({ kind: "connecting" });
    const url = buildTerminalAttachUrl(sessionId, terminalId, readOnly);
    sessionRef.current = new TerminalSession(node, url, setState, isDark);
  };

  return (
    <div className="relative flex min-h-0 flex-1 flex-col bg-card">
      <div ref={containerRef} className="min-h-0 flex-1 overflow-hidden p-1" />
      {state.kind !== "connected" && (
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-background/80 px-4 text-center text-sm backdrop-blur-sm">
          {(state.kind === "connecting" || reconnecting) && (
            <>
              <Loader2 className="h-5 w-5 animate-spin text-primary" />
              <span className="text-muted-foreground">
                {reconnecting ? "重新连接终端…" : "连接终端…"}
              </span>
            </>
          )}
          {state.kind === "closed" && !reconnecting && (
            <>
              <span className="text-muted-foreground">终端已断开 ({state.reason})</span>
              <button
                type="button"
                onClick={manualReconnect}
                className="flex items-center gap-1 text-xs text-primary"
              >
                <RefreshCw className="h-3 w-3" /> 重新连接
              </button>
            </>
          )}
          {state.kind === "error" && (
            <>
              <span className="text-destructive">终端连接失败</span>
              <button type="button" onClick={manualReconnect} className="text-xs text-primary">
                重试
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}
