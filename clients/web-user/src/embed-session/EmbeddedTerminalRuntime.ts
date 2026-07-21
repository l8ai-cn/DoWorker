import type {
  AgentConnectionStatus,
  TerminalControlLease,
  TerminalResource,
  TerminalRuntime,
} from "@agent-cloud/agent-ui";

import type { EmbedSessionClient } from "@/embed-session-api";
import {
  clearPendingControl,
  type EmbeddedTerminalConnection,
} from "./embeddedTerminalConnection";
import {
  requestTerminalControl,
  requireTerminalLease,
} from "./embeddedTerminalControl";
import { handleEmbeddedTerminalFrame } from "./embeddedTerminalFrameHandler";
import { encodeInputFrame, encodeResizeFrame } from "./relayFrameCodec";
import { TerminalListenerRegistry } from "./terminalListenerRegistry";
import { createEmbeddedTerminalSocket } from "./embeddedTerminalSocket";

export class EmbeddedTerminalRuntime implements TerminalRuntime {
  private readonly client: EmbedSessionClient;
  private readonly connections = new Map<string, EmbeddedTerminalConnection>();
  private readonly listeners = new TerminalListenerRegistry();

  constructor(client: EmbedSessionClient) {
    this.client = client;
  }

  async connect(resource: TerminalResource): Promise<void> {
    const existing = this.connections.get(resource.id);
    if (existing) return existing.ready;
    this.listeners.publishStatus(resource.id, "connecting");
    let resolveReady: () => void = () => {};
    let rejectReady: (error: Error) => void = () => {};
    const ready = new Promise<void>((resolve, reject) => {
      resolveReady = resolve;
      rejectReady = reject;
    });
    const connection: EmbeddedTerminalConnection = {
      lease: null,
      pending: null,
      ready,
      rejectReady,
      resolveReady,
      socket: null,
    };
    this.connections.set(resource.id, connection);
    void this.openSocket(resource.id, connection);
    return ready;
  }

  private async openSocket(
    resourceId: string,
    connection: EmbeddedTerminalConnection,
  ): Promise<void> {
    let socket: WebSocket;
    try {
      socket = await createEmbeddedTerminalSocket(
        this.client,
        (status) => this.listeners.publishStatus(resourceId, status),
      );
    } catch (cause) {
      this.closeConnection(
        resourceId,
        connection,
        cause instanceof Error ? cause : new Error(String(cause)),
      );
      return;
    }
    if (this.connections.get(resourceId) !== connection) {
      socket.close();
      return;
    }
    connection.socket = socket;
    socket.onopen = () => {
      if (this.connections.get(resourceId) !== connection) return;
      connection.resolveReady();
      this.listeners.publishStatus(resourceId, "connected");
    };
    socket.onmessage = (event) =>
      handleEmbeddedTerminalFrame(
        connection,
        event,
        (bytes) => this.listeners.publishOutput(resourceId, bytes),
      );
    socket.onclose = () => this.closeConnection(resourceId, connection);
  }

  disconnect(resourceId: string): void {
    const connection = this.connections.get(resourceId);
    if (!connection) return;
    connection.socket?.close();
    this.closeConnection(resourceId, connection);
  }

  subscribeOutput(
    resourceId: string,
    listener: (bytes: Uint8Array) => void,
  ): () => void {
    return this.listeners.subscribeOutput(resourceId, listener);
  }

  subscribeStatus(
    resourceId: string,
    listener: (status: AgentConnectionStatus) => void,
  ): () => void {
    return this.listeners.subscribeStatus(resourceId, listener);
  }

  async write(resourceId: string, bytes: Uint8Array): Promise<void> {
    this.requireOpenSocket(this.requireLease(resourceId)).send(encodeInputFrame(bytes));
  }

  async resize(resourceId: string, columns: number, rows: number): Promise<void> {
    this.requireOpenSocket(this.requireLease(resourceId)).send(
      encodeResizeFrame(columns, rows),
    );
  }

  acquireControl(
    resourceId: string,
    clientLabel: string,
  ): Promise<TerminalControlLease> {
    return this.requestControl(
      resourceId,
      "acquire",
      clientLabel,
    ) as Promise<TerminalControlLease>;
  }

  renewControl(resourceId: string, leaseId: string): Promise<void> {
    this.requireLease(resourceId, leaseId);
    return this.requestControl(resourceId, "renew", leaseId) as Promise<void>;
  }

  releaseControl(resourceId: string, leaseId: string): Promise<void> {
    this.requireLease(resourceId, leaseId);
    return this.requestControl(resourceId, "release", leaseId) as Promise<void>;
  }

  private requestControl(
    resourceId: string,
    action: "acquire" | "renew" | "release",
    value: string,
  ): Promise<TerminalControlLease | void> {
    const connection = this.requireConnected(resourceId);
    return requestTerminalControl(connection, action, value);
  }

  private requireConnected(resourceId: string): EmbeddedTerminalConnection {
    const connection = this.connections.get(resourceId);
    if (!connection?.socket || connection.socket.readyState !== WebSocket.OPEN) {
      throw new Error(`Terminal ${resourceId} is not connected`);
    }
    return connection;
  }

  private requireOpenSocket(connection: EmbeddedTerminalConnection): WebSocket {
    if (!connection.socket) throw new Error("Terminal socket is not available");
    return connection.socket;
  }

  private requireLease(
    resourceId: string,
    leaseId?: string,
  ): EmbeddedTerminalConnection {
    const connection = this.requireConnected(resourceId);
    requireTerminalLease(connection, resourceId, leaseId);
    return connection;
  }

  private closeConnection(
    resourceId: string,
    connection: EmbeddedTerminalConnection,
    error = new Error(`Relay connection for ${resourceId} closed`),
  ): void {
    if (this.connections.get(resourceId) !== connection) return;
    connection.rejectReady(error);
    clearPendingControl(connection)?.reject(error);
    connection.lease = null;
    this.connections.delete(resourceId);
    this.listeners.publishStatus(resourceId, "disconnected");
  }

}
