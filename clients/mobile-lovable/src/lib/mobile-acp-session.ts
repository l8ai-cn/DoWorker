import { create, toBinary } from "@bufbuild/protobuf";
import { AddPermissionRequestRequestSchema } from "@agent-cloud/proto/acp_state/v1/acp_state_pb";

export type MobileAcpManager = {
  add_content_chunk(podKey: string, text: string, role: string): void;
  add_log(podKey: string, level: string, message: string): void;
  add_permission_request(request: Uint8Array): void;
  clear_session(podKey: string): void;
  get_session_json(podKey: string): unknown;
  mark_last_message_complete(podKey: string): void;
  update_session_state(podKey: string, state: string): void;
};

export type MobileAcpSession = {
  state: string;
  messages: Array<{ text: string; role: string; complete?: boolean }>;
  pendingPermissions: Array<{
    requestId: string;
    toolName: string;
    argumentsJson: string;
    description: string;
  }>;
};

const emptySession: MobileAcpSession = {
  state: "idle",
  messages: [],
  pendingPermissions: [],
};

export function applyMobileAcpRelayMessage(
  manager: MobileAcpManager,
  podKey: string,
  messageType: number,
  payload: unknown,
): void {
  if (!payload || typeof payload !== "object") return;
  const data = payload as Record<string, unknown>;
  if (messageType === 0x0d) {
    applySnapshot(manager, podKey, data);
    return;
  }
  if (messageType !== 0x0b || typeof data.type !== "string") return;
  if (data.type === "contentChunk") {
    manager.add_content_chunk(podKey, text(data.text), text(data.role));
    return;
  }
  if (data.type === "permissionRequest") {
    addPermission(manager, podKey, data);
    return;
  }
  if (data.type === "sessionState") {
    const state = text(data.state);
    manager.update_session_state(podKey, state);
    if (state === "idle") manager.mark_last_message_complete(podKey);
    return;
  }
  if (data.type === "log") manager.add_log(podKey, text(data.level), text(data.message));
}

export function readMobileAcpSession(
  manager: MobileAcpManager,
  podKey: string,
): MobileAcpSession {
  const raw = manager.get_session_json(podKey);
  const data = parseObject(raw);
  if (!data) return emptySession;
  return {
    state: stateName(data.state),
    messages: messages(data.messages),
    pendingPermissions: permissions(data.pending_permissions),
  };
}

function applySnapshot(manager: MobileAcpManager, podKey: string, data: Record<string, unknown>) {
  manager.clear_session(podKey);
  for (const message of array(data.messages)) {
    const item = object(message);
    if (item) manager.add_content_chunk(podKey, text(item.text), text(item.role));
  }
  for (const permission of array(data.pendingPermissions)) {
    const item = object(permission);
    if (item) addPermission(manager, podKey, item);
  }
  if (data.state) manager.update_session_state(podKey, text(data.state));
}

function addPermission(
  manager: MobileAcpManager,
  podKey: string,
  data: Record<string, unknown>,
) {
  const requestJson = JSON.stringify({
    id: text(data.requestId),
    tool_name: text(data.toolName),
    args: parseArguments(data.argumentsJson),
    description: text(data.description),
  });
  const request = create(AddPermissionRequestRequestSchema, { podKey, requestJson });
  manager.add_permission_request(toBinary(AddPermissionRequestRequestSchema, request));
}

function parseObject(raw: unknown): Record<string, unknown> | null {
  if (typeof raw === "string") {
    try {
      return object(JSON.parse(raw));
    } catch {
      return null;
    }
  }
  return object(raw);
}

function messages(raw: unknown): MobileAcpSession["messages"] {
  return array(raw).flatMap((value) => {
    const item = object(value);
    return item ? [{ text: text(item.text), role: text(item.role), complete: Boolean(item.complete) }] : [];
  });
}

function permissions(raw: unknown): MobileAcpSession["pendingPermissions"] {
  return array(raw).flatMap((value) => {
    const item = object(value);
    if (!item) return [];
    return [{
      requestId: text(item.id),
      toolName: text(item.tool_name),
      argumentsJson: JSON.stringify(item.args ?? null),
      description: text(item.description),
    }];
  });
}

function stateName(value: unknown): string {
  if (typeof value === "string") return value;
  const item = object(value);
  return item ? Object.keys(item)[0] ?? "idle" : "idle";
}

function parseArguments(value: unknown): unknown {
  if (typeof value !== "string") return null;
  try {
    return JSON.parse(value);
  } catch {
    return null;
  }
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function array(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function object(value: unknown): Record<string, unknown> | null {
  return value && typeof value === "object" ? value as Record<string, unknown> : null;
}
