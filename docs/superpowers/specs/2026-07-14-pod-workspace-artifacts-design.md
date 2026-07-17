# Pod Workspace Artifacts Design

## Goal

Let an ACP Worker expose generated files such as Seedance MP4 output directly
in the existing Agent workspace, without requiring an `agent_sessions` row or
an agent-specific output event.

## Current Problem

The Agent UI already renders `AgentArtifactItem` values and includes inline
video playback. ACP tool calls can also project `fileChange` entries into
workspace artifacts.

The remaining path is incomplete:

- standalone ACP Workers do not necessarily have an `agent_sessions` row;
- the current workspace artifact loader resolves the Worker through a session;
- files written by shell commands or downloaded by a provider do not
  necessarily emit a `fileChange` tool event.

As a result, a valid MP4 can exist in the Worker workspace but remain invisible
and unreadable in the platform.

## Decision

Add a Pod-scoped, read-only workspace artifact API and let
`WebAcpSessionRuntime` discover deliverable files from that API.

This keeps ownership with the Worker resource that owns the sandbox. It also
works for every ACP agent and every file-producing tool, rather than adding a
Seedance-only branch or inventing a non-standard ACP event.

## Backend API

The organization-scoped Pod routes gain two read-only endpoints:

```text
GET /pods/:key/resources/workspace/changes
GET /pods/:key/resources/workspace/filesystem/*filepath
```

Both endpoints:

- load the Pod by `pod_key`;
- enforce the existing `PodPolicy.AllowRead` authorization;
- require a connected Runner and a configured sandbox;
- execute the existing `SandboxFsService` operation against the Pod's Runner;
- preserve existing sandbox path containment and binary read limits.

`changes` returns the existing filesystem change wire shape. For standalone
workspaces, every regular file is reported as created. The filesystem endpoint
returns the existing UTF-8 or base64 file-content wire shape.

No write endpoint is added. Artifact viewing must not expand Worker mutation
permissions.

## Frontend Data Flow

The Web ACP adapter adds two dependencies:

```text
listWorkspaceArtifacts(podKey) -> AgentArtifactItem[]
loadWorkspaceArtifact(podKey, path) -> Blob
```

The list call reads Pod workspace changes and passes them through the existing
`workspaceFileArtifacts` filter. Only supported deliverable extensions become
timeline artifacts.

`WebAcpSessionRuntime` refreshes artifacts:

1. once after the Relay subscription opens;
2. whenever the ACP session transitions from a non-idle state to idle.

The runtime deduplicates concurrent refreshes and keeps the last successful
artifact list. A refresh failure is surfaced through the runtime error state;
it does not fabricate an empty successful result.

The projected snapshot merges tool-derived artifacts and discovered workspace
artifacts by `artifactId`. The existing `ArtifactCard` loads the Blob and
renders MP4 files with the existing inline `<video controls>` surface.

## Error Handling

- Missing or unauthorized Pod: `404` or `403`.
- Runner unavailable: `503`.
- Sandbox filesystem operation failure: existing `400`, `404`, or `502`
  mapping.
- Oversized or truncated preview file: frontend displays the existing artifact
  error card.
- Unsupported file extension: omitted from artifact discovery.
- Artifact deleted between listing and reading: artifact card displays the
  returned read error.

There is no fallback to a session route. A Pod-scoped Worker artifact must be
resolved through the Pod-scoped contract.

## Verification

Backend tests prove:

- Pod read authorization is enforced;
- the correct Runner and `pod_key` are used;
- changes and binary file content use the established wire contracts;
- disconnected Runners fail explicitly.

Frontend tests prove:

- Pod workspace API responses become MP4 artifacts;
- artifact discovery runs on open and after a busy-to-idle transition;
- duplicate artifacts are removed;
- MP4 Blob loading uses the Pod endpoint, not a session lookup.

Browser verification creates a deterministic local MP4 inside a Seedance
Expert Worker workspace, reloads the Worker, and confirms:

- a video artifact card is visible;
- the `<video>` element has loaded metadata;
- the media request succeeds;
- no browser console error is introduced.

Real Ark generation remains a separate final verification requiring a rotated,
authorized API key.
