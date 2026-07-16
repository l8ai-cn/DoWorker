import { describe, expect, it } from "vitest";

import {
  RelayFrameType,
  decodeRelayFrame,
  encodeControlLeaseFrame,
  encodeInputFrame,
  encodePongFrame,
  encodeResizeFrame,
} from "./relayFrameCodec";

const encoder = new TextEncoder();

describe("relayFrameCodec", () => {
  it("encodes input and resize frames with the Relay binary prefix", () => {
    expect(encodeInputFrame(new Uint8Array([65, 66]))).toEqual(
      new Uint8Array([RelayFrameType.Input, 65, 66]),
    );
    expect(encodeResizeFrame(0x1234, 0xabcd)).toEqual(
      new Uint8Array([RelayFrameType.Resize, 0x12, 0x34, 0xab, 0xcd]),
    );
  });

  it("encodes pong and control lease requests", () => {
    expect(encodePongFrame()).toEqual(new Uint8Array([RelayFrameType.Pong]));

    const frame = encodeControlLeaseFrame({
      action: "acquire",
      clientLabel: "embedded iframe",
    });
    expect(frame[0]).toBe(RelayFrameType.Control);
    expect(JSON.parse(new TextDecoder().decode(frame.subarray(1)))).toEqual({
      type: "control_lease",
      action: "acquire",
      client_label: "embedded iframe",
    });
  });

  it("decodes output and snapshot serialized content", () => {
    const output = decodeRelayFrame(
      new Uint8Array([RelayFrameType.Output, ...encoder.encode("hello")]),
    );
    expect(output.kind).toBe("output");
    expect(Array.from("bytes" in output ? output.bytes : [])).toEqual(
      Array.from(encoder.encode("hello")),
    );

    const snapshot = encoder.encode(
      JSON.stringify({ serialized_content: "\u001b[31mready\u001b[0m" }),
    );
    const decoded = decodeRelayFrame(
      new Uint8Array([RelayFrameType.Snapshot, ...snapshot]),
    );
    expect(decoded.kind).toBe("snapshot");
    expect(Array.from("bytes" in decoded ? decoded.bytes : [])).toEqual(
      Array.from(encoder.encode("\u001b[31mready\u001b[0m")),
    );
  });

  it("decodes ping, pong, and control lease status frames", () => {
    expect(decodeRelayFrame(new Uint8Array([RelayFrameType.Ping]))).toEqual({
      kind: "ping",
    });
    expect(decodeRelayFrame(new Uint8Array([RelayFrameType.Pong]))).toEqual({
      kind: "pong",
    });

    const payload = encoder.encode(
      JSON.stringify({
        type: "control_lease",
        status: "granted",
        lease_id: "lease-1",
        expires_at: 1_234,
      }),
    );
    expect(
      decodeRelayFrame(
        new Uint8Array([RelayFrameType.Control, ...payload]),
      ),
    ).toEqual({
      kind: "control",
      status: "granted",
      leaseId: "lease-1",
      expiresAt: 1_234,
    });
  });

  it("rejects malformed frames and out-of-range resize dimensions", () => {
    expect(() => decodeRelayFrame(new Uint8Array())).toThrow(
      "Relay frame is empty",
    );
    expect(() =>
      decodeRelayFrame(
        new Uint8Array([
          RelayFrameType.Snapshot,
          ...encoder.encode(JSON.stringify({ lines: [] })),
        ]),
      ),
    ).toThrow("Relay snapshot is invalid");
    expect(() => encodeResizeFrame(65_536, 24)).toThrow(
      "Terminal columns must be an integer between 1 and 65535",
    );
  });
});
