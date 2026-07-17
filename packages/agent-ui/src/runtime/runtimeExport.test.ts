import { describe, expect, it } from "vitest";

import {
  AgentSessionStore,
  applyDeltaBatch,
  applySessionSnapshot,
} from "@do-worker/agent-ui/runtime";

describe("runtime package export", () => {
  it("exposes the generated session store and reducer", () => {
    expect(AgentSessionStore).toBeTypeOf("function");
    expect(applySessionSnapshot).toBeTypeOf("function");
    expect(applyDeltaBatch).toBeTypeOf("function");
  });
});
