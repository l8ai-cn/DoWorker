import { create } from "zustand";
import { useMemo } from "react";
import { create as protoCreate, toBinary, fromBinary } from "@bufbuild/protobuf";
import type { WorkflowData, WorkflowRunData, RunStatus } from "@/lib/viewModels/workflow";
import { getWorkflowState } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";
import { reconnectRegistry } from "@/lib/realtime";
import { getErrorMessage } from "@/lib/utils";
import {
  listWorkflowsRaw as listWorkflowsRawConnect,
  getWorkflowRaw as getWorkflowRawConnect,
  enableWorkflow as enableWorkflowConnect,
  disableWorkflow as disableWorkflowConnect,
  triggerWorkflow as triggerWorkflowConnect,
  listWorkflowRunsRaw as listWorkflowRunsRawConnect,
  cancelWorkflowRun as cancelWorkflowRunConnect,
} from "@/lib/api/facade/workflowConnect";
import {
  ClearCurrentWorkflowRequestSchema,
  ClearWorkflowRunsRequestSchema, InsertWorkflowRunRequestSchema,
  PatchWorkflowFromActionRequestSchema, PatchWorkflowRunStatusRequestSchema,
  ReplaceCachedWorkflowRunsRequestSchema, ReplaceCachedWorkflowsRequestSchema,
  SetCurrentWorkflowRequestSchema,
} from "@proto/workflow_state/v1/workflow_state_pb";
import { ListWorkflowRunsResponseSchema } from "@proto/workflow/v1/workflow_pb";
import { workflowToProtoWorkflow, workflowRunToProtoWorkflowRun } from "@/lib/api/workflowProtoMap";
import { workflowToCache, workflowRunToCache } from "@/lib/api/projections";

export type { WorkflowData, WorkflowRunData, RunStatus };

const svc = () => getWorkflowState();
const bump = () => useWorkflowStore.setState((s) => ({ _tick: s._tick + 1 }));

function orgSlug(): string {
  return readCurrentOrg()?.slug ?? "";
}

// Read side (B, zero-JSON): UI is a projection of state proto bytes decoded via
// fromBinary + workflowToCache/workflowRunToCache (shared projection). WorkflowData is a
// lossy subset of proto.workflow.v1.Workflow on the Rust side — the same fields the old
// workflows_json read path dropped, so no UI regression.
export function useWorkflows(): WorkflowData[] {
  const tick = useWorkflowStore((s) => s._tick);
  return useMemo(
    () => fromBinary(ReplaceCachedWorkflowsRequestSchema, svc().workflows_bytes()).workflows.map(workflowToCache),
    [tick],
  );
}

export function useCurrentWorkflow(): WorkflowData | null {
  const tick = useWorkflowStore((s) => s._tick);
  return useMemo(() => {
    const bytes = svc().current_workflow_bytes();
    if (bytes.length === 0) return null;
    const workflow = fromBinary(SetCurrentWorkflowRequestSchema, bytes).workflow;
    return workflow ? workflowToCache(workflow) : null;
  }, [tick]);
}

export function useWorkflowRuns(): WorkflowRunData[] {
  const tick = useWorkflowStore((s) => s._tick);
  return useMemo(
    () => fromBinary(ReplaceCachedWorkflowRunsRequestSchema, svc().runs_bytes()).runs.map(workflowRunToCache),
    [tick],
  );
}

function setCurrentWorkflow(workflow: WorkflowData): void {
  const req = protoCreate(SetCurrentWorkflowRequestSchema, { workflow: workflowToProtoWorkflow(workflow) });
  svc().set_current_workflow(toBinary(SetCurrentWorkflowRequestSchema, req));
}

function clearCurrentWorkflow(): void {
  const req = protoCreate(ClearCurrentWorkflowRequestSchema, {});
  svc().clear_current_workflow(toBinary(ClearCurrentWorkflowRequestSchema, req));
}

function patchWorkflowFromAction(slug: string, workflow: WorkflowData): void {
  const req = protoCreate(PatchWorkflowFromActionRequestSchema, {
    slug, workflow: workflowToProtoWorkflow(workflow),
  });
  svc().patch_workflow_from_action(toBinary(PatchWorkflowFromActionRequestSchema, req));
}

function insertWorkflowRun(run: WorkflowRunData): void {
  const req = protoCreate(InsertWorkflowRunRequestSchema, { run: workflowRunToProtoWorkflowRun(run) });
  svc().insert_workflow_run(toBinary(InsertWorkflowRunRequestSchema, req));
}

function patchWorkflowRunStatus(runId: number, status: string): void {
  const req = protoCreate(PatchWorkflowRunStatusRequestSchema, {
    runId: BigInt(runId), status,
  });
  svc().patch_workflow_run_status(toBinary(PatchWorkflowRunStatusRequestSchema, req));
}

function clearWorkflowRuns(): void {
  const req = protoCreate(ClearWorkflowRunsRequestSchema, {});
  svc().clear_workflow_runs(toBinary(ClearWorkflowRunsRequestSchema, req));
}

interface WorkflowStoreState {
  _tick: number;
  loading: boolean; workflowLoading: boolean; runsLoading: boolean;
  error: string | null; totalCount: number; runsTotalCount: number;
  fetchWorkflows: (filters?: { query?: string; status?: string }) => Promise<void>;
  fetchWorkflow: (slug: string) => Promise<void>;
  enableWorkflow: (slug: string) => Promise<void>;
  disableWorkflow: (slug: string) => Promise<void>;
  triggerWorkflow: (slug: string) => Promise<{ run?: WorkflowRunData; skipped?: boolean; reason?: string }>;
  fetchRuns: (slug: string, filters?: { status?: string; limit?: number; offset?: number }) => Promise<void>;
  loadMoreRuns: (slug: string) => Promise<void>;
  cancelRun: (slug: string, runId: number) => Promise<void>;
  setCurrentWorkflow: (workflow: WorkflowData | null) => void;
  getWorkflowBySlug: (slug: string) => WorkflowData | undefined;
  clearError: () => void;
}

export const useWorkflowStore = create<WorkflowStoreState>((set, get) => ({
  _tick: 0,
  loading: false, workflowLoading: false, runsLoading: false,
  error: null, totalCount: 0, runsTotalCount: 0,

  fetchWorkflows: async (filters) => {
    set({ loading: true, error: null });
    try {
      const respBytes = await listWorkflowsRawConnect(orgSlug(), {
        status: filters?.status, query: filters?.query, limit: 500,
      });
      svc().apply_fetched_workflows(respBytes);
      set({ loading: false, _tick: get()._tick + 1 });
    } catch (err) { set({ error: getErrorMessage(err, "An error occurred"), loading: false }); }
  },

  fetchWorkflow: async (slug) => {
    const curBytes = svc().current_workflow_bytes();
    const curWorkflow = curBytes.length === 0 ? null : fromBinary(SetCurrentWorkflowRequestSchema, curBytes).workflow;
    if ((curWorkflow?.slug ?? null) !== slug) {
      clearWorkflowRuns();
      set({ runsTotalCount: 0, _tick: get()._tick + 1 });
    }
    set({ workflowLoading: true, error: null });
    try {
      const respBytes = await getWorkflowRawConnect(orgSlug(), slug);
      svc().apply_fetched_current_workflow(respBytes);
      set({ workflowLoading: false, _tick: get()._tick + 1 });
    } catch (err) { set({ error: getErrorMessage(err, "An error occurred"), workflowLoading: false }); }
  },

  enableWorkflow: async (slug) => {
    const workflow = await enableWorkflowConnect(orgSlug(), slug);
    patchWorkflowFromAction(slug, workflow);
    bump();
  },
  disableWorkflow: async (slug) => {
    const workflow = await disableWorkflowConnect(orgSlug(), slug);
    patchWorkflowFromAction(slug, workflow);
    bump();
  },

  triggerWorkflow: async (slug) => {
    try {
      const result = await triggerWorkflowConnect(orgSlug(), slug);
      if (result.run) {
        insertWorkflowRun(result.run);
      }
      bump();
      get().fetchRuns(slug, { limit: 20, offset: 0 });
      get().fetchWorkflow(slug);
      return result;
    } catch { return { skipped: true, reason: "trigger skipped or failed" }; }
  },

  fetchRuns: async (slug, filters) => {
    set({ runsLoading: true });
    try {
      const respBytes = await listWorkflowRunsRawConnect(orgSlug(), slug, {
        status: filters?.status, limit: filters?.limit, offset: filters?.offset,
      });
      // total is pagination metadata (not state) — decode it off the same wire
      // bytes that feed apply_*_runs, so no second RPC and no JSON state read.
      const total = Number(fromBinary(ListWorkflowRunsResponseSchema, respBytes).total);
      if ((filters?.offset ?? 0) > 0) svc().apply_appended_runs(respBytes);
      else svc().apply_fetched_runs(respBytes);
      set({ runsTotalCount: total, runsLoading: false, _tick: get()._tick + 1 });
    } catch (err) { set({ error: getErrorMessage(err, "An error occurred"), runsLoading: false }); }
  },

  loadMoreRuns: async (slug) => {
    if (get().runsLoading) return;
    const loaded = fromBinary(ReplaceCachedWorkflowRunsRequestSchema, svc().runs_bytes()).runs.length;
    await get().fetchRuns(slug, { limit: 20, offset: loaded });
  },

  cancelRun: async (slug, runId) => {
    await cancelWorkflowRunConnect(orgSlug(), slug, runId);
    patchWorkflowRunStatus(runId, "cancelled");
    bump();
    get().fetchRuns(slug, { limit: 20, offset: 0 });
    get().fetchWorkflow(slug);
  },

  setCurrentWorkflow: (workflow) => {
    if (workflow) setCurrentWorkflow(workflow);
    else clearCurrentWorkflow();
    bump();
  },

  getWorkflowBySlug: (slug) => {
    const workflows = fromBinary(ReplaceCachedWorkflowsRequestSchema, svc().workflows_bytes()).workflows;
    const found = workflows.find((l) => l.slug === slug);
    return found ? workflowToCache(found) : undefined;
  },

  clearError: () => set({ error: null }),
}));

reconnectRegistry.register({
  name: "workflow:list",
  fn: () => useWorkflowStore.getState().fetchWorkflows?.(),
  priority: "low",
});
