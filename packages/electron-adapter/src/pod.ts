import { invoke } from "./invoke";
import type { IPodService } from "@agentsmesh/service-interface";
import { fromBinary } from "@bufbuild/protobuf";
import {
  ReplaceCachedPodsRequestSchema,
  AppendCachedPodsRequestSchema,
  InsertCreatedPodRequestSchema,
  MarkPodTerminatedRequestSchema,
  PatchPodPerpetualRequestSchema,
  ApplyPodStatusEventRequestSchema,
  ApplyPodTitleEventRequestSchema,
  ApplyPodAliasEventRequestSchema,
  ApplyAgentStatusEventRequestSchema,
} from "@agentsmesh/proto/pod_state/v1/pod_state_pb";
import type { Pod as ProtoPod } from "@agentsmesh/proto/pod/v1/pod_pb";

// ProtoPod (camelCase + BigInt) → PodData (snake_case + number). Same logic as
// fromProtoPod in clients/web/src/lib/api/connect/podConnect.ts but inlined to
// keep electron-adapter independent of the web/src tree.
function podToCache(p: ProtoPod): Record<string, unknown> {
  const ZERO = BigInt(0);
  return {
    id: Number(p.id),
    pod_key: p.podKey,
    status: p.status,
    agent_status: p.agentStatus,
    alias: p.alias,
    title: p.title,
    runner: p.runner ? {
      id: p.runner.id === undefined || p.runner.id === ZERO ? undefined : Number(p.runner.id),
      node_id: p.runner.nodeId,
      status: p.runner.status,
    } : undefined,
    agent: p.agent ? { name: p.agent.name, slug: p.agent.slug } : undefined,
    repository: p.repository ? {
      id: p.repository.id === undefined || p.repository.id === ZERO ? undefined : Number(p.repository.id),
      name: p.repository.name,
      slug: p.repository.slug,
      provider_type: p.repository.providerType,
    } : undefined,
    ticket: p.ticket ? {
      id: p.ticket.id === undefined || p.ticket.id === ZERO ? undefined : Number(p.ticket.id),
      slug: p.ticket.slug,
      title: p.ticket.title,
    } : undefined,
    loop: p.loop ? {
      id: p.loop.id === undefined || p.loop.id === ZERO ? undefined : Number(p.loop.id),
      slug: p.loop.slug,
      name: p.loop.name,
    } : undefined,
    created_by: p.createdBy ? {
      id: p.createdBy.id === undefined || p.createdBy.id === ZERO ? undefined : Number(p.createdBy.id),
      username: p.createdBy.username,
      name: p.createdBy.name,
    } : undefined,
    repository_id: undefined,
    ticket_id: undefined,
    loop_id: undefined,
    organization_id: undefined,
    runner_id: p.runnerId !== undefined && p.runnerId !== ZERO ? Number(p.runnerId) : undefined,
    interaction_mode: p.interactionMode,
    agentfile: undefined,
    branch: p.branchName,
    initial_prompt: p.prompt,
    perpetual: p.perpetual,
    error_code: p.errorCode || undefined,
    error_message: p.errorMessage || undefined,
    created_at: p.createdAt,
    updated_at: p.updatedAt,
  };
}

// Apply scalar status patch on a single cached pod. Used by the proto-bytes
// status/title/alias/agent-status mutators below; not exposed publicly.
function patchPodInCache(
  cache: string,
  podKey: string,
  patch: Record<string, unknown>,
): string {
  const pods = JSON.parse(cache) as Array<Record<string, unknown>>;
  const p = pods.find((x) => x.pod_key === podKey);
  if (p) Object.assign(p, patch);
  return JSON.stringify(pods);
}

export class ElectronPodService implements IPodService {
  private _podsCache = "[]";
  private _currentPodCache: string | null = null;

  // ── Read selectors ──

  pods_json(): string { return this._podsCache; }
  current_pod_json(): unknown { return this._currentPodCache; }
  get_pod_json(pod_key: string): unknown {
    const pods = JSON.parse(this._podsCache) as { pod_key: string }[];
    const p = pods.find(x => x.pod_key === pod_key);
    return p ? JSON.stringify(p) : null;
  }

  // ── Proto-bytes mutators (mirror clients/core/crates/wasm WasmPodState) ──
  // Decode locally so `pods_json()` reflects the mutation synchronously, AND
  // fire-and-forget the same bytes to the `app_pod_*` commands so the
  // main-process `runtime.state.pods` (the realtime dispatch + snapshot SSOT)
  // gets the same fetch/user-action baseline. Not awaited — IPC latency would
  // defeat the sync-cache invariant the renderer's _tick reactivity assumes.

  replace_cached_pods(reqBytes: Uint8Array): void {
    const req = fromBinary(ReplaceCachedPodsRequestSchema, reqBytes);
    this._podsCache = JSON.stringify(req.pods.map(podToCache));
    void invoke<void>("appPodReplaceCachedPods", Array.from(reqBytes)).catch(() => undefined);
  }

  append_cached_pods(reqBytes: Uint8Array): void {
    const req = fromBinary(AppendCachedPodsRequestSchema, reqBytes);
    const existing = JSON.parse(this._podsCache) as { pod_key: string }[];
    const seen = new Set(existing.map((p) => p.pod_key));
    for (const p of req.pods) {
      const c = podToCache(p);
      if (!seen.has(c.pod_key as string)) existing.push(c as { pod_key: string });
    }
    this._podsCache = JSON.stringify(existing);
    void invoke<void>("appPodAppendCachedPods", Array.from(reqBytes)).catch(() => undefined);
  }

  insert_created_pod(reqBytes: Uint8Array): void {
    const req = fromBinary(InsertCreatedPodRequestSchema, reqBytes);
    if (!req.pod) return;
    const cache = JSON.parse(this._podsCache) as { pod_key: string }[];
    const c = podToCache(req.pod);
    const idx = cache.findIndex((p) => p.pod_key === c.pod_key);
    if (idx >= 0) cache[idx] = { ...cache[idx], ...c };
    else cache.unshift(c as { pod_key: string });
    this._podsCache = JSON.stringify(cache);
    void invoke<void>("appPodInsertCreated", Array.from(reqBytes)).catch(() => undefined);
  }

  mark_pod_terminated(reqBytes: Uint8Array): void {
    const req = fromBinary(MarkPodTerminatedRequestSchema, reqBytes);
    this._podsCache = patchPodInCache(this._podsCache, req.podKey, { status: "terminated" });
    void invoke<void>("appPodMarkTerminated", Array.from(reqBytes)).catch(() => undefined);
  }

  patch_pod_perpetual(reqBytes: Uint8Array): void {
    const req = fromBinary(PatchPodPerpetualRequestSchema, reqBytes);
    this._podsCache = patchPodInCache(this._podsCache, req.podKey, { perpetual: req.perpetual });
    void invoke<void>("appPodPatchPerpetual", Array.from(reqBytes)).catch(() => undefined);
  }

  // Surgical realtime mirror: the main-pushed snapshot carries one Rust-computed
  // pod. Update it in place ONLY if already cached — the pod sidebar is a
  // FILTERED set, so adding an absent pod would corrupt the filtered view (a
  // brand-new pod is added by the handler's fetchPod refetch instead).
  apply_pod_snapshot(json: string): void {
    let pod: { pod_key?: string };
    try {
      pod = JSON.parse(json) as { pod_key?: string };
    } catch {
      return;
    }
    if (!pod.pod_key) return;
    const cache = JSON.parse(this._podsCache) as Array<Record<string, unknown>>;
    const idx = cache.findIndex((p) => p.pod_key === pod.pod_key);
    if (idx >= 0) {
      cache[idx] = { ...cache[idx], ...pod };
      this._podsCache = JSON.stringify(cache);
    }
  }

  apply_pod_status_event(reqBytes: Uint8Array): void {
    const req = fromBinary(ApplyPodStatusEventRequestSchema, reqBytes);
    const patch: Record<string, unknown> = {};
    if (req.status) {
      patch.status = req.status;
      if (req.errorCode !== undefined) patch.error_code = req.errorCode ?? undefined;
      if (req.errorMessage !== undefined) patch.error_message = req.errorMessage ?? undefined;
    }
    if (req.agentStatus !== undefined) patch.agent_status = req.agentStatus ?? undefined;
    this._podsCache = patchPodInCache(this._podsCache, req.podKey, patch);
  }

  apply_pod_title_event(reqBytes: Uint8Array): void {
    const req = fromBinary(ApplyPodTitleEventRequestSchema, reqBytes);
    this._podsCache = patchPodInCache(this._podsCache, req.podKey, { title: req.title });
  }

  apply_pod_alias_event(reqBytes: Uint8Array): void {
    const req = fromBinary(ApplyPodAliasEventRequestSchema, reqBytes);
    this._podsCache = patchPodInCache(this._podsCache, req.podKey, { alias: req.alias ?? "" });
  }

  apply_agent_status_event(reqBytes: Uint8Array): void {
    const req = fromBinary(ApplyAgentStatusEventRequestSchema, reqBytes);
    this._podsCache = patchPodInCache(this._podsCache, req.podKey, { agent_status: req.agentStatus });
  }
}
