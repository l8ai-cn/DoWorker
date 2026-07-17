import { clone, equals } from "@bufbuild/protobuf";

import type { SessionSnapshot } from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionSnapshotSchema } from "@do-worker/proto/agent_workbench/v2/session_pb";
import { AgentSessionReductionError } from "./agentSessionReducer";

export function validateSnapshotAdvance(
  current: SessionSnapshot,
  next: SessionSnapshot,
): "advance" | "identical" | "metadata" {
  if (next.sessionId !== current.sessionId) {
    throw new AgentSessionReductionError("snapshot_session_mismatch");
  }
  if (next.streamEpoch !== current.streamEpoch) return "advance";
  if (
    next.revision < current.revision ||
    next.latestSequence < current.latestSequence
  ) {
    throw new AgentSessionReductionError("snapshot_stale");
  }
  if (
    next.revision !== current.revision ||
    next.latestSequence !== current.latestSequence
  ) {
    return "advance";
  }
  if (equals(SessionSnapshotSchema, current, next)) return "identical";

  const currentContent = clone(SessionSnapshotSchema, current);
  const nextContent = clone(SessionSnapshotSchema, next);
  clearViewerMetadata(currentContent);
  clearViewerMetadata(nextContent);
  if (equals(SessionSnapshotSchema, currentContent, nextContent)) {
    return "metadata";
  }
  throw new AgentSessionReductionError("snapshot_cursor_conflict");
}

function clearViewerMetadata(snapshot: SessionSnapshot): void {
  snapshot.digest = undefined;
  snapshot.grants = [];
  for (const artifact of snapshot.artifacts) artifact.grants = [];
}
