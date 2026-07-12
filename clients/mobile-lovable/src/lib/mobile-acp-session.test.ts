import { create, fromBinary } from "@bufbuild/protobuf";
import { AddPermissionRequestRequestSchema } from "@do-worker/proto/acp_state/v1/acp_state_pb";
import { describe, expect, it, vi } from "vitest";
import {
  applyMobileAcpRelayMessage,
  readMobileAcpSession,
  type MobileAcpManager,
} from "./mobile-acp-session";

function manager(): MobileAcpManager {
  return {
    add_content_chunk: vi.fn(),
    add_log: vi.fn(),
    add_permission_request: vi.fn(),
    clear_session: vi.fn(),
    get_session_json: vi.fn(),
    mark_last_message_complete: vi.fn(),
    update_session_state: vi.fn(),
  };
}

describe("mobile ACP session", () => {
  it("hydrates the authoritative Wasm cache from an ACP snapshot", () => {
    const acp = manager();

    applyMobileAcpRelayMessage(acp, "pod-1", 0x0d, {
      state: "waiting_permission",
      messages: [{ text: "hello", role: "assistant" }],
      pendingPermissions: [
        {
          requestId: "perm-1",
          toolName: "shell",
          argumentsJson: "{\"command\":\"ls\"}",
          description: "List files",
        },
      ],
    });

    expect(acp.clear_session).toHaveBeenCalledWith("pod-1");
    expect(acp.update_session_state).toHaveBeenCalledWith("pod-1", "waiting_permission");
    expect(acp.add_content_chunk).toHaveBeenCalledWith("pod-1", "hello", "assistant");
    const request = fromBinary(
      AddPermissionRequestRequestSchema,
      vi.mocked(acp.add_permission_request).mock.calls[0][0],
    );
    expect(request).toEqual(
      create(AddPermissionRequestRequestSchema, {
        podKey: "pod-1",
        requestJson:
          "{\"id\":\"perm-1\",\"tool_name\":\"shell\",\"args\":{\"command\":\"ls\"},\"description\":\"List files\"}",
      }),
    );
    expect(
      vi.mocked(acp.add_content_chunk).mock.invocationCallOrder[0],
    ).toBeLessThan(vi.mocked(acp.update_session_state).mock.invocationCallOrder[0]);
  });

  it("marks the streamed assistant message complete when the Worker becomes idle", () => {
    const acp = manager();

    applyMobileAcpRelayMessage(acp, "pod-1", 0x0b, {
      type: "contentChunk",
      text: "answer",
      role: "assistant",
    });
    applyMobileAcpRelayMessage(acp, "pod-1", 0x0b, { type: "sessionState", state: "idle" });

    expect(acp.add_content_chunk).toHaveBeenCalledWith("pod-1", "answer", "assistant");
    expect(acp.mark_last_message_complete).toHaveBeenCalledWith("pod-1");
  });

  it("reads only renderable ACP state from the Wasm cache", () => {
    const acp = manager();
    vi.mocked(acp.get_session_json).mockReturnValue(
      JSON.stringify({
        state: { processing: {} },
        messages: [{ text: "hello", role: "assistant", complete: true }],
        pending_permissions: [
          { id: "perm-1", tool_name: "shell", args: { command: "ls" }, description: "List" },
        ],
      }),
    );

    expect(readMobileAcpSession(acp, "pod-1")).toEqual({
      state: "processing",
      messages: [{ text: "hello", role: "assistant", complete: true }],
      pendingPermissions: [
        {
          requestId: "perm-1",
          toolName: "shell",
          argumentsJson: "{\"command\":\"ls\"}",
          description: "List",
        },
      ],
    });
  });
});
