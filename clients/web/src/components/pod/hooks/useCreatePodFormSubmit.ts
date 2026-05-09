import type { PodData } from "@/lib/api";
import { getPodService } from "@/lib/wasm-core";

export interface CreatePodResult {
  pod: PodData;
  warning?: string;
}

export async function submitCreatePod(params: {
  selectedAgent: string;
  alias: string;
  perpetual?: boolean;
  selectedRunnerId: number | null | undefined;
  agentfileLayer?: string;
  options?: { ticketSlug?: string; cols?: number; rows?: number };
}): Promise<CreatePodResult | null> {
  const { selectedAgent, alias, perpetual, selectedRunnerId, agentfileLayer, options } = params;

  const raw = await getPodService().create_pod(JSON.stringify({
    agent_slug: selectedAgent,
    runner_id: selectedRunnerId || undefined,
    alias: alias.trim() || undefined,
    ticket_slug: options?.ticketSlug,
    cols: options?.cols,
    rows: options?.rows,
    agentfile_layer: agentfileLayer || undefined,
    perpetual: perpetual || undefined,
  }));

  const parsed = JSON.parse(raw) as { pod: PodData; warning?: string };
  if (!parsed?.pod) return null;
  return { pod: parsed.pod, warning: parsed.warning };
}
