import { useEffect, type Dispatch } from "react";
import type {
  ResourceDraftAction,
  ResourceDraftState,
} from "./resource-draft-reducer";

export function useResourcePlanExpiry(
  plan: ResourceDraftState["plan"],
  applyInProgress: boolean,
  dispatch: Dispatch<ResourceDraftAction>,
) {
  useEffect(() => {
    if (applyInProgress || plan.status !== "ready" || !plan.response.plan) return;
    const expiresAt = Date.parse(plan.response.plan.expiresAt);
    const delay = Number.isFinite(expiresAt)
      ? Math.max(0, expiresAt - Date.now())
      : 0;
    const timer = window.setTimeout(() => {
      dispatch({ type: "plan_expired" });
    }, Math.min(delay, 2_147_483_647));
    return () => window.clearTimeout(timer);
  }, [applyInProgress, dispatch, plan]);
}
