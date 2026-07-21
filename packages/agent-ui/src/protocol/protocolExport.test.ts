import { describe, expect, it } from "vitest";

import {
  SessionSnapshotSchema,
  createLosslessSessionFixture,
  resolveSourceTool,
} from "@agent-cloud/agent-ui/protocol";

describe("@agent-cloud/agent-ui/protocol", () => {
  it("exports generated V2 schemas, the fixture builder, and the source catalog", () => {
    expect(createLosslessSessionFixture().snapshot.$typeName).toBe(
      "proto.agent_workbench.v2.SessionSnapshot",
    );
    expect(SessionSnapshotSchema).toBeDefined();
    expect(resolveSourceTool("claude", "Read")?.semanticKey).toBe("filesystem.read");
  });
});
