import { SourceFormat } from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import type { ResourceDocument } from "@/lib/api/facade/orchestrationResource";
import { RESOURCE_API_VERSION } from "@/components/resource-editor/resource-editor-types";

interface LoopResourceDocumentInput {
  canonicalSource: string;
  program?: LoopProgram;
  workerTemplateName: string;
  namespace: string;
}

export function createGoalLoopResourceDocument({
  canonicalSource,
  program,
  workerTemplateName,
  namespace,
}: LoopResourceDocumentInput): ResourceDocument {
  if (!program?.loop || !program.limits || !program.repeat) {
    throw new Error("Loop draft must pass validation before resource planning.");
  }
  const loopName = required(program.loop.localId, "loop name");
  const repeat = program.repeat;
  const agent = requiredNode(repeat.agent, "agent task");
  const verifier = requiredNode(repeat.verifier, "verification step");
  const until = requiredNode(repeat.until, "repeat condition");
  const limits = program.limits;
  const iterations = positiveInteger(limits.iterations);
  if (positiveInteger(repeat.max) !== iterations) {
    throw new Error("Loop repeat max must match the resource iteration limit.");
  }
  if (until.localId !== verifier.identity?.localId || until.field !== "passed") {
    throw new Error("Loop repeat condition must target the verification step.");
  }
  const tokenBudget = positiveInteger(limits.tokens);
  const customBlock = repeat.customBlock
    ? {
      nodeId: required(repeat.customBlock.nodeId, "custom block node"),
      definitionId: required(repeat.customBlock.definitionId, "custom block definition"),
      slug: required(repeat.customBlock.slug, "custom block slug"),
      version: positiveInteger(repeat.customBlock.version),
      definitionDigest: required(
        repeat.customBlock.definitionDigest,
        "custom block definition digest",
      ),
    }
    : undefined;
  return {
    format: SourceFormat.JSON,
    content: JSON.stringify({
      apiVersion: RESOURCE_API_VERSION,
      kind: "GoalLoop",
      metadata: {
        name: loopName,
        namespace,
        displayName: loopName,
        labels: {},
      },
      spec: {
        workerTemplateRef: {
          kind: "WorkerTemplate",
          name: required(workerTemplateName, "worker template"),
        },
        description: loopName,
        objective: required(agent.prompt, "agent prompt"),
        acceptanceCriteria: [required(verifier.accept, "acceptance criterion")],
        verificationCommand: required(verifier.command, "verification command"),
        maxIterations: iterations,
        tokenBudget,
        timeoutMinutes: positiveInteger(limits.timeoutMinutes),
        noProgressLimit: positiveInteger(limits.noProgress),
        sameErrorLimit: positiveInteger(limits.sameError),
        escalationPolicy: program.failurePolicy === "fail" ? "fail" : "pause",
        loopProgram: {
          canonicalSource: required(canonicalSource, "canonical source"),
          ...(customBlock ? { customBlock } : {}),
        },
      },
    }),
  };
}

function required(value: string | undefined, label: string): string {
  const text = value?.trim() ?? "";
  if (!text) throw new Error(`Loop ${label} is required for resource planning.`);
  return text;
}

function requiredNode<T>(value: T | undefined, label: string): T {
  if (!value) throw new Error(`Loop ${label} is required for resource planning.`);
  return value;
}

function positiveInteger(value: bigint | number | undefined): number {
  const next = Number(value ?? 0);
  if (!Number.isSafeInteger(next) || next < 1) {
    throw new Error("Loop numeric limits must be positive safe integers.");
  }
  return next;
}
