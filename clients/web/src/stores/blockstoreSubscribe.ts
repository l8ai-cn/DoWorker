import { create as protoCreate, toBinary } from "@bufbuild/protobuf";
import { reconnectRegistry } from "@/lib/realtime";
import type { RealtimeEvent } from "@/lib/realtime";
import type { BlockOp } from "@/lib/viewModels/blockstore";
import { getBlockstoreService } from "@/lib/wasm-core";
import { ApplyRemoteOpRequestSchema } from "@proto/blockstore_state/v1/blockstore_state_pb";
import { useBlockstoreStore, readLastOpIds } from "./blockstore";

// Coalesce _tick bumps: per-op bumps during catchup triggered React #185.
let bumpTimer: ReturnType<typeof setTimeout> | null = null;
function scheduleBump() {
  if (bumpTimer) return;
  bumpTimer = setTimeout(() => {
    bumpTimer = null;
    useBlockstoreStore.setState((s) => ({ _tick: s._tick + 1 }));
  }, 100);
}

function dispatchApplyRemoteOp(op: BlockOp): void {
  const req = protoCreate(ApplyRemoteOpRequestSchema, { opJson: JSON.stringify(op) });
  getBlockstoreService().apply_remote_op(toBinary(ApplyRemoteOpRequestSchema, req));
}

// Backend fans out ops org-wide; drop ops for workspaces the user has not loaded.
export function handleBlockstoreEvent(event: RealtimeEvent) {
  if (event.type !== "blockstore:op") return;
  const op = event.data as BlockOp;
  if (!(op.workspace_id in readLastOpIds())) return;
  // Backend serialises `applied_at` as Unix ms; Rust BlockOp expects string.
  // Same normalisation as Web's `apply_remote_op`.
  const normalized: BlockOp = {
    ...op,
    applied_at:
      typeof op.applied_at === "number"
        ? new Date(op.applied_at).toISOString()
        : op.applied_at,
  };
  try {
    dispatchApplyRemoteOp(normalized);
  } catch {
    return;
  }
  scheduleBump();
}

reconnectRegistry.register({
  name: "blockstore:catchup",
  fn: () => {
    const lastOpId = readLastOpIds();
    const actions = useBlockstoreStore.getState().actions;
    Object.keys(lastOpId).forEach((wsID) => {
      void actions.catchup(wsID);
    });
  },
  priority: "deferred",
});
