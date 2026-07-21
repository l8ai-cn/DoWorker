import type { SessionSnapshot } from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";

import type {
  AgentConnectionStatus,
  AgentSessionSnapshot,
} from "../contracts";
import {
  createArtifactCatalog,
  projectArtifactDescriptor,
} from "./projectGeneratedSessionSnapshotArtifacts";
import { formatAgentError } from "./projectGeneratedSessionSnapshotPayload";
import { projectConfiguration } from "./projectGeneratedSessionSnapshotConfiguration";
import {
  projectCapabilities,
  projectPermissions,
  projectTerminals,
} from "./projectGeneratedSessionSnapshotState";
import { projectSessionStatus } from "./projectGeneratedSessionSnapshotStatuses";
import { projectTimeline } from "./projectGeneratedSessionSnapshotTimeline";

export interface GeneratedSessionSnapshotProjectionOptions {
  title: string;
  agentLabel: string;
  connection: AgentConnectionStatus;
  interactionMode: "acp" | "pty";
  hasOlderItems: boolean;
}

export function projectGeneratedSessionSnapshot(
  snapshot: SessionSnapshot,
  options: GeneratedSessionSnapshotProjectionOptions,
): AgentSessionSnapshot {
  const catalog = createArtifactCatalog(snapshot.artifacts);
  const permissionMap = new Map(
    snapshot.permissionRequests.map((request) => [
      request.permissionRequestId,
      request,
    ]),
  );
  const timeline = projectTimeline(snapshot.history, catalog, permissionMap);
  const permissionProjection = projectPermissions(snapshot.permissionRequests);
  const representedArtifacts = new Set(
    timeline.items.flatMap((item) =>
      item.kind === "artifact" ? [item.artifactId] : [],
    ),
  );
  const catalogItems = snapshot.artifacts
    .filter((artifact) => !representedArtifacts.has(artifact.artifactId))
    .map((artifact) =>
      projectArtifactDescriptor(
        artifact,
        `artifact:${artifact.artifactId}`,
      ),
    );
  const terminals = projectTerminals(snapshot.resources);
  const configuration = projectConfiguration(snapshot);
  const error =
    formatAgentError(snapshot.error) ??
    (snapshot.status === SessionStatus.RESYNC_REQUIRED
      ? "[resync_required] Session requires resynchronization."
      : null);
  return {
    sessionId: snapshot.sessionId,
    title: options.title,
    agentLabel: options.agentLabel,
    status: projectSessionStatus(snapshot.status),
    connection: options.connection,
    interactionMode: options.interactionMode,
    ...(timeline.latestUserCommandId
      ? { latestUserCommandId: timeline.latestUserCommandId }
      : {}),
    capabilities: projectCapabilities(
      snapshot,
      options.connection,
      terminals,
    ),
    ...(configuration ? { configuration } : {}),
    items: [
      ...timeline.items,
      ...catalogItems,
      ...permissionProjection.evidence,
    ],
    plan: timeline.plan,
    permissions: permissionProjection.permissions,
    terminals,
    hasOlderItems: options.hasOlderItems,
    error,
  };
}
