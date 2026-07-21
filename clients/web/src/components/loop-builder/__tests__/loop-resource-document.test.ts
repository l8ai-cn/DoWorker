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
      canonicalSource: loopSource,
    });

    expect(document.format).toBe(SourceFormat.JSON);
    const resource = JSON.parse(String(document.content));
    expect(resource).toMatchObject({
      apiVersion: "agentcloud.io/v1alpha1",
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
        loopProgram: {
          canonicalSource: loopSource,
        },
      },
    });
    expect(JSON.stringify(resource.spec).toLowerCase()).not.toContain("worker_ref");
  });

  it("fails closed before planning invalid or incomplete drafts", () => {
    expect(() => createGoalLoopResourceDocument({
      namespace: "dev-org",
      workerTemplateName: "ppt-worker",
      program: undefined,
      canonicalSource: loopSource,
    })).toThrow("must pass validation");
  });

  it("persists the exact custom block pin with the canonical program", () => {
    const program = loopProgram();
    program.repeat!.customBlock = {
      nodeId: "n-ppt-step",
      definitionId: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
      slug: "ppt-step",
      version: 2n,
      definitionDigest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
    };

    const document = createGoalLoopResourceDocument({
      namespace: "dev-org",
      workerTemplateName: "ppt-worker",
      program,
      canonicalSource: loopSource,
    });

    expect(JSON.parse(String(document.content)).spec.loopProgram.customBlock).toEqual({
      nodeId: "n-ppt-step",
      definitionId: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
      slug: "ppt-step",
      version: 2,
      definitionDigest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
    });
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

const loopSource = `@id(n-loop)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-repeat)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-agent)
    agent ppt-task { prompt """制作季度复盘 PPT""" }
    @id(n-verify)
    verify tests { command "test -f output.pptx" accept "PPT 文件可打开" }
  }
  on_failure pause
}`;
