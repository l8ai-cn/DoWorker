import { describe, expect, it } from "vitest";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import {
  assertGoalLoopProgramSnapshot,
  parseGoalLoopProgramSnapshot,
} from "../loop-resource-snapshot";

const pin = {
  nodeId: "n-custom",
  definitionId: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
  slug: "ppt-step",
  version: 2,
  definitionDigest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
};

describe("GoalLoop resource snapshot", () => {
  it("restores the exact persisted custom block pin", () => {
    const snapshot = parseGoalLoopProgramSnapshot(new TextEncoder().encode(JSON.stringify({
      apiVersion: "agentsmesh.io/v1alpha1",
      kind: "GoalLoop",
      metadata: { name: "checkout-loop", namespace: "acme" },
      spec: {
        workerTemplateRef: { kind: "WorkerTemplate", name: "reviewer" },
        description: "Checkout",
        objective: "Fix checkout",
        acceptanceCriteria: ["Tests pass"],
        verificationCommand: "go test ./...",
        maxIterations: 5,
        tokenBudget: 80000,
        timeoutMinutes: 60,
        noProgressLimit: 3,
        sameErrorLimit: 2,
        escalationPolicy: "pause",
        loopProgram: { canonicalSource: "loop source", customBlock: pin },
      },
    })));

    expect(snapshot).toEqual({ canonicalSource: "loop source", customBlock: pin });
    expect(() => assertGoalLoopProgramSnapshot(loopProgram(pin), snapshot)).not.toThrow();
  });

  it("fails closed for a missing or substituted pin", () => {
    const snapshot = { canonicalSource: "loop source", customBlock: pin };

    expect(() => assertGoalLoopProgramSnapshot(loopProgram({
      ...pin,
      definitionDigest: "f".repeat(64),
    }), snapshot)).toThrow("does not match");
    expect(() => assertGoalLoopProgramSnapshot(loopProgram(), snapshot)).toThrow("does not match");
  });
});

function loopProgram(customBlock?: typeof pin): LoopProgram {
  return {
    $typeName: "proto.goalloop.v1.LoopProgram",
    schemaVersion: 1,
    repeat: customBlock ? { customBlock } : {},
  };
}
