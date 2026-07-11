import { useCallback, useEffect, useRef } from "react";
import type { Dispatch } from "react";
import type { PodData } from "@/lib/api";
import { podApi } from "@/lib/api";
import { estimateWorkspaceTerminalSize } from "@/lib/terminal-size";
import {
  workerCreateError,
  workerPreflightHasBlockingIssues,
} from "./workerCreateController";
import type {
  WorkerCreateDraftAction,
  WorkerCreateDraftState,
} from "./workerCreateDraft";

interface WorkerCreateSubmissionParams {
  dispatch: Dispatch<WorkerCreateDraftAction>;
  state: WorkerCreateDraftState;
  ticketSlug?: string;
  onSuccess?: (pod: PodData) => void;
  onError?: (error: Error) => void;
}

export function useWorkerCreateSubmission(
  params: WorkerCreateSubmissionParams,
): () => Promise<PodData | null> {
  const createInFlightRef = useRef(false);
  const createCompletedRef = useRef(false);

  useEffect(() => {
    if (
      params.state.create.status === "idle" &&
      !createInFlightRef.current
    ) {
      createCompletedRef.current = false;
    }
  }, [params.state.create.status]);

  return useCallback(async () => {
    const checked = params.state.preflight.status === "ready"
      ? params.state.preflight.data
      : null;
    if (
      createInFlightRef.current ||
      createCompletedRef.current ||
      params.state.create.status === "loading" ||
      params.state.create.status === "ready" ||
      !checked ||
      workerPreflightHasBlockingIssues(checked) ||
      !checked.resolved_spec_json
    ) {
      return null;
    }

    createInFlightRef.current = true;
    params.dispatch({ type: "create_loading" });
    try {
      const { cols, rows } = estimateWorkspaceTerminalSize();
      const result = await podApi.create({
        agent_slug: "",
        ticket_slug: params.ticketSlug,
        cols,
        rows,
        worker_spec: params.state.draft,
      });
      params.dispatch({ type: "create_succeeded", pod: result.pod });
      createCompletedRef.current = true;
      params.onSuccess?.(result.pod);
      return result.pod;
    } catch (error) {
      const resolved = workerCreateError(error);
      params.dispatch({ type: "create_failed", error: resolved.message });
      params.onError?.(resolved);
      return null;
    } finally {
      createInFlightRef.current = false;
    }
  }, [params]);
}
