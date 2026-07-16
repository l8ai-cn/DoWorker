import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { useEffect, useRef, type RefObject } from "react";
import type { RelayLease } from "@/hooks/use-mobile-control-lease";
import { sendGrantedTerminalResize } from "@/lib/relay-terminal-resize";
import { relayTerminalTheme } from "@/lib/relay-terminal-state";

type TerminalRelay = {
  send(podKey: string, data: string): Promise<void>;
  send_resize(podKey: string, columns: number, rows: number): Promise<void>;
};

type RelayTerminalViewportInput = {
  containerRef: RefObject<HTMLDivElement | null>;
  leaseRef: RefObject<RelayLease>;
  podKeyRef: RefObject<string | null>;
  relayRef: RefObject<TerminalRelay | null>;
};

export function useRelayTerminalViewport({
  containerRef,
  leaseRef,
  podKeyRef,
  relayRef,
}: RelayTerminalViewportInput) {
  const terminalRef = useRef<Terminal | null>(null);
  const sendSizeRef = useRef<() => void>(() => {});

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
    let animationFrame: number | undefined;
    const scheduleSize = () => {
      if (animationFrame !== undefined) cancelAnimationFrame(animationFrame);
      animationFrame = requestAnimationFrame(() => {
        animationFrame = undefined;
        if (!disposed) sendSize();
      });
    };
    const resizeObserver = new ResizeObserver(scheduleSize);
    resizeObserver.observe(node);
    scheduleSize();
    const input = terminal.onData((data) => {
      const relay = relayRef.current;
      const podKey = podKeyRef.current;
      if (relay && podKey && leaseRef.current.status === "granted") {
        void relay.send(podKey, data);
      }
    });

    return () => {
      disposed = true;
      if (animationFrame !== undefined) cancelAnimationFrame(animationFrame);
      resizeObserver.disconnect();
      input.dispose();
      terminal.dispose();
      terminalRef.current = null;
      sendSizeRef.current = () => {};
    };
  }, [containerRef, leaseRef, podKeyRef, relayRef]);

  return { sendSizeRef, terminalRef };
}
