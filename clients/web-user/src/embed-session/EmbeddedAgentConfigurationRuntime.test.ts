import { expect, it, vi } from "vitest";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedAgentSessionRuntime } from "./EmbeddedAgentSessionRuntime";
import {
  encodeControlLeaseFrame,
  RelayFrameType,
} from "./relayFrameCodec";

class FakeAcpWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 3;
  static instances: FakeAcpWebSocket[] = [];

  binaryType = "";
  onclose: (() => void) | null = null;
  onmessage: ((event: MessageEvent<ArrayBuffer>) => void) | null = null;
  onopen: (() => void) | null = null;
  readyState = FakeAcpWebSocket.CONNECTING;
  sent: Uint8Array[] = [];
  readonly url: string;

  constructor(url: string) {
    this.url = url;
    FakeAcpWebSocket.instances.push(this);
  }

  close() {
    this.readyState = FakeAcpWebSocket.CLOSED;
    this.onclose?.();
  }

  emit(type: number, payload: unknown) {
    const body = new TextEncoder().encode(JSON.stringify(payload));
    const frame = new Uint8Array(body.byteLength + 1);
    frame[0] = type;
    frame.set(body, 1);
    this.onmessage?.({ data: frame.buffer } as MessageEvent<ArrayBuffer>);
  }

  open() {
    this.readyState = FakeAcpWebSocket.OPEN;
    this.onopen?.();
  }

  send(data: ArrayBufferView) {
    this.sent.push(new Uint8Array(data.buffer, data.byteOffset, data.byteLength).slice());
  }
}

function openStream(signal: AbortSignal): Promise<Response> {
  return Promise.resolve(
    new Response(
      new ReadableStream<Uint8Array>({
        start(controller) {
          signal.addEventListener("abort", () => controller.close(), { once: true });
        },
      }),
      { status: 200 },
    ),
  );
}

function createAcpClient(): EmbedSessionClient {
  return {
    getAcpRelayConnection: vi.fn().mockResolvedValue({
      relayUrl: "ws://relay.example/relay",
      token: "relay-token",
      podKey: "pod-1",
    }),
    getItems: vi.fn().mockResolvedValue({ hasMore: false, items: [] }),
    getSession: vi.fn().mockResolvedValue({
      agentLabel: "codex-cli",
      id: "session-1",
      interactionMode: "acp",
      podKey: "pod-1",
      status: "idle",
      title: "Auth review",
    }),
    openStream: vi.fn(openStream),
  };
}

it("hydrates live ACP configuration and sends selection changes over Relay", async () => {
  FakeAcpWebSocket.instances = [];
  vi.stubGlobal("WebSocket", FakeAcpWebSocket as unknown as typeof WebSocket);
  const runtime = new EmbeddedAgentSessionRuntime(createAcpClient());

  await runtime.open("session-1");
  await vi.waitFor(() => expect(FakeAcpWebSocket.instances).toHaveLength(1));
  const socket = FakeAcpWebSocket.instances[0]!;
  socket.open();

  expect(socket.sent[0]).toEqual(
    encodeControlLeaseFrame({
      action: "acquire",
      clientLabel: "agent-embed",
    }),
  );
  expect(socket.sent[1]).toEqual(new Uint8Array([0x0a]));
  socket.emit(0x0d, {
    sessionId: "external-session",
    state: "idle",
    configuration: {
      permissionMode: "ask_dangerous",
      model: "gpt-5.3-codex",
      supportedPermissionModes: ["bypass", "ask_dangerous", "ask_any_write"],
    },
  });
  expect(runtime.getSnapshot("session-1").capabilities.updateConfiguration).toBe(false);

  socket.emit(RelayFrameType.Control, {
    type: "control_lease",
    status: "granted",
    lease_id: "lease-1",
    expires_at: Date.now() + 60_000,
  });

  await vi.waitFor(() =>
    expect(runtime.getSnapshot("session-1")).toMatchObject({
      capabilities: { updateConfiguration: true },
      configuration: [
        {
          id: "permissionMode",
          value: "ask_dangerous",
          options: [
            { value: "bypass" },
            { value: "ask_dangerous" },
            { value: "ask_any_write" },
          ],
        },
      ],
    }),
  );

  await runtime.updateConfiguration("session-1", "config-1", {
    permissionMode: "bypass",
  });
  expect(socket.sent.at(-1)?.[0]).toBe(0x0c);
  expect(new TextDecoder().decode(socket.sent.at(-1)?.subarray(1))).toBe(
    JSON.stringify({ type: "set_permission_mode", mode: "bypass" }),
  );

  socket.emit(0x0b, {
    type: "configChanged",
    permissionMode: "bypass",
  });
  await vi.waitFor(() =>
    expect(runtime.getSnapshot("session-1").configuration?.[0]?.value).toBe("bypass"),
  );
  runtime.close("session-1");
  expect(socket.sent.at(-1)).toEqual(
    encodeControlLeaseFrame({
      action: "release",
      leaseId: "lease-1",
    }),
  );
  vi.unstubAllGlobals();
});
