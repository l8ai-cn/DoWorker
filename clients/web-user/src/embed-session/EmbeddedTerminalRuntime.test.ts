import type { TerminalResource } from "@agent-cloud/agent-ui";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedTerminalRuntime } from "./EmbeddedTerminalRuntime";
import { RelayFrameType, decodeRelayFrame, encodeControlLeaseFrame } from "./relayFrameCodec";

const encoder = new TextEncoder();
const resource: TerminalResource = {
  id: "terminal_tui_main",
  label: "main:tui",
  status: "connected",
  writable: true,
};

class FakeWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 3;
  static instances: FakeWebSocket[] = [];

  binaryType = "blob";
  onclose: ((event: CloseEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onopen: ((event: Event) => void) | null = null;
  readyState = FakeWebSocket.CONNECTING;
  sent: Uint8Array[] = [];
  readonly url: string;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  send(data: ArrayBufferLike | Blob | ArrayBufferView): void {
    if (data instanceof Uint8Array) {
      this.sent.push(data);
      return;
    }
    if (ArrayBuffer.isView(data)) {
      this.sent.push(new Uint8Array(data.buffer, data.byteOffset, data.byteLength));
      return;
    }
    throw new Error("FakeWebSocket only accepts binary frames");
  }

  close(): void {
    if (this.readyState === FakeWebSocket.CLOSED) return;
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.({ code: 1000, reason: "" } as CloseEvent);
  }

  open(): void {
    this.readyState = FakeWebSocket.OPEN;
    this.onopen?.(new Event("open"));
  }

  emit(frame: Uint8Array): void {
    const data = frame.slice().buffer as ArrayBuffer;
    this.onmessage?.({ data } as MessageEvent<ArrayBuffer>);
  }

  drop(): void {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.({ code: 1006, reason: "" } as CloseEvent);
  }
}

function latestSocket(): FakeWebSocket {
  const socket = FakeWebSocket.instances.at(-1);
  if (!socket) throw new Error("No WebSocket was created");
  return socket;
}

function createClient(
  relayUrl = "https://relay.example.test/relay/",
  token = "token + value",
): EmbedSessionClient {
  return {
    getSession: vi.fn(),
    getTerminals: vi.fn(),
    loadDownload: vi.fn(),
    loadResource: vi.fn(),
    uploadAttachment: vi.fn(),
    getRelayConnection: vi.fn(async () => ({
      relayUrl,
      token,
      podKey: "pod-1",
    })),
  };
}

async function connect(runtime: EmbeddedTerminalRuntime): Promise<FakeWebSocket> {
  const pending = runtime.connect(resource);
  await vi.waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));
  const socket = latestSocket();
  socket.open();
  await pending;
  return socket;
}

describe("EmbeddedTerminalRuntime", () => {
  beforeEach(() => {
    FakeWebSocket.instances = [];
    vi.stubGlobal("WebSocket", FakeWebSocket as unknown as typeof WebSocket);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("normalizes the Relay URL and publishes connection and terminal output", async () => {
    const runtime = new EmbeddedTerminalRuntime(createClient());
    const statuses: string[] = [];
    const output = vi.fn();
    runtime.subscribeStatus(resource.id, (status) => statuses.push(status));
    runtime.subscribeOutput(resource.id, output);

    const pending = runtime.connect(resource);
    await vi.waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));
    const socket = latestSocket();
    expect(socket.url).toBe("wss://relay.example.test/relay/browser/relay?token=token+%2B+value");
    expect(socket.binaryType).toBe("arraybuffer");
    expect(statuses).toEqual(["connecting"]);

    socket.open();
    await pending;
    expect(statuses).toEqual(["connecting", "connected"]);

    socket.emit(new Uint8Array([RelayFrameType.Output, ...encoder.encode("output")]));
    socket.emit(
      new Uint8Array([
        RelayFrameType.Snapshot,
        ...encoder.encode(JSON.stringify({ serialized_content: "snapshot" })),
      ]),
    );
    expect(Array.from(output.mock.calls[0][0] as Uint8Array)).toEqual(
      Array.from(encoder.encode("output")),
    );
    expect(Array.from(output.mock.calls[1][0] as Uint8Array)).toEqual(
      Array.from(encoder.encode("\u001b[2J\u001b[H\u001b[3Jsnapshot")),
    );

    socket.emit(new Uint8Array([RelayFrameType.Ping]));
    expect(socket.sent.at(-1)).toEqual(new Uint8Array([RelayFrameType.Pong]));
  });

  it("waits for granted or busy instead of fabricating control", async () => {
    const runtime = new EmbeddedTerminalRuntime(createClient("ws://relay.test"));
    const socket = await connect(runtime);

    let settled = false;
    const pending = runtime.acquireControl(resource.id, "iframe");
    void pending.then(
      () => {
        settled = true;
      },
      () => {
        settled = true;
      },
    );
    await Promise.resolve();
    expect(settled).toBe(false);
    expect(decodeRelayFrame(socket.sent.at(-1)!)).toEqual({
      kind: "other",
      type: RelayFrameType.Control,
    });
    expect(socket.sent.at(-1)).toEqual(
      encodeControlLeaseFrame({ action: "acquire", clientLabel: "iframe" }),
    );

    socket.emit(controlStatus("granted", "lease-1", Date.now() + 60_000));
    await expect(pending).resolves.toEqual({
      leaseId: "lease-1",
      expiresAt: expect.any(Number),
    });
  });

  it("rejects busy acquisition and gates writes and resizes on a valid lease", async () => {
    const runtime = new EmbeddedTerminalRuntime(createClient());
    await expect(runtime.write(resource.id, encoder.encode("blocked"))).rejects.toThrow(
      "Terminal terminal_tui_main is not connected",
    );

    const socket = await connect(runtime);
    await expect(runtime.resize(resource.id, 120, 36)).rejects.toThrow(
      "Terminal terminal_tui_main does not hold a valid control lease",
    );

    const busy = runtime.acquireControl(resource.id, "iframe");
    socket.emit(controlStatus("busy"));
    await expect(busy).rejects.toThrow("Terminal control is busy");

    const granted = runtime.acquireControl(resource.id, "iframe");
    socket.emit(controlStatus("granted", "lease-2", Date.now() + 60_000));
    await granted;
    await runtime.write(resource.id, encoder.encode("ls\r"));
    await runtime.resize(resource.id, 0x1234, 0xabcd);
    expect(socket.sent.at(-2)).toEqual(
      new Uint8Array([RelayFrameType.Input, ...encoder.encode("ls\r")]),
    );
    expect(socket.sent.at(-1)).toEqual(
      new Uint8Array([RelayFrameType.Resize, 0x12, 0x34, 0xab, 0xcd]),
    );

    socket.emit(controlStatus("expired"));
    await expect(runtime.write(resource.id, encoder.encode("blocked"))).rejects.toThrow(
      "Terminal terminal_tui_main does not hold a valid control lease",
    );
  });

  it("waits for renew and release status and rejects pending work on close", async () => {
    const runtime = new EmbeddedTerminalRuntime(createClient());
    const statuses: string[] = [];
    runtime.subscribeStatus(resource.id, (status) => statuses.push(status));
    const socket = await connect(runtime);

    const acquire = runtime.acquireControl(resource.id, "iframe");
    socket.emit(controlStatus("granted", "lease-3", Date.now() + 60_000));
    await acquire;

    const renew = runtime.renewControl(resource.id, "lease-3");
    socket.emit(controlStatus("granted", "lease-3", Date.now() + 120_000));
    await expect(renew).resolves.toBeUndefined();

    const release = runtime.releaseControl(resource.id, "lease-3");
    socket.emit(controlStatus("released"));
    await expect(release).resolves.toBeUndefined();

    const waiting = runtime.acquireControl(resource.id, "iframe");
    socket.drop();
    await expect(waiting).rejects.toThrow("Relay connection for terminal_tui_main closed");
    expect(statuses.at(-1)).toBe("disconnected");
  });

  it("rejects connect when the socket closes before opening", async () => {
    const runtime = new EmbeddedTerminalRuntime(createClient());
    const pending = runtime.connect(resource);
    await vi.waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));
    latestSocket().drop();
    await expect(pending).rejects.toThrow("Relay connection for terminal_tui_main closed");
  });

  it("deduplicates credential loading and cancels a pending connect", async () => {
    let resolveRelay = (_value: { relayUrl: string; token: string; podKey: string }) => {};
    const relayConnection = new Promise<{
      relayUrl: string;
      token: string;
      podKey: string;
    }>((resolve) => {
      resolveRelay = resolve;
    });
    const client = createClient();
    vi.mocked(client.getRelayConnection!).mockReturnValue(relayConnection);
    const runtime = new EmbeddedTerminalRuntime(client);
    const first = runtime.connect(resource);
    const second = runtime.connect(resource);

    expect(client.getRelayConnection).toHaveBeenCalledOnce();
    runtime.disconnect(resource.id);
    resolveRelay({
      relayUrl: "ws://relay.example.test/relay",
      token: "token",
      podKey: "pod-1",
    });

    const results = await Promise.allSettled([first, second]);
    expect(results).toEqual([
      expect.objectContaining({
        status: "rejected",
        reason: expect.objectContaining({
          message: "Relay connection for terminal_tui_main closed",
        }),
      }),
      expect.objectContaining({
        status: "rejected",
        reason: expect.objectContaining({
          message: "Relay connection for terminal_tui_main closed",
        }),
      }),
    ]);
    await vi.waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));
    expect(latestSocket().readyState).toBe(FakeWebSocket.CLOSED);
  });

  it("rejects a control request when Relay never answers", async () => {
    const runtime = new EmbeddedTerminalRuntime(createClient());
    await connect(runtime);
    vi.useFakeTimers();

    const pending = runtime.acquireControl(resource.id, "iframe");
    const rejection = expect(pending).rejects.toThrow("Terminal control request timed out");
    await vi.advanceTimersByTimeAsync(5_000);

    await rejection;
  });
});

function controlStatus(
  status: "granted" | "busy" | "released" | "expired" | "control_required",
  leaseId?: string,
  expiresAt?: number,
): Uint8Array {
  const payload = {
    type: "control_lease",
    status,
    ...(leaseId ? { lease_id: leaseId } : {}),
    ...(expiresAt ? { expires_at: expiresAt } : {}),
  };
  return new Uint8Array([RelayFrameType.Control, ...encoder.encode(JSON.stringify(payload))]);
}
