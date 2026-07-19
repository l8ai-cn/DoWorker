import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import type { WorkerCreationSelection } from "./workerSessionPlan";

type WorkerCreationAgent = AvailableAgent & {
  workerTypeSlug: string;
  supportedModes: NonNullable<AvailableAgent["supportedModes"]>;
  requiresModelResource: boolean;
};

export function workerCreationSelection(agent: AvailableAgent): WorkerCreationSelection {
  if (
    agent.workerTypeSlug === undefined ||
    agent.supportedModes === undefined ||
    agent.requiresModelResource === undefined
  ) {
    throw new Error("Worker creation metadata is unavailable for this Agent");
  }
  return {
    workerTypeSlug: agent.workerTypeSlug,
    supportedModes: agent.supportedModes,
    requiresModelResource: agent.requiresModelResource,
  };
}

export function hasWorkerCreationSelection(agent: AvailableAgent): agent is WorkerCreationAgent {
  return (
    agent.workerTypeSlug !== undefined &&
    agent.supportedModes !== undefined &&
    agent.requiresModelResource !== undefined
  );
}
