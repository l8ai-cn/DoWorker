import {
  decodeLoopAIDraft,
  decodeLoopAIRepair,
  requestLoopAIDraft,
  requestLoopAIRepair,
} from "@/lib/api/facade/loopProgramConnect";
import type { LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import type {
  LoopAIProposal,
  LoopAIRepairTarget,
} from "./loop-ai-assistant-types";

interface RequestLoopAIProposalInput {
  orgSlug: string;
  locale: string;
  snapshot: LoopWorkbenchSnapshot;
  modelResourceId: number;
  prompt: string;
  repairTarget?: LoopAIRepairTarget;
}

export async function requestLoopAIProposal({
  orgSlug,
  locale,
  snapshot,
  modelResourceId,
  prompt,
  repairTarget,
}: RequestLoopAIProposalInput): Promise<LoopAIProposal> {
  if (repairTarget) {
    const response = await requestLoopAIRepair({
      orgSlug,
      source: snapshot.source,
      modelResourceId,
      locale,
      revision: snapshot.revision,
      diagnosticCode: repairTarget.diagnosticCode,
      nodeId: repairTarget.nodeId,
      fieldPath: repairTarget.fieldPath,
      prompt,
    });
    const decoded = decodeLoopAIRepair(response, {
      revision: snapshot.revision,
      nodeId: repairTarget.nodeId,
      fieldPath: repairTarget.fieldPath,
    });
    return {
      response: decoded.proposalBytes,
      currentSource: snapshot.source,
      proposedSource: decoded.proposal.canonicalSource,
      repair: decoded.patch,
    };
  }

  const response = await requestLoopAIDraft({
    orgSlug,
    prompt,
    currentSource: snapshot.source,
    modelResourceId,
    locale,
    revision: snapshot.revision,
  });
  const decoded = decodeLoopAIDraft(response);
  return {
    response,
    currentSource: snapshot.source,
    proposedSource: decoded.canonicalSource,
  };
}
