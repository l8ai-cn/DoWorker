import { useAutopilotStore } from "@/stores/autopilot";
import { useWorkflowStore } from "@/stores/workflow";
import { getWorkflowState, parseWasmAny } from "@/lib/wasm-core";
import type { DebounceRef } from "./realtimeEventHandlers";
import {
  type RealtimeEvent,
  decodeEventData,
  WorkflowRunEventDataSchema,
  WorkflowRunWarningEventDataSchema,
} from "@/lib/realtime";

export function handleAutopilotEvent(event: RealtimeEvent) {
  const store = useAutopilotStore.getState();
  switch (event.type) {
    case "autopilot:status_changed":
    case "autopilot:iteration":
    case "autopilot:thinking":
    case "autopilot:terminated": {
      // Rust event_dispatch owns the controller/iteration/thinking mutation in
      // runtime.state (update_controller / add_iteration / update_thinking /
      // remove_controller); bump triggers the React selectors to re-read. On
      // Web autopilot state updates the renderer caches.
      useAutopilotStore.setState((s) => ({ _tick: s._tick + 1 }));
      break;
    }
    case "autopilot:created": {
      // New controller needs its full payload from the server.
      store.fetchAutopilotControllers?.();
      break;
    }
  }
}

export function handleWorkflowEvent(
  event: RealtimeEvent,
  debounceRef: DebounceRef | undefined,
  t: (key: string, params?: Record<string, string | number>) => string,
  showWarning: (title: string, description: string) => void
) {
  switch (event.type) {
    case "workflow_run:started":
    case "workflow_run:completed":
    case "workflow_run:failed": {
      if (!debounceRef) return;
      if (debounceRef.current) clearTimeout(debounceRef.current);
      debounceRef.current = setTimeout(() => {
        debounceRef.current = null;
        const s = useWorkflowStore.getState();
        s.fetchWorkflows?.();
        const currentWorkflow = parseWasmAny<{ id: number; slug: string }>(getWorkflowState().current_workflow_json());
        const workflowRunData = decodeEventData(WorkflowRunEventDataSchema, event.data);
        if (currentWorkflow?.id === Number(workflowRunData.workflowId)) {
          s.fetchWorkflow?.(currentWorkflow.slug);
          // Eventual-consistency retry: if the first fetch races the
          // publish path and returns no new rows, retry once at 750ms.
          // Cheap insurance against the multitab broadcast race.
          const slug = currentWorkflow.slug;
          const expectedRunId = Number(workflowRunData.runId);
          s.fetchRuns?.(slug, { limit: 20, offset: 0 }).then(() => {
            if (!Number.isFinite(expectedRunId) || expectedRunId <= 0) return;
            const seen = parseWasmAny<Array<{ id?: number | string }>>(getWorkflowState().runs_json()) ?? [];
            const found = seen.some((r) => Number(r.id) === expectedRunId);
            if (!found) setTimeout(() => s.fetchRuns?.(slug, { limit: 20, offset: 0 }), 750);
          });
        }
      }, 500);
      break;
    }
    case "workflow_run:warning": {
      const data = decodeEventData(WorkflowRunWarningEventDataSchema, event.data);
      showWarning(t("workflows.runWarningTitle", { runNumber: data.runNumber }), data.detail || data.warning);
      break;
    }
  }
}
