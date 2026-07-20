import { describe, expect, it } from "vitest";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import { SourceFormat } from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import { createGoalLoopResourceDocument } from "../loop-resource-document";

describe("createGoalLoopResourceDocument", () => {
  it("maps a validated Loop program to the GoalLoop resource contract", () => {
    const document = createGoalLoopResourceDocument({
      namespace: "dev-org",
      workerTemplateName: "ppt-worker",
      program: loopProgram(),
    });

    expect(document.format).toBe(SourceFormat.JSON);
    const resource = JSON.parse(String(document.content));
    expect(resource).toMatchObject({
      apiVersion: "agentsmesh.io/v1alpha1",
      kind: "GoalLoop",
      metadata: {
        name: "checkout-fix",
        namespace: "dev-org",
        displayName: "checkout-fix",
      },
      spec: {
        workerTemplateRef: {
          kind: "WorkerTemplate",
          name: "ppt-worker",
        },
        objective: "制作季度复盘 PPT",
        acceptanceCriteria: ["PPT 文件可打开"],
        verificationCommand: "test -f output.pptx",
        maxIterations: 5,
        tokenBudget: 80000,
        timeoutMinutes: 60,
        noProgressLimit: 3,
        sameErrorLimit: 2,
        escalationPolicy: "pause",
      },
    });
    expect(JSON.stringify(resource.spec).toLowerCase()).not.toContain("worker_ref");
  });

  it("fails closed before planning invalid or incomplete drafts", () => {
    expect(() => createGoalLoopResourceDocument({
      namespace: "dev-org",
      workerTemplateName: "ppt-worker",
      program: undefined,
    })).toThrow("must pass validation");
  });
});

function loopProgram(): LoopProgram {
  return {
    $typeName: "proto.goalloop.v1.LoopProgram",
    schemaVersion: 1,
    loop: { nodeId: "n-loop", localId: "checkout-fix" },
    limits: {
      iterations: 5n,
      tokens: 80000n,
      timeoutMinutes: 60n,
      noProgress: 3n,
      sameError: 2n,
    },
    repeat: {
      identity: { nodeId: "n-repeat", localId: "fix-cycle" },
      max: 5n,
      until: { localId: "tests", field: "passed" },
      agent: {
        identity: { nodeId: "n-agent", localId: "ppt-task" },
        prompt: "制作季度复盘 PPT",
      },
      verifier: {
        identity: { nodeId: "n-verify", localId: "tests" },
        command: "test -f output.pptx",
        accept: "PPT 文件可打开",
      },
    },
    failurePolicy: "pause",
  };
}
