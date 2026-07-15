const encoder = new TextEncoder();
const decoder = new TextDecoder();
const snapshotClear = encoder.encode("\u001b[2J\u001b[H\u001b[3J");

export const RelayFrameType = {
  Snapshot: 0x01,
  Output: 0x02,
  Input: 0x03,
  Resize: 0x04,
  Ping: 0x05,
  Pong: 0x06,
  Control: 0x07,
} as const;

export type ControlLeaseStatus =
  | "granted"
  | "busy"
  | "released"
  | "expired"
  | "control_required";

export type DecodedRelayFrame =
  | { kind: "snapshot"; bytes: Uint8Array }
  | { kind: "output"; bytes: Uint8Array }
  | { kind: "ping" }
  | { kind: "pong" }
  | {
      kind: "control";
      status: ControlLeaseStatus;
      leaseId?: string;
      expiresAt?: number;
    }
  | { kind: "other"; type: number };

interface ControlLeaseRequest {
  action: "acquire" | "renew" | "release";
  clientLabel?: string;
  leaseId?: string;
}

export function encodeInputFrame(bytes: Uint8Array): Uint8Array<ArrayBuffer> {
  return encodeFrame(RelayFrameType.Input, bytes);
}

export function encodeResizeFrame(columns: number, rows: number): Uint8Array<ArrayBuffer> {
  assertDimension("columns", columns);
  assertDimension("rows", rows);
  const frame = new Uint8Array(5);
  const view = new DataView(frame.buffer);
  frame[0] = RelayFrameType.Resize;
  view.setUint16(1, columns, false);
  view.setUint16(3, rows, false);
  return frame;
}

export function encodePongFrame(): Uint8Array<ArrayBuffer> {
  return new Uint8Array([RelayFrameType.Pong]);
}

export function encodeControlLeaseFrame(
  request: ControlLeaseRequest,
): Uint8Array<ArrayBuffer> {
  const payload = {
    type: "control_lease",
    action: request.action,
    ...(request.leaseId ? { lease_id: request.leaseId } : {}),
    ...(request.clientLabel ? { client_label: request.clientLabel } : {}),
  };
  return encodeFrame(RelayFrameType.Control, encoder.encode(JSON.stringify(payload)));
}

export function decodeRelayFrame(data: Uint8Array): DecodedRelayFrame {
  if (data.byteLength === 0) throw new Error("Relay frame is empty");
  const type = data[0];
  const payload = data.subarray(1);
  switch (type) {
    case RelayFrameType.Snapshot:
      return { kind: "snapshot", bytes: decodeSnapshot(payload) };
    case RelayFrameType.Output:
      return { kind: "output", bytes: payload.slice() };
    case RelayFrameType.Ping:
      return { kind: "ping" };
    case RelayFrameType.Pong:
      return { kind: "pong" };
    case RelayFrameType.Control:
      return decodeControl(payload) ?? { kind: "other", type };
    default:
      return { kind: "other", type };
  }
}

export function buildRelayWebSocketUrl(
  relayUrl: string,
  token: string,
): string {
  const url = new URL(relayUrl);
  if (url.protocol === "http:") url.protocol = "ws:";
  else if (url.protocol === "https:") url.protocol = "wss:";
  else if (url.protocol !== "ws:" && url.protocol !== "wss:") {
    throw new Error(`Unsupported Relay URL scheme: ${url.protocol}`);
  }
  url.pathname = `${url.pathname.replace(/\/+$/, "")}/browser/relay`;
  url.search = "";
  url.hash = "";
  url.searchParams.set("token", token);
  return url.toString();
}

export function snapshotOutput(bytes: Uint8Array): Uint8Array {
  const result = new Uint8Array(snapshotClear.byteLength + bytes.byteLength);
  result.set(snapshotClear);
  result.set(bytes, snapshotClear.byteLength);
  return result;
}

export function controlLeaseFromStatus(frame: {
  status: ControlLeaseStatus;
  leaseId?: string;
  expiresAt?: number;
}): { leaseId: string; expiresAt: number } | null {
  if (
    frame.status !== "granted" ||
    !frame.leaseId ||
    !frame.expiresAt ||
    frame.expiresAt <= Date.now()
  ) {
    return null;
  }
  return { leaseId: frame.leaseId, expiresAt: frame.expiresAt };
}

function encodeFrame(type: number, payload: Uint8Array): Uint8Array<ArrayBuffer> {
  const frame = new Uint8Array(payload.byteLength + 1);
  frame[0] = type;
  frame.set(payload, 1);
  return frame;
}

function decodeSnapshot(payload: Uint8Array): Uint8Array {
  const value = parseJson(payload) as { serialized_content?: unknown };
  if (typeof value.serialized_content !== "string") {
    throw new Error("Relay snapshot is invalid");
  }
  return encoder.encode(value.serialized_content);
}

function decodeControl(payload: Uint8Array): DecodedRelayFrame | null {
  const value = parseJson(payload) as {
    type?: unknown;
    status?: unknown;
    lease_id?: unknown;
    expires_at?: unknown;
  };
  if (value.type !== "control_lease" || !isControlStatus(value.status)) {
    return null;
  }
  if (value.lease_id !== undefined && typeof value.lease_id !== "string") {
    throw new Error("Relay control lease status is invalid");
  }
  if (value.expires_at !== undefined && typeof value.expires_at !== "number") {
    throw new Error("Relay control lease status is invalid");
  }
  return {
    kind: "control",
    status: value.status,
    ...(value.lease_id ? { leaseId: value.lease_id } : {}),
    ...(value.expires_at ? { expiresAt: value.expires_at } : {}),
  };
}

function parseJson(payload: Uint8Array): unknown {
  try {
    return JSON.parse(decoder.decode(payload));
  } catch {
    throw new Error("Relay JSON payload is invalid");
  }
}

function isControlStatus(value: unknown): value is ControlLeaseStatus {
  return (
    value === "granted" ||
    value === "busy" ||
    value === "released" ||
    value === "expired" ||
    value === "control_required"
  );
}

function assertDimension(name: "columns" | "rows", value: number): void {
  if (!Number.isInteger(value) || value < 1 || value > 65_535) {
    throw new Error(
      `Terminal ${name} must be an integer between 1 and 65535`,
    );
  }
}
