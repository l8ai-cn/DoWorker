import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  CompileLoopProgramRequestSchema,
  GoalLoopActionRequestSchema,
  LoopDraftSnapshotSchema,
} from "@proto/goalloop/v1/goalloop_pb";
import {
  IssueSeverity,
  SourceFormat,
} from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import {
  getGoalLoopService,
  getLoopBuilderState,
  initWasmCore,
} from "@/lib/wasm-core";
import type { ResourceDocument } from "./orchestrationResourceConnect";
import {
  exportResource,
  listResources,
  planResource,
  validateResource,
} from "./orchestrationResourceConnect";
import { createGoalLoopFromPlan } from "./orchestrationResourceApplyConnect";
import type {
  LoopEditor,
  LoopRuntimeTemplate,
  LoopWorkbenchSnapshot,
} from "@/lib/viewModels/loop-program";
import {
  assertGoalLoopProgramSnapshot,
  parseGoalLoopProgramSnapshot,
} from "@/components/loop-builder/loop-resource-snapshot";

export interface StartedLoopResource {
  loopSlug: string;
  snapshot: LoopWorkbenchSnapshot;
}

function toSnapshot(): LoopWorkbenchSnapshot {
  const snapshot = fromBinary(
    LoopDraftSnapshotSchema,
    new Uint8Array(getLoopBuilderState().snapshot_bytes()),
  );
  return {
    source: snapshot.source,
    canonicalSource: snapshot.canonicalSource,
    program: snapshot.program,
    diagnostics: snapshot.diagnostics,
    parseStatus: snapshot.parseStatus,
    activeEditor: snapshot.activeEditor as LoopEditor,
    revision: Number(snapshot.revision),
    semanticRevision: Number(snapshot.semanticRevision),
    run: snapshot.run,
  };
}

export async function readLoopSnapshot(): Promise<LoopWorkbenchSnapshot> {
  await initWasmCore();
  return toSnapshot();
}

export async function setLoopSource(
  source: string,
  editor: LoopEditor,
): Promise<LoopWorkbenchSnapshot> {
  await initWasmCore();
  getLoopBuilderState().set_source(source, editor);
  return toSnapshot();
}

export async function setLoopActiveEditor(
  editor: LoopEditor,
): Promise<LoopWorkbenchSnapshot> {
  await initWasmCore();
  getLoopBuilderState().set_active_editor(editor);
  return toSnapshot();
}

export async function requestLoopCompile(
  orgSlug: string,
  source: string,
  revision: number,
): Promise<Uint8Array> {
  await initWasmCore();
  const request = create(CompileLoopProgramRequestSchema, {
    orgSlug,
    source,
    revision: BigInt(revision),
  });
  return new Uint8Array(
    await getGoalLoopService().compileLoopProgramConnect(
      toBinary(CompileLoopProgramRequestSchema, request),
    ),
  );
}

export async function applyLoopCompile(
  response: Uint8Array,
): Promise<LoopWorkbenchSnapshot> {
  await initWasmCore();
  getLoopBuilderState().apply_compile_response(response);
  return toSnapshot();
}

export async function listLoopRuntimeTemplates(
  orgSlug: string,
): Promise<LoopRuntimeTemplate[]> {
  const response = await listResources(orgSlug, {
    kind: "WorkerTemplate",
    limit: 100,
  });
  return response.items.map((resource) => ({
    id: resource.identity?.target?.name ?? resource.id.toString(),
    alias: resource.displayName || resource.identity?.target?.name ||
      "WorkerTemplate",
    workerType: "WorkerTemplate",
    createdAt: resource.createdAt,
  }));
}

export async function runLoopResourceProgram(
  orgSlug: string,
  document: ResourceDocument,
): Promise<StartedLoopResource> {
  await initWasmCore();
  const validation = await validateResource(orgSlug, document);
  assertNoBlockingIssues(validation.issues);
  const plan = await planResource(orgSlug, document);
  assertNoBlockingIssues(plan.issues);
  if (!plan.plan?.planId) {
    throw new Error("Resource planning did not return an applicable GoalLoop plan.");
  }
  const applied = await createGoalLoopFromPlan(orgSlug, plan.plan.planId);
  const loopSlug = applied.resource?.identity?.target?.name;
  if (!loopSlug) throw new Error("Applied GoalLoop resource did not include a loop name.");
  const request = create(GoalLoopActionRequestSchema, { orgSlug, loopSlug });
  const response = new Uint8Array(
    await getGoalLoopService().startGoalLoopConnect(
      toBinary(GoalLoopActionRequestSchema, request),
    ),
  );
  getLoopBuilderState().apply_run_response(response);
  return { loopSlug, snapshot: toSnapshot() };
}

export async function restoreLoopResourceProgram(
  orgSlug: string,
  loopSlug: string,
): Promise<LoopWorkbenchSnapshot> {
  const content = await exportResource(orgSlug, {
    apiVersion: "agentsmesh.io/v1alpha1",
    kind: "GoalLoop",
    namespace: orgSlug,
    name: loopSlug,
  }, SourceFormat.JSON);
  const stored = parseGoalLoopProgramSnapshot(content);
  const draft = await setLoopSource(stored.canonicalSource, "blocks");
  const response = await requestLoopCompile(
    orgSlug,
    stored.canonicalSource,
    draft.revision,
  );
  const snapshot = await applyLoopCompile(response);
  assertGoalLoopProgramSnapshot(snapshot.program, stored);
  return snapshot;
}

function assertNoBlockingIssues(issues: { severity: IssueSeverity; message: string }[]) {
  const blocking = issues.filter(
    (issue) => issue.severity === IssueSeverity.BLOCKING,
  );
  if (blocking.length > 0) {
    throw new Error(blocking.map((issue) => issue.message).join("\n"));
  }
}
