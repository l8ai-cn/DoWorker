import { useCallback, useEffect, useRef, useState } from "react";
import { useMobileControlLease, type RelayLease } from "@/hooks/use-mobile-control-lease";
import { getMobilePodConnection } from "@/lib/mobile-pod-api";
import {
  applyMobileAcpRelayMessage,
  readMobileAcpSession,
  type MobileAcpSession,
} from "@/lib/mobile-acp-session";
import { createMobileAcpPromptConfirmation } from "@/lib/mobile-acp-prompt-confirmation";
import { getMobileRelayManager } from "@/lib/mobile-relay-manager";
import { resetMobileRelayConnection } from "@/lib/mobile-relay-reset";
import { getMobileAcpManager } from "@/lib/mobile-wasm";
import { relayConnectionState, type RelayConnectionState } from "@/lib/relay-connection-state";
import { observerRelayLease, relayLeaseFromStatus } from "@/lib/relay-terminal-state";

const EMPTY_SESSION: MobileAcpSession = {
  state: "idle",
  messages: [],
  pendingPermissions: [],
};

let nextPromptRequestId = 0;

export function useMobileAcpRelay(podKey: string) {
  const relayRef = useRef<Awaited<ReturnType<typeof getMobileRelayManager>> | null>(null);
  const leaseRef = useRef<RelayLease>(observerRelayLease);
  const promptConfirmationRef = useRef(createMobileAcpPromptConfirmation());
  const listenerId = `mobile-acp-${podKey}`;
  const [relayControl, setRelayControl] = useState<
    Awaited<ReturnType<typeof getMobileRelayManager>> | null
  >(null);
  const [connection, setConnection] = useState<RelayConnectionState>("connecting");
  const [lease, setLease] = useState<RelayLease>(observerRelayLease);
  const [session, setSession] = useState<MobileAcpSession>(EMPTY_SESSION);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const control = useMobileControlLease(relayControl, podKey, lease);

  useEffect(() => {
    leaseRef.current = lease;
  }, [lease]);

  useEffect(() => {
    let disposed = false;
    let relay: Awaited<ReturnType<typeof getMobileRelayManager>> | null = null;
    let lease: RelayLease = observerRelayLease;

    const connect = async () => {
      try {
        const [nextRelay, manager, info] = await Promise.all([
          getMobileRelayManager(),
          getMobileAcpManager(),
          getMobilePodConnection(podKey),
        ]);
        if (disposed) return;
        relay = nextRelay;
        relayRef.current = nextRelay;
        setRelayControl(nextRelay);
        setConnection("connecting");
        setError(null);
        const statusListener = (raw: unknown) => {
          if (disposed) return;
          const status =
            raw && typeof raw === "object" ? (raw as { status?: unknown }).status : undefined;
          const next = relayConnectionState(status);
          lease = relayLeaseFromStatus(raw);
          leaseRef.current = lease;
          setConnection(next);
          setLease(lease);
          if (next === "failed") setError("Worker 连接已失效，请重新连接");
        };
        const acpListener = (messageType: number, payload: unknown) => {
          if (disposed) return;
          if (messageType === 0x0b) promptConfirmationRef.current.consume(payload);
          applyMobileAcpRelayMessage(manager, podKey, messageType, payload);
          setSession(readMobileAcpSession(manager, podKey));
        };
        await Promise.all([
          nextRelay.set_status_listener(podKey, listenerId, statusListener),
          nextRelay.set_acp_listener(podKey, listenerId, acpListener),
        ]);
        if (disposed) {
          nextRelay.remove_status_listener(podKey, listenerId);
          nextRelay.remove_acp_listener(podKey, listenerId);
          return;
        }
        await nextRelay.subscribe(
          podKey,
          `mobile-acp-${podKey}-${Math.random().toString(36).slice(2, 10)}`,
          info.relayUrl,
          info.token,
          () => {},
        );
        if (disposed) void resetMobileRelayConnection(nextRelay, podKey);
      } catch (cause) {
        if (!disposed) {
          promptConfirmationRef.current.rejectAll("Worker 连接失败");
          setConnection("failed");
          setError(cause instanceof Error ? cause.message : "Worker 连接失败");
        }
      }
    };
    void connect();

    return () => {
      disposed = true;
      promptConfirmationRef.current.rejectAll("Worker 连接已关闭");
      setRelayControl(null);
      if (relay) {
        relay.remove_status_listener(podKey, listenerId);
        relay.remove_acp_listener(podKey, listenerId);
      }
      if (relay && lease.status === "granted" && lease.leaseId) {
        void relay.release_control(podKey, lease.leaseId);
      }
      if (relay) void resetMobileRelayConnection(relay, podKey);
    };
  }, [attempt, podKey]);

  const send = useCallback(
    async (command: Record<string, unknown>) => {
      const relay = relayRef.current;
      if (!relay || connection !== "connected") throw new Error("Worker 尚未连接");
      if (leaseRef.current.status !== "granted") throw new Error("请先接管输入");
      await relay.send_acp_command(podKey, JSON.stringify(command));
    },
    [connection, podKey],
  );

  const sendPrompt = useCallback(
    async (prompt: string) => {
      const requestId = `mobile-prompt-${Date.now()}-${++nextPromptRequestId}`;
      const confirmed = promptConfirmationRef.current.waitFor(requestId);
      try {
        await send({ type: "prompt", prompt, requestId });
      } catch (cause) {
        promptConfirmationRef.current.consume({
          type: "commandFailed",
          requestId,
          message: cause instanceof Error ? cause.message : "消息发送失败",
        });
      }
      await confirmed;
    },
    [send],
  );

  const reconnect = useCallback(() => {
    setConnection("connecting");
    setError(null);
    setAttempt((value) => value + 1);
  }, []);

  return {
    connection,
    control,
    error,
    lease,
    reconnect,
    respondPermission: (requestId: string, approved: boolean) =>
      send({ type: "permission_response", requestId, approved }),
    sendPrompt,
    session,
    interrupt: () => send({ type: "interrupt" }),
  };
}
