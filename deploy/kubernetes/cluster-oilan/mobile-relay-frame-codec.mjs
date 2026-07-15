export const ACP_EVENT = 0x0b;
export const ACP_COMMAND = 0x0c;
export const ACP_SNAPSHOT = 0x0d;
export const CONTROL = 0x07;
export const TERMINAL_INPUT = 0x03;
export const TERMINAL_OUTPUT = 0x02;
export const TERMINAL_RESIZE = 0x04;
export const TERMINAL_SNAPSHOT = 0x01;

export function browserRelayUrl(relayUrl, token) {
  const url = new URL(relayUrl);
  url.pathname = `${url.pathname.replace(/\/$/, "")}/browser/relay`;
  url.searchParams.set("token", token);
  return url.toString();
}

export function jsonFrame(type, value) {
  return binaryFrame(type, new TextEncoder().encode(JSON.stringify(value)));
}

export function binaryFrame(type, payload) {
  return Uint8Array.from([type, ...payload]);
}

export function resizeFrame(cols, rows) {
  const payload = new Uint8Array(4);
  new DataView(payload.buffer).setUint16(0, cols);
  new DataView(payload.buffer).setUint16(2, rows);
  return binaryFrame(TERMINAL_RESIZE, payload);
}

export function parseJson(bytes) {
  if (bytes.length === 0) return null;
  try {
    return JSON.parse(new TextDecoder().decode(bytes));
  } catch {
    return null;
  }
}

export async function messageBytes(data) {
  if (data instanceof ArrayBuffer) return new Uint8Array(data);
  if (ArrayBuffer.isView(data)) return new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
  if (data && typeof data.arrayBuffer === "function") return new Uint8Array(await data.arrayBuffer());
  throw new Error("Relay returned an unsupported frame");
}
