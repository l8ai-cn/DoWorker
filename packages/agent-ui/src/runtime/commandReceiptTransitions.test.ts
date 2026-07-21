import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  CommandReceiptSchema,
  CommandReceiptState,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import { mergeTransportCommandReceipt } from "./commandReceiptTransitions";

describe("transport command receipt merge", () => {
  it("keeps same-state metadata updates", () => {
    const merged = mergeTransportCommandReceipt(
      [receipt(CommandReceiptState.ACCEPTED, "old")],
      receipt(CommandReceiptState.ACCEPTED, "new"),
      "session-1",
    );

    expect(merged?.[0]?.updatedAt).toBe("new");
  });

  it("ignores a stale predecessor response", () => {
    const merged = mergeTransportCommandReceipt(
      [receipt(CommandReceiptState.RUNNING, "new")],
      receipt(CommandReceiptState.RECEIVED, "old"),
      "session-1",
    );

    expect(merged).toBeUndefined();
  });

  it("rejects changes to a terminal receipt at the same state", () => {
    const current = receipt(
      CommandReceiptState.SUCCEEDED,
      "2026-07-16T00:00:02Z",
    );
    current.resultingRevision = 2n;
    const changed = receipt(
      CommandReceiptState.SUCCEEDED,
      "2026-07-16T00:00:03Z",
    );
    changed.resultingRevision = 3n;

    expect(() =>
      mergeTransportCommandReceipt([current], changed, "session-1"),
    ).toThrow("receipt_terminal");
  });

  it("accepts a higher non-terminal revision despite clock skew", () => {
    const current = receipt(
      CommandReceiptState.ACCEPTED,
      "2026-07-16T00:00:02Z",
    );
    current.resultingRevision = 2n;
    const candidate = receipt(
      CommandReceiptState.ACCEPTED,
      "2026-07-16T00:00:01Z",
    );
    candidate.resultingRevision = 3n;

    expect(
      mergeTransportCommandReceipt([current], candidate, "session-1")?.[0],
    ).toMatchObject({
      resultingRevision: 3n,
      updatedAt: "2026-07-16T00:00:02Z",
    });
  });
});

function receipt(state: CommandReceiptState, updatedAt: string) {
  return create(CommandReceiptSchema, {
    sessionId: "session-1",
    commandId: "command-1",
    payloadDigest: "sha256:command",
    state,
    updatedAt,
  });
}
