import "@xterm/xterm/css/xterm.css";

import { Eye, Keyboard } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type { TerminalResource, TerminalRuntime } from "./contracts";
import { startTerminalLeaseRenewal } from "./terminalControlLease";

export function TerminalSurface({
  clientLabel,
  resource,
  runtime,
}: {
  clientLabel: string;
  resource: TerminalResource;
  runtime: TerminalRuntime;
}) {
  const text = useAgentWorkspaceText();
  const containerRef = useRef<HTMLDivElement>(null);
  const resourceRef = useRef(resource);
  const hostControlled = resource.controlMode === "host";
  const leaseIdRef = useRef<string | null>(null);
  const sizeRef = useRef<{ columns: number; rows: number } | null>(null);
  const stopRenewalRef = useRef<(() => void) | null>(null);
  const [leaseId, setLeaseId] = useState<string | null>(null);
  const [status, setStatus] = useState(resource.status);
  const [error, setError] = useState<string | null>(null);
  resourceRef.current = resource;

  useEffect(() => {
    let disposed = false;
    let teardown = () => undefined;
    void Promise.all([import("@xterm/xterm"), import("@xterm/addon-fit")])
      .then(([{ Terminal }, { FitAddon }]) => {
        if (disposed || !containerRef.current) return;
        const terminal = new Terminal({
          convertEol: true,
          cursorBlink: true,
          fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
          fontSize: 13,
          theme: { background: "#111315" },
        });
        const fit = new FitAddon();
        terminal.loadAddon(fit);
        terminal.open(containerRef.current);
        fit.fit();
        const outputOff = runtime.subscribeOutput(resource.id, (bytes) =>
          terminal.write(bytes),
        );
        const statusOff = runtime.subscribeStatus(resource.id, setStatus);
        const input = terminal.onData((data) => {
          if (hostControlled || leaseIdRef.current) {
            void runtime
              .write(resource.id, new TextEncoder().encode(data))
              .catch((cause) => setError(errorMessage(cause)));
          }
        });
        const observer = new ResizeObserver(() => {
          fit.fit();
          sizeRef.current = { columns: terminal.cols, rows: terminal.rows };
          if (hostControlled || leaseIdRef.current) {
            void runtime
              .resize(resource.id, terminal.cols, terminal.rows)
              .catch((cause) => setError(errorMessage(cause)));
          }
        });
        observer.observe(containerRef.current);
        void runtime
          .connect(resourceRef.current)
          .catch((cause) => setError(errorMessage(cause)));
        teardown = () => {
          observer.disconnect();
          input.dispose();
          outputOff();
          statusOff();
          terminal.dispose();
          runtime.disconnect(resource.id);
        };
      })
      .catch((cause) => setError(errorMessage(cause)));
    return () => {
      disposed = true;
      stopRenewalRef.current?.();
      const activeLeaseId = leaseIdRef.current;
      if (activeLeaseId) {
        void Promise.resolve()
          .then(() => runtime.releaseControl(resource.id, activeLeaseId))
          .then(teardown, teardown);
        return;
      }
      teardown();
    };
  }, [hostControlled, resource.id, runtime]);

  const toggleControl = async () => {
    setError(null);
    try {
      if (leaseId) {
        await runtime.releaseControl(resource.id, leaseId);
        stopRenewalRef.current?.();
        stopRenewalRef.current = null;
        leaseIdRef.current = null;
        setLeaseId(null);
        return;
      }
      const lease = await runtime.acquireControl(resource.id, clientLabel);
      leaseIdRef.current = lease.leaseId;
      stopRenewalRef.current = startTerminalLeaseRenewal({
        expiresAt: lease.expiresAt,
        leaseId: lease.leaseId,
        onError: (cause) => {
          stopRenewalRef.current?.();
          stopRenewalRef.current = null;
          leaseIdRef.current = null;
          setLeaseId(null);
          setError(errorMessage(cause));
        },
        renew: (activeLeaseId) =>
          runtime.renewControl(resource.id, activeLeaseId),
      });
      setLeaseId(lease.leaseId);
      if (sizeRef.current) {
        await runtime.resize(
          resource.id,
          sizeRef.current.columns,
          sizeRef.current.rows,
        );
      }
    } catch (cause) {
      setError(errorMessage(cause));
    }
  };

  return (
    <section className="flex min-h-0 flex-1 flex-col bg-[#111315] text-white">
      <div className="flex h-10 items-center gap-2 border-b border-white/10 px-3 text-xs">
        <span className="truncate">{resource.label}</span>
        <span className="text-white/50">{status}</span>
        {error && <span className="truncate text-red-300">{error}</span>}
        {resource.writable && !hostControlled && (
          <button
            aria-label={leaseId ? text.releaseControl : text.takeControl}
            className="ml-auto flex h-7 items-center gap-1.5 rounded-md border border-white/20 px-2"
            onClick={() => void toggleControl()}
            type="button"
          >
            {leaseId ? (
              <Eye className="size-3.5" />
            ) : (
              <Keyboard className="size-3.5" />
            )}
            {leaseId ? text.releaseControl : text.takeControl}
          </button>
        )}
      </div>
      <div className="min-h-0 flex-1 overflow-hidden p-2" ref={containerRef} />
    </section>
  );
}

function errorMessage(value: unknown): string {
  return value instanceof Error ? value.message : String(value);
}
