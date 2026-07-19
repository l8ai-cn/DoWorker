import { apiFetch } from "./api-fetch";
import type { AvailableAgent, SessionInteractionMode } from "./sessions-api";

interface AgentWire {
  id: string;
  worker_type_slug?: string;
  name: string;
  harness?: string;
  builtin?: boolean;
  supported_modes?: string[];
  requires_model_resource?: boolean;
}

export async function listMobileAgents(): Promise<AvailableAgent[]> {
  const response = await apiFetch("/v1/agents");
  if (!response.ok) throw new Error((await response.text()) || `HTTP ${response.status}`);
  const body = (await response.json()) as { data: AgentWire[] };
  return (body.data ?? []).map(agentFromWire);
}

function agentFromWire(agent: AgentWire): AvailableAgent {
  if (typeof agent.requires_model_resource !== "boolean") {
    throw new Error(`Worker ${agent.id} 未声明模型资源要求`);
  }
  const workerTypeSlug =
    agent.worker_type_slug ?? (agent.builtin === true ? agent.id : undefined);
  return {
    id: agent.id,
    workerTypeSlug,
    name: agent.name,
    harness: agent.harness ?? null,
    supportedModes: supportedModes(agent.id, agent.supported_modes),
    requiresModelResource: agent.requires_model_resource,
  };
}

function supportedModes(agentID: string, modes: string[] | undefined): SessionInteractionMode[] {
  if (!modes || modes.length === 0) {
    throw new Error(`Worker ${agentID} 未声明支持的交互模式`);
  }
  const supported = modes.filter(
    (mode): mode is SessionInteractionMode => mode === "acp" || mode === "pty",
  );
  if (supported.length !== modes.length || new Set(supported).size !== supported.length) {
    throw new Error(`Worker ${agentID} 返回了无效的交互模式`);
  }
  return supported;
}
