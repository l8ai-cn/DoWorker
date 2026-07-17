import { fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  CommandEnvelopeSchema,
  CommandReceiptState,
  SessionSnapshotSchema,
  SessionStatus,
  createLosslessSessionFixture,
} from "../index";

const unsupportedBytes = Uint8Array.from([0, 255, 16, 128, 64, 10]);
const terminalInputBytes = new TextEncoder().encode("pnpm test\r");

describe("lossless Agent Workbench V2 session fixture", () => {
  it("round-trips every required session field without changing binary bytes", () => {
    const fixture = createLosslessSessionFixture();
    const bytes = toBinary(SessionSnapshotSchema, fixture.snapshot);
    const decoded = fromBinary(SessionSnapshotSchema, bytes);

    expect(toBinary(SessionSnapshotSchema, decoded)).toEqual(bytes);
    expect(decoded.revision).toBe(9_007_199_254_740_993n);
    expect(decoded.latestSequence).toBe(9_007_199_254_740_995n);
    expect(decoded.status).toBe(SessionStatus.FAILED);
    expect(decoded.error?.code).toBe("fixture_runner_failed");

    const message = decoded.history[0]?.content?.content;
    expect(message?.case).toBe("message");
    if (message?.case !== "message") {
      throw new Error("fixture message timeline item is missing");
    }

    expect(message.value.content.map((block) => block.content.case)).toEqual([
      "markdown",
      "image",
      "video",
      "presentation",
      "unsupported",
    ]);
    const unknown = message.value.content[4]?.content;
    expect(unknown?.case).toBe("unsupported");
    if (unknown?.case !== "unsupported") {
      throw new Error("fixture unsupported content is missing");
    }
    expect(unknown.value.payload?.data).toEqual(unsupportedBytes);

    const tool = decoded.history[1]?.content?.content;
    expect(tool?.case).toBe("toolExecution");
    if (tool?.case !== "toolExecution") {
      throw new Error("fixture tool execution is missing");
    }
    expect(Array.from(tool.value.input?.data ?? [])).toEqual(
      Array.from(new TextEncoder().encode('{"path":"/workspace/README.md"}')),
    );
    expect(tool.value.results[0]?.resultId).toBe("tool-result-1");
    expect(tool.value.results[0]?.blocks[0]?.content.case).toBe("markdown");
    expect(decoded.history[1]?.envelope?.causationCommandId).toBe("running-command-1");

    const permission = decoded.permissionRequests[0];
    expect(permission?.permissionRequestId).toBe("permission-1");
    expect(permission?.artifactRevision).toBe(3n);
    expect(permission?.request.case).toBe("approval");
    const resolvedPermission = decoded.permissionRequests.find(
      (request) => request.permissionRequestId === "permission-2",
    );
    expect(resolvedPermission?.resolution?.resolvedAt).toBe(
      "2026-07-16T00:00:04Z",
    );

    const deck = decoded.artifacts.find((artifact) => artifact.artifactId === "deck-1");
    expect(deck?.revision).toBe(3n);
    expect(deck?.revisions[0]?.baseRevision).toBe(2n);
    expect(deck?.manifest?.manifest.case).toBe("presentation");

    const terminal = decoded.resources[0]?.resource;
    expect(terminal?.case).toBe("terminal");
    if (terminal?.case !== "terminal") {
      throw new Error("fixture terminal resource is missing");
    }
    expect(terminal.value.lease?.leaseId).toBe("terminal-lease-1");
    expect(terminal.value.lease?.fencingEpoch).toBe(4_294_967_297n);

    const runningReceipt = decoded.commandReceipts.find(
      (receipt) => receipt.commandId === "running-command-1",
    );
    const terminalReceipt = decoded.commandReceipts.find(
      (receipt) => receipt.commandId === "terminal-command-1",
    );
    expect(runningReceipt?.state).toBe(CommandReceiptState.RUNNING);
    expect(terminalReceipt?.state).toBe(CommandReceiptState.SUCCEEDED);
    expect(terminalReceipt?.resultingRevision).toBe(9_007_199_254_740_993n);
  });

  it("round-trips the terminal command and its bigint fencing values byte-stably", () => {
    const fixture = createLosslessSessionFixture();
    const bytes = toBinary(CommandEnvelopeSchema, fixture.terminalCommand);
    const decoded = fromBinary(CommandEnvelopeSchema, bytes);

    expect(toBinary(CommandEnvelopeSchema, decoded)).toEqual(bytes);
    expect(decoded.expectedRevision).toBe(9_007_199_254_740_992n);
    expect(decoded.command.case).toBe("terminalOperation");
    if (decoded.command.case !== "terminalOperation") {
      throw new Error("fixture terminal command is missing");
    }
    expect(decoded.command.value.leaseId).toBe("terminal-lease-1");
    expect(decoded.command.value.fencingEpoch).toBe(4_294_967_297n);
    expect(decoded.command.value.operation.case).toBe("input");
    if (decoded.command.value.operation.case !== "input") {
      throw new Error("fixture terminal input is missing");
    }
    expect(Array.from(decoded.command.value.operation.value.data)).toEqual(
      Array.from(terminalInputBytes),
    );
  });
});
