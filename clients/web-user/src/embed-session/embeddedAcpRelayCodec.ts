const encoder = new TextEncoder();
const decoder = new TextDecoder();

export const EmbeddedAcpFrameType = {
  Ping: 0x05,
  Pong: 0x06,
  SnapshotRequest: 0x0a,
  Event: 0x0b,
  Command: 0x0c,
  Snapshot: 0x0d,
} as const;

export interface EmbeddedAcpConfiguration {
  permissionMode: string;
  model: string;
  supportedPermissionModes: string[];
}

export type EmbeddedAcpRelayFrame =
  | { kind: "ping" }
  | { kind: "configuration"; configuration: Partial<EmbeddedAcpConfiguration> }
  | { kind: "configuration-error"; message: string }
  | { kind: "other" };

export function decodeEmbeddedAcpFrame(data: ArrayBuffer): EmbeddedAcpRelayFrame {
  const frame = new Uint8Array(data);
  if (frame.byteLength === 0) throw new Error("ACP Relay frame is empty");
  const type = frame[0];
  if (type === EmbeddedAcpFrameType.Ping) return { kind: "ping" };
  if (type === EmbeddedAcpFrameType.Snapshot) {
    const body = parseObject(frame.subarray(1));
    return {
      kind: "configuration",
      configuration: parseConfiguration(body.configuration),
    };
  }
  if (type !== EmbeddedAcpFrameType.Event) return { kind: "other" };
  const body = parseObject(frame.subarray(1));
  if (body.type === "configChanged") {
    return {
      kind: "configuration",
      configuration: parseConfiguration(body),
    };
  }
  if (body.type === "configChangeFailed") {
    return {
      kind: "configuration-error",
      message:
        typeof body.message === "string"
          ? body.message
          : "Agent configuration update failed",
    };
  }
  return { kind: "other" };
}

export function encodeAcpSnapshotRequest(): Uint8Array<ArrayBuffer> {
  return new Uint8Array([EmbeddedAcpFrameType.SnapshotRequest]);
}

export function encodeAcpPong(): Uint8Array<ArrayBuffer> {
  return new Uint8Array([EmbeddedAcpFrameType.Pong]);
}

export function encodeAcpConfigurationCommand(
  patch: Record<string, unknown>,
): Uint8Array<ArrayBuffer> {
  if (typeof patch.permissionMode === "string") {
    return encodeCommand({
      type: "set_permission_mode",
      mode: patch.permissionMode,
    });
  }
  if (typeof patch.model === "string") {
    return encodeCommand({ type: "set_model", model: patch.model });
  }
  throw new Error("Agent configuration patch is unsupported");
}

function encodeCommand(command: Record<string, unknown>): Uint8Array<ArrayBuffer> {
  const payload = encoder.encode(JSON.stringify(command));
  const frame = new Uint8Array(payload.byteLength + 1);
  frame[0] = EmbeddedAcpFrameType.Command;
  frame.set(payload, 1);
  return frame;
}

function parseObject(payload: Uint8Array): Record<string, unknown> {
  try {
    const value = JSON.parse(decoder.decode(payload));
    if (value && typeof value === "object" && !Array.isArray(value)) {
      return value as Record<string, unknown>;
    }
  } catch {
    throw new Error("ACP Relay payload is invalid JSON");
  }
  throw new Error("ACP Relay payload is invalid");
}

function parseConfiguration(value: unknown): Partial<EmbeddedAcpConfiguration> {
  if (!value || typeof value !== "object" || Array.isArray(value)) return {};
  const body = value as Record<string, unknown>;
  const configuration: Partial<EmbeddedAcpConfiguration> = {};
  if (typeof body.permissionMode === "string") {
    configuration.permissionMode = body.permissionMode;
  }
  if (typeof body.model === "string") configuration.model = body.model;
  if (Array.isArray(body.supportedPermissionModes)) {
    if (!body.supportedPermissionModes.every((mode) => typeof mode === "string")) {
      throw new Error("ACP permission modes are invalid");
    }
    configuration.supportedPermissionModes = body.supportedPermissionModes;
  }
  return configuration;
}
