import { describe, expect, it } from "vitest";
import { codexItemsToAcpSnapshot } from "@/lib/codexItemsToAcpSnapshot";

describe("codexItemsToAcpSnapshot", () => {
  it("converts messages and tool calls into an ACP snapshot", () => {
    const snapshot = codexItemsToAcpSnapshot("sess-1", [
      {
        id: "m1",
        type: "message",
        role: "user",
        content: [{ type: "input_text", text: "hello" }],
      },
      {
        id: "tc1",
        type: "function_call",
        call_id: "call-1",
        name: "read_file",
        arguments: '{"path":"main.ts"}',
      },
      {
        id: "out1",
        type: "function_call_output",
        call_id: "call-1",
        output: "file contents",
      },
      {
        id: "m2",
        type: "message",
        role: "assistant",
        content: [{ type: "output_text", text: "Done." }],
      },
    ]);

    expect(snapshot.sessionId).toBe("sess-1");
    expect(snapshot.messages).toEqual([
      { role: "user", text: "hello" },
      { role: "assistant", text: "Done." },
    ]);
    expect(snapshot.toolCalls).toHaveLength(1);
    expect(snapshot.toolCalls[0]).toMatchObject({
      toolCallId: "call-1",
      toolName: "read_file",
      success: true,
      resultText: "file contents",
    });
  });

  it("skips is_meta messages", () => {
    const snapshot = codexItemsToAcpSnapshot("sess-2", [
      {
        id: "meta",
        type: "message",
        role: "user",
        is_meta: true,
        content: [{ type: "input_text", text: "hidden" }],
      },
    ]);
    expect(snapshot.messages).toHaveLength(0);
  });
});
