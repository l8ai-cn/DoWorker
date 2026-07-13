import assert from "node:assert/strict";
import test from "node:test";
import { runMobileRelayDataPlaneSmoke } from "./mobile-relay-data-plane-smoke.mjs";

class FakeWebSocket {
  static instances = [];

  constructor(url) {
    this.url = url;
    this.binaryType = "";
    this.sent = [];
    this.controlFrames = 0;
    FakeWebSocket.instances.push(this);
    queueMicrotask(() => {
      this.onopen?.();
      this.onmessage?.({ data: frame(this.mode === "acp" ? 0x0d : 0x01, {}) });
    });
  }

  get mode() {
    return this.url.includes("mode=acp") ? "acp" : "pty";
  }

  close() {
    this.onclose?.();
  }

  send(data) {
    const bytes = new Uint8Array(data);
    this.sent.push(bytes);
    if (bytes[0] === 0x07) {
      this.controlFrames += 1;
      queueMicrotask(() => this.onmessage?.({
        data: frame(0x07, {
          expires_at: Date.now() + 30_000,
          lease_id: "lease-1",
          status: "granted",
          type: "control_lease",
        }),
      }));
      if (this.controlFrames === 2) {
        queueMicrotask(() => this.onmessage?.({
          data: frame(0x0b, { role: "assistant", text: "marker-", type: "contentChunk" }),
        }));
        queueMicrotask(() => this.onmessage?.({
          data: frame(0x0b, { role: "assistant", text: "acp", type: "contentChunk" }),
        }));
      }
    }
    if (bytes[0] === 0x0c) {
      this.command = JSON.parse(new TextDecoder().decode(bytes.slice(1)));
    }
  }
}

test("ACP smoke renews control until the assistant answers", async () => {
  FakeWebSocket.instances = [];
  const diagnostics = [];
  let timerScheduled = false;
  await runMobileRelayDataPlaneSmoke({
    debug: (message) => diagnostics.push(message),
    marker: "marker-acp",
    mode: "acp",
    relayUrl: "wss://relay.example/relay?mode=acp",
    token: "secret",
    WebSocketImpl: FakeWebSocket,
    schedule: (callback) => {
      if (!timerScheduled) {
        timerScheduled = true;
        queueMicrotask(callback);
      }
      return 1;
    },
    cancelSchedule: () => {},
  });

  const [socket] = FakeWebSocket.instances;
  assert.equal(socket.sent[0][0], 0x07);
  assert.equal(socket.sent[1][0], 0x0c);
  assert.equal(socket.sent[2][0], 0x07);
  assert.equal(socket.controlFrames, 2);
  assert.equal(socket.command.prompt, "Reply with exactly marker-acp.");
  assert.match(socket.url, /\/relay\/browser\/relay\?mode=acp&token=secret/);
  assert.deepEqual(diagnostics, [
    "received ACP snapshot",
    "requested ACP control lease",
    "ACP control lease status=granted",
    "ACP control lease granted",
    "sent ACP prompt",
    "renewed ACP control lease",
    "ACP control lease status=granted",
    "ACP control lease granted",
    "received matching ACP assistant chunk",
  ]);
  assert.equal(diagnostics.join(" ").includes("secret"), false);
});

test("PTY smoke requires a terminal snapshot and granted control", async () => {
  FakeWebSocket.instances = [];
  await runMobileRelayDataPlaneSmoke({
    mode: "pty",
    relayUrl: "wss://relay.example/relay",
    token: "secret",
    WebSocketImpl: FakeWebSocket,
  });

  const [socket] = FakeWebSocket.instances;
  assert.equal(socket.sent.length, 1);
  assert.equal(socket.sent[0][0], 0x07);
});

function frame(type, payload) {
  return Uint8Array.from([type, ...new TextEncoder().encode(JSON.stringify(payload))]);
}
