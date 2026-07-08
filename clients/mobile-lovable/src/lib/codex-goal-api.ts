import { apiFetch } from "./api-fetch";

export type CodexGoalMode = "active" | "paused";

export interface SetCodexGoalParams {
  objective: string;
  tokenBudget?: number | null;
  status?: CodexGoalMode;
}

export async function setCodexGoal(sessionId: string, params: SetCodexGoalParams): Promise<void> {
  const body: Record<string, string | number | null> = {
    objective: params.objective.trim(),
  };
  if (params.tokenBudget !== undefined) {
    body.token_budget = params.tokenBudget;
  }
  if (params.status !== undefined) {
    body.status = params.status;
  }
  const res = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/codex_goal`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(await res.text());
}
