import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import { hasWorkerCreationSelection, workerCreationSelection } from "@/lib/workerCreationSelection";
import { createWorkerSession } from "@/lib/workerSessionMutations";
import type { SessionInteractionMode } from "@/lib/workerSessionPlan";
import { isNativeCodingAgent } from "@/lib/nativeCodingAgents";

export interface NewChatCreateInput {
  agent: AvailableAgent | null | undefined;
  hostId: string | null;
  workspace: string;
  sandboxSelected: boolean;
  sandboxRepoUrl: string;
  sandboxRepoBranch: string;
  branchName: string;
  modelResourceId: number | null;
  tokenBudget: number | null;
}

export function newChatCreateDisabledReason(input: NewChatCreateInput): string | null {
  const agent = input.agent;
  if (!agent) return "Select an agent";
  if (!hasWorkerCreationSelection(agent)) {
    return "This agent is missing Worker creation metadata. Register a WorkerTemplate before launching it.";
  }
  const mode = newChatInteractionMode(agent);
  if (!agent.supportedModes.includes(mode)) {
    return `This agent does not support ${mode.toUpperCase()} sessions`;
  }
  if (input.sandboxSelected && (input.sandboxRepoUrl.trim() || input.sandboxRepoBranch.trim())) {
    return "Repository-backed sandbox starts need a WorkerTemplate workspace option.";
  }
  if (input.branchName.trim()) {
    return "Git worktree starts need a WorkerTemplate workspace option.";
  }
  return null;
}

export async function createNewChatSession(input: NewChatCreateInput): Promise<{ id: string }> {
  const reason = newChatCreateDisabledReason(input);
  if (reason) throw new Error(reason);
  const agent = input.agent!;
  return createWorkerSession({
    agentId: agent.id,
    initialItems: [],
    mode: newChatInteractionMode(agent),
    hostId: input.sandboxSelected ? undefined : (input.hostId ?? undefined),
    workspace: input.sandboxSelected ? undefined : input.workspace,
    modelResourceId: input.modelResourceId ?? undefined,
    tokenBudget: input.tokenBudget,
    ...workerCreationSelection(agent),
  });
}

export function newChatInteractionMode(agent: AvailableAgent): SessionInteractionMode {
  return isNativeCodingAgent(agent) ? "pty" : "acp";
}
