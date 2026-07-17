import type { EmbedSessionClient } from "@/embed-session-api";
import {
  decodeEmbeddedAcpFrame,
  encodeAcpConfigurationCommand,
  encodeAcpPong,
  encodeAcpSnapshotRequest,
  type EmbeddedAcpConfiguration,
} from "./embeddedAcpRelayCodec";
import {
  buildRelayWebSocketUrl,
  decodeRelayFrame,
  encodeControlLeaseFrame,
} from "./relayFrameCodec";

const CONTROL_CLIENT_LABEL = "agent-embed";
const CONTROL_RENEWAL_LEAD_MS = 5_000;

export interface EmbeddedAcpConnectionState {
  connected?: boolean;
  configuration?: Partial<EmbeddedAcpConfiguration>;
}

export class EmbeddedAcpConfigurationConnection {
  private readonly client: EmbedSessionClient;
  private readonly onError: (error: unknown) => void;
  private readonly onState: (state: EmbeddedAcpConnectionState) => void;
  private lease: { id: string; expiresAt: number } | null = null;
  private renewalTimer: number | null = null;
  private socket: WebSocket | null = null;
  private opening = false;

  constructor(
    client: EmbedSessionClient,
    onState: (state: EmbeddedAcpConnectionState) => void,
    onError: (error: unknown) => void,
  ) {
    this.client = client;
    this.onState = onState;
    this.onError = onError;
  }

  connect(): void {
    if (!this.client.getAcpRelayConnection || this.socket || this.opening) return;
    this.opening = true;
    void this.open();
  }

  close(): void {
    this.opening = false;
    const socket = this.socket;
    if (socket?.readyState === WebSocket.OPEN && this.lease) {
      socket.send(
        encodeControlLeaseFrame({
          action: "release",
          leaseId: this.lease.id,
        }),
      );
    }
    this.clearLease();
    this.socket = null;
    socket?.close();
    this.onState({ connected: false });
  }

  update(patch: Record<string, unknown>): Promise<void> {
    const socket = this.socket;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return Promise.reject(new Error("Agent configuration connection is unavailable"));
    }
    if (!this.lease || this.lease.expiresAt <= Date.now()) {
      this.clearLease();
      this.onState({ connected: false });
      return Promise.reject(new Error("Agent configuration control is unavailable"));
    }
    socket.send(encodeAcpConfigurationCommand(patch));
    return Promise.resolve();
  }

  private async open(): Promise<void> {
    try {
      const relay = await this.client.getAcpRelayConnection!();
      if (!this.opening) return;
      const socket = new WebSocket(buildRelayWebSocketUrl(relay.relayUrl, relay.token));
      socket.binaryType = "arraybuffer";
      this.socket = socket;
      socket.onopen = () => {
        if (this.socket !== socket) return;
        this.opening = false;
        socket.send(
          encodeControlLeaseFrame({
            action: "acquire",
            clientLabel: CONTROL_CLIENT_LABEL,
          }),
        );
        socket.send(encodeAcpSnapshotRequest());
      };
      socket.onmessage = (event) => this.handleMessage(socket, event);
      socket.onclose = () => {
        if (this.socket !== socket) return;
        this.socket = null;
        this.opening = false;
        this.clearLease();
        this.onState({ connected: false });
      };
    } catch (cause) {
      this.opening = false;
      this.onState({ connected: false });
      this.onError(cause);
    }
  }

  private handleMessage(socket: WebSocket, event: MessageEvent): void {
    if (this.socket !== socket || !(event.data instanceof ArrayBuffer)) return;
    try {
      const relayFrame = decodeRelayFrame(new Uint8Array(event.data));
      if (relayFrame.kind === "control") {
        this.handleControlStatus(socket, relayFrame);
        return;
      }
      const frame = decodeEmbeddedAcpFrame(event.data);
      if (frame.kind === "ping") {
        socket.send(encodeAcpPong());
      } else if (frame.kind === "configuration") {
        this.onState({ configuration: frame.configuration });
      } else if (frame.kind === "configuration-error") {
        this.onError(new Error(frame.message));
      }
    } catch (cause) {
      this.onError(cause);
    }
  }

  private handleControlStatus(
    socket: WebSocket,
    frame: ReturnType<typeof decodeRelayFrame> & { kind: "control" },
  ): void {
    if (
      frame.status === "granted" &&
      frame.leaseId &&
      frame.expiresAt &&
      frame.expiresAt > Date.now()
    ) {
      this.lease = { id: frame.leaseId, expiresAt: frame.expiresAt };
      this.scheduleRenewal(socket);
      this.onState({ connected: true });
      return;
    }
    this.clearLease();
    this.onState({ connected: false });
    if (frame.status === "busy") {
      this.onError(new Error("Agent configuration is controlled by another client"));
    } else if (frame.status === "expired" || frame.status === "control_required") {
      socket.send(
        encodeControlLeaseFrame({
          action: "acquire",
          clientLabel: CONTROL_CLIENT_LABEL,
        }),
      );
    }
  }

  private scheduleRenewal(socket: WebSocket): void {
    if (!this.lease) return;
    if (this.renewalTimer !== null) window.clearTimeout(this.renewalTimer);
    const leaseId = this.lease.id;
    const delay = Math.max(
      1_000,
      this.lease.expiresAt - Date.now() - CONTROL_RENEWAL_LEAD_MS,
    );
    this.renewalTimer = window.setTimeout(() => {
      this.renewalTimer = null;
      if (
        this.socket !== socket ||
        socket.readyState !== WebSocket.OPEN ||
        this.lease?.id !== leaseId
      ) {
        return;
      }
      socket.send(encodeControlLeaseFrame({ action: "renew", leaseId }));
    }, delay);
  }

  private clearLease(): void {
    if (this.renewalTimer !== null) {
      window.clearTimeout(this.renewalTimer);
      this.renewalTimer = null;
    }
    this.lease = null;
  }
}
