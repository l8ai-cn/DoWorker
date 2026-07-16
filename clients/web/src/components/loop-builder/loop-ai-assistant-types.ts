export type LoopAIMode = "generate" | "explain";

export interface LoopAIResource {
  id: string;
  label: string;
}

export interface LoopAIRepairTarget {
  diagnosticCode: string;
  diagnosticLabel: string;
  nodeId: string;
  fieldPath: string;
}

export interface LoopAIRepairPatch {
  nodeId: string;
  fieldPath: string;
  oldValue: bigint;
  newValue: bigint;
}

export interface LoopAIProposal {
  response: Uint8Array;
  currentSource: string;
  proposedSource: string;
  repair?: LoopAIRepairPatch;
}
