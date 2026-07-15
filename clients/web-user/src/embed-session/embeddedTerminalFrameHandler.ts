import type { EmbeddedTerminalConnection } from "./embeddedTerminalConnection";
import { applyTerminalControlStatus } from "./embeddedTerminalControl";
import {
  decodeRelayFrame,
  encodePongFrame,
  snapshotOutput,
} from "./relayFrameCodec";

export function handleEmbeddedTerminalFrame(
  connection: EmbeddedTerminalConnection,
  event: MessageEvent,
  publishOutput: (bytes: Uint8Array) => void,
): void {
  const socket = connection.socket;
  if (!socket) throw new Error("Terminal socket is not available");
  if (!(event.data instanceof ArrayBuffer)) {
    socket.close(1002, "Relay sent a non-binary frame");
    return;
  }
  try {
    const frame = decodeRelayFrame(new Uint8Array(event.data));
    if (frame.kind === "output") publishOutput(frame.bytes);
    if (frame.kind === "snapshot") publishOutput(snapshotOutput(frame.bytes));
    if (frame.kind === "ping") socket.send(encodePongFrame());
    if (frame.kind === "control") applyTerminalControlStatus(connection, frame);
  } catch {
    socket.close(1002, "Relay sent an invalid frame");
  }
}
