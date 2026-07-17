import type {
  AuthorizationGrant,
  PermissionQuestionOption,
  PermissionRequest,
  SessionResource,
  SessionSnapshot,
  SupportCapabilities,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  PermissionRequestState,
  TerminalControlMode,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";

import type {
  AgentActivityItem,
  AgentConnectionStatus,
  AgentPermissionRequest,
  AgentWorkspaceCapabilities,
  TerminalResource,
} from "../contracts";
import {
  decodeStructuredPayload,
  formatUnsupported,
} from "./projectGeneratedSessionSnapshotPayload";
import { projectResourceStatus } from "./projectGeneratedSessionSnapshotStatuses";

const SEND_PROMPT = "send_prompt";
const INTERRUPT = "interrupt";
const CHANGE_CONFIGURATION = "change_configuration";
const RESOLVE_PERMISSION = "resolve_permission";

export function projectPermissions(
  requests: readonly PermissionRequest[],
): {
  evidence: AgentActivityItem[];
  permissions: AgentPermissionRequest[];
} {
  const permissions: AgentPermissionRequest[] = [];
  const evidence: AgentActivityItem[] = [];
  requests.forEach((request) => {
    if (request.state !== PermissionRequestState.PENDING) return;
    if (request.request.case === "approval") {
      permissions.push({
        id: request.permissionRequestId,
        kind: "approval",
        title: request.request.value.title,
        description: request.request.value.description ?? "",
      });
      return;
    }
    if (request.request.case === "questionnaire") {
      permissions.push({
        id: request.permissionRequestId,
        kind: "question",
        title: request.request.value.title,
        questions: request.request.value.questions.map((question) => ({
          id: question.questionId,
          prompt: question.prompt,
          header: question.header ?? "",
          options: question.options.map((option) => ({
            label: option.label,
            description: optionDescription(option),
          })),
          multiple: question.multiple,
          allowCustom: question.allowCustom,
          secret: question.secret,
        })),
      });
      return;
    }
    evidence.push({
      id: `permission:${request.permissionRequestId}`,
      kind: "system",
      title: "Unsupported permission request",
      detail:
        request.request.case === "unsupported"
          ? formatUnsupported(request.request.value)
          : "missing permission request payload",
      status: "failed",
    });
  });
  return { permissions, evidence };
}

export function projectTerminals(
  resources: readonly SessionResource[],
): TerminalResource[] {
  return resources.flatMap((resource) => {
    if (resource.resource.case !== "terminal") return [];
    const controlMode = resource.resource.value.controlMode;
    return [
      {
        id: resource.resourceId,
        label: resource.label,
        status: projectResourceStatus(resource.status),
        writable: resource.resource.value.writable,
        controlMode:
          controlMode === TerminalControlMode.SURFACE
            ? "surface"
            : controlMode === TerminalControlMode.HOST
              ? "host"
              : undefined,
      },
    ];
  });
}

export function projectCapabilities(
  snapshot: SessionSnapshot,
  connection: AgentConnectionStatus,
  terminals: readonly TerminalResource[],
): AgentWorkspaceCapabilities {
  const interactive = connection === "connected";
  const capabilities = snapshot.capabilities;
  return {
    sendMessage:
      interactive &&
      supportsCommand(capabilities, SEND_PROMPT, snapshot.grants, snapshot),
    interrupt:
      interactive &&
      supportsCommand(capabilities, INTERRUPT, snapshot.grants, snapshot),
    resolvePermission:
      interactive &&
      supportsCommand(
        capabilities,
        RESOLVE_PERMISSION,
        snapshot.grants,
        snapshot,
      ),
    updateConfiguration:
      interactive &&
      supportsCommand(
        capabilities,
        CHANGE_CONFIGURATION,
        snapshot.grants,
        snapshot,
      ),
    terminal:
      interactive &&
      terminals.length > 0 &&
      authorizedActions(
        capabilities?.terminalOperations ?? [],
        snapshot.grants,
        snapshot,
      ),
  };
}

function supportsCommand(
  capabilities: SupportCapabilities | undefined,
  semanticKey: string,
  grants: readonly AuthorizationGrant[],
  snapshot: SessionSnapshot,
): boolean {
  const descriptor = capabilities?.commandSchemas.find(
    (candidate) => candidate.semanticKey === semanticKey,
  );
  return Boolean(
    descriptor &&
      authorizedActions(descriptor.actions, grants, snapshot),
  );
}

function authorizedActions(
  actions: readonly string[],
  grants: readonly AuthorizationGrant[],
  snapshot: SessionSnapshot,
): boolean {
  if (actions.length === 0) return false;
  return grants.some(
    (grant) =>
      grantApplies(grant, snapshot) &&
      actions.some((action) => grant.actions.includes(action)),
  );
}

function grantApplies(
  grant: AuthorizationGrant,
  snapshot: SessionSnapshot,
): boolean {
  return (
    grant.sessionId === snapshot.sessionId &&
    (grant.minimumRevision === undefined ||
      snapshot.revision >= grant.minimumRevision) &&
    (grant.maximumRevision === undefined ||
      snapshot.revision <= grant.maximumRevision)
  );
}

function optionDescription(
  option: PermissionQuestionOption,
): string {
  const value = decodeStructuredPayload(option.value)?.text;
  return [option.description ?? "", value ? `Value: ${value}` : undefined]
    .filter(Boolean)
    .join("\n");
}
