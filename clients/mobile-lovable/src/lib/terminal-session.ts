import { FitAddon } from "@xterm/addon-fit";
import { Terminal, type ITheme } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";

export type ConnectionState =
  | { kind: "connecting" }
  | { kind: "connected" }
  | { kind: "closed"; reason: string; code: number }
  | { kind: "error" };

export type ConnectionStateListener = (state: ConnectionState) => void;

const CARD_LIGHT = "#ffffff";
const CARD_DARK = "#131517";

function terminalTheme(isDark: boolean): ITheme {
  const bg = isDark ? CARD_DARK : CARD_LIGHT;
  return isDark
    ? { background: bg, foreground: "#e4e4e7", cursor: "#22d3ee", cursorAccent: bg }
    : {
        background: bg,
        foreground: "#18181b",
        cursor: "#0891b2",
        cursorAccent: bg,
        white: "#3f3f46",
        brightWhite: "#18181b",
      };
}

export class TerminalSession {
  private readonly term: Terminal;
  private readonly fit: FitAddon;
  private readonly ws: WebSocket;
  private readonly ctl = new AbortController();
  private readonly resizeObserver: ResizeObserver;
  private readonly dataDispose: { dispose: () => void };
  private disposed = false;

  constructor(
    container: HTMLElement,
    url: string,
    onState: ConnectionStateListener,
    isDark = false,
  ) {
    this.term = new Terminal({
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
      fontSize: 12,
      scrollback: 8000,
      cursorBlink: true,
      theme: terminalTheme(isDark),
      minimumContrastRatio: 4.5,
      macOptionClickForcesSelection: true,
    });
    this.fit = new FitAddon();
    this.term.loadAddon(this.fit);
    this.term.open(container);
    try {
      this.fit.fit();
    } catch {
      this.term.resize(80, 24);
    }

    this.ws = new WebSocket(url);
    this.ws.binaryType = "arraybuffer";
    const { signal } = this.ctl;

    this.ws.addEventListener(
      "open",
      () => {
        this.sendResize();
        this.term.focus();
        onState({ kind: "connected" });
      },
      { signal },
    );

    this.ws.addEventListener(
      "message",
      (ev) => {
        if (ev.data instanceof ArrayBuffer) {
          this.term.write(new Uint8Array(ev.data));
        }
      },
      { signal },
    );

    this.ws.addEventListener(
      "close",
      (ev) => onState({ kind: "closed", reason: ev.reason || `code ${ev.code}`, code: ev.code }),
      { signal },
    );

    this.ws.addEventListener("error", () => onState({ kind: "error" }), { signal });

    this.dataDispose = this.term.onData((data) => {
      if (this.ws.readyState === WebSocket.OPEN) this.ws.send(data);
    });

    this.resizeObserver = new ResizeObserver(() => this.sendResize());
    this.resizeObserver.observe(container);
    onState({ kind: "connecting" });
  }

  private sendResize(): void {
    if (this.disposed) return;
    try {
      this.fit.fit();
    } catch {
      return;
    }
    const { cols, rows } = this.term;
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: "resize", cols, rows }));
    }
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.ctl.abort();
    this.resizeObserver.disconnect();
    this.dataDispose.dispose();
    if (this.ws.readyState === WebSocket.OPEN) this.ws.close();
    this.term.dispose();
  }
}
