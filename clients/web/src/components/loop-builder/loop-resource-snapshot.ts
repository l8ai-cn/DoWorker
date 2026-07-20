import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import {
  parseCanonicalResourceJson,
} from "@/components/resource-editor/resource-yaml-codec";
import type {
  GoalLoopDraft,
  GoalLoopProgramSnapshot,
} from "@/components/resource-editor/resource-definition-draft-types";

export function parseGoalLoopProgramSnapshot(
  content: Uint8Array,
): GoalLoopProgramSnapshot {
  const draft = parseCanonicalResourceJson(content, "GoalLoop") as GoalLoopDraft;
  if (!draft.spec.loopProgram) {
    throw new Error("GoalLoop resource does not contain a persisted loop program.");
  }
  return draft.spec.loopProgram;
}

export function assertGoalLoopProgramSnapshot(
  program: LoopProgram | undefined,
  snapshot: GoalLoopProgramSnapshot,
): void {
  const source = program?.repeat?.customBlock;
  const pin = snapshot.customBlock;
  if (!source && !pin) return;
  if (
    !source ||
    !pin ||
    source.nodeId !== pin.nodeId ||
    source.definitionId !== pin.definitionId ||
    source.slug !== pin.slug ||
    Number(source.version) !== pin.version ||
    source.definitionDigest !== pin.definitionDigest
  ) {
    throw new Error(
      "GoalLoop custom block pin does not match the persisted canonical program.",
    );
  }
}
