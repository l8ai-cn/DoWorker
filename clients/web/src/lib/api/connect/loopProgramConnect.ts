import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  CompileLoopProgramRequestSchema,
  LoopDraftSnapshotSchema,
  RunLoopProgramRequestSchema,
} from "@proto/goalloop/v1/goalloop_pb";
import {
  getGoalLoopService,
  getLoopBuilderState,
  initWasmCore,
} from "@/lib/wasm-core";
import type { LoopEditor, LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";

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

export async function runLoopProgram(
  orgSlug: string,
  source: string,
): Promise<LoopWorkbenchSnapshot> {
  await initWasmCore();
  const request = create(RunLoopProgramRequestSchema, { orgSlug, source });
  const response = new Uint8Array(
    await getGoalLoopService().runLoopProgramConnect(
      toBinary(RunLoopProgramRequestSchema, request),
    ),
  );
  getLoopBuilderState().apply_run_response(response);
  return toSnapshot();
}
