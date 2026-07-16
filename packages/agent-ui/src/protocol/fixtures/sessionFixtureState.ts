import { create } from "@bufbuild/protobuf";

import {
  ArtifactDescriptorSchema,
  ArtifactStatus,
  VideoStage,
  type ArtifactDescriptor,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import {
  CommandReceiptSchema,
  CommandReceiptState,
  PermissionDecision,
  type CommandReceipt,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  PermissionRequestSchema,
  SessionResourceSchema,
  type PermissionRequest,
  type SessionResource,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  PermissionRequestState,
  SessionResourceStatus,
  TerminalControlMode,
  TerminalLeaseState,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";

export function createFixtureArtifacts(): ArtifactDescriptor[] {
  return [
    create(ArtifactDescriptorSchema, {
      artifactId: "image-1",
      revision: 1n,
      filename: "result.png",
      mediaType: "image/png",
      status: ArtifactStatus.READY,
      representations: [
        {
          representationId: "image-source",
          revision: 1n,
          mediaType: "image/png",
          status: ArtifactStatus.READY,
          byteSize: 4_294_967_297n,
        },
      ],
      revisions: [{ revision: 1n, representationIds: ["image-source"] }],
    }),
    create(ArtifactDescriptorSchema, {
      artifactId: "video-1",
      revision: 2n,
      filename: "result.mp4",
      mediaType: "video/mp4",
      status: ArtifactStatus.READY,
      durationMillis: 65_535n,
      representations: [
        {
          representationId: "video-playable",
          revision: 2n,
          mediaType: "video/mp4",
          status: ArtifactStatus.READY,
        },
      ],
      revisions: [{ revision: 2n, representationIds: ["video-playable"] }],
      manifest: {
        manifest: {
          case: "video",
          value: {
            stage: VideoStage.READY,
            durationMillis: 65_535n,
            playableRepresentationId: "video-playable",
          },
        },
      },
    }),
    create(ArtifactDescriptorSchema, {
      artifactId: "deck-1",
      revision: 3n,
      filename: "workbench.pptx",
      mediaType:
        "application/vnd.openxmlformats-officedocument.presentationml.presentation",
      status: ArtifactStatus.READY,
      representations: [
        {
          representationId: "deck-source",
          revision: 3n,
          mediaType:
            "application/vnd.openxmlformats-officedocument.presentationml.presentation",
          status: ArtifactStatus.READY,
        },
      ],
      revisions: [
        {
          revision: 3n,
          baseRevision: 2n,
          representationIds: ["deck-source"],
          digest: "sha256:deck-revision-3",
        },
      ],
      manifest: {
        manifest: {
          case: "presentation",
          value: {
            deckRevision: 3n,
            slides: [
              {
                slideId: "slide-1",
                position: 1,
                title: "Workbench",
                pageRepresentationId: "deck-source",
              },
            ],
            versions: [{ versionId: "version-3", revision: 3n, label: "Current" }],
            selectedVersionId: "version-3",
          },
        },
      },
    }),
  ];
}

export function createFixtureCommandReceipts(): CommandReceipt[] {
  return [
    create(CommandReceiptSchema, {
      sessionId: "session-lossless-1",
      commandId: "running-command-1",
      state: CommandReceiptState.RUNNING,
      payloadDigest: "sha256:running-command",
      receivedAt: "2026-07-16T00:00:00Z",
      updatedAt: "2026-07-16T00:00:01Z",
    }),
    create(CommandReceiptSchema, {
      sessionId: "session-lossless-1",
      commandId: "terminal-command-1",
      state: CommandReceiptState.SUCCEEDED,
      payloadDigest: "sha256:terminal-command",
      resultingRevision: 9_007_199_254_740_993n,
      receivedAt: "2026-07-16T00:00:02Z",
      updatedAt: "2026-07-16T00:00:03Z",
    }),
  ];
}

export function createFixturePermissionRequests(): PermissionRequest[] {
  return [
    create(PermissionRequestSchema, {
      permissionRequestId: "permission-1",
      state: PermissionRequestState.PENDING,
      commandId: "running-command-1",
      artifactId: "deck-1",
      representationId: "deck-source",
      artifactRevision: 3n,
      actionType: "presentation.publish",
      requiredGrantActions: ["artifact.publish"],
      createdAt: "2026-07-16T00:00:01Z",
      request: {
        case: "approval",
        value: {
          title: "Publish presentation",
          description: "Allow the agent to publish revision 3.",
        },
      },
    }),
    create(PermissionRequestSchema, {
      permissionRequestId: "permission-2",
      state: PermissionRequestState.RESOLVED,
      commandId: "terminal-command-1",
      resolution: {
        permissionRequestId: "permission-2",
        decision: PermissionDecision.ACCEPT,
        resolvedAt: "2026-07-16T00:00:04Z",
      },
    }),
  ];
}

export function createFixtureResources(): SessionResource[] {
  return [
    create(SessionResourceSchema, {
      resourceId: "terminal-1",
      label: "Workbench terminal",
      status: SessionResourceStatus.READY,
      resource: {
        case: "terminal",
        value: {
          writable: true,
          controlMode: TerminalControlMode.SHARED,
          lease: {
            leaseId: "terminal-lease-1",
            holder: "web-client-1",
            state: TerminalLeaseState.ACTIVE,
            expiresAt: "2026-07-16T00:05:00Z",
            fencingEpoch: 4_294_967_297n,
          },
        },
      },
    }),
  ];
}
