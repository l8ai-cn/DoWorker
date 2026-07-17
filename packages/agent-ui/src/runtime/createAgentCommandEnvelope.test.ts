import { describe, expect, it } from "vitest";

import { createAgentCommandEnvelope } from "./createAgentCommandEnvelope";

describe("createAgentCommandEnvelope", () => {
  it("matches the Go deterministic protobuf digest", async () => {
    const command = await createAgentCommandEnvelope({
      sessionId: "conv-1",
      streamEpoch: "epoch-1",
      commandId: "command-1",
      expectedRevision: 42n,
      issuedAt: "2026-07-16T10:00:00Z",
      command: {
        case: "sendPrompt",
        value: {
          text: "创建一个视频预览",
          attachments: [],
        },
      },
    });

    expect(command.payloadDigest).toBe(
      "sha256:f47c597114f8300cba123545827f460f98e98978a9e5e7036ac437fdd9c1e47b",
    );
  });

  it("changes the digest when the command payload changes", async () => {
    const base = {
      sessionId: "conv-1",
      streamEpoch: "epoch-1",
      commandId: "command-1",
      expectedRevision: 42n,
      issuedAt: "2026-07-16T10:00:00Z",
    } as const;
    const first = await createAgentCommandEnvelope({
      ...base,
      command: {
        case: "sendPrompt",
        value: { text: "first", attachments: [] },
      },
    });
    const second = await createAgentCommandEnvelope({
      ...base,
      command: {
        case: "sendPrompt",
        value: { text: "second", attachments: [] },
      },
    });

    expect(first.payloadDigest).not.toBe(second.payloadDigest);
  });
});
