import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  FillWorkerDraftRequestSchema,
  FillWorkerDraftResponseSchema,
  PreflightWorkerRequestSchema,
  PreflightWorkerResponseSchema,
} from "@proto/pod/v1/worker_creation_pb";

import { getPodService } from "@/lib/wasm-core";
import { workerDraftFromProto, workerDraftToProto } from "./podWorkerDraftProto";
import type {
  WorkerDraftFillResult,
  WorkerPreflightIssue,
  WorkerPreflightResult,
  WorkerSpecDraft,
} from "./podWorkerCreationTypes";

export async function preflightWorker(
  orgSlug: string,
  draft: WorkerSpecDraft,
): Promise<WorkerPreflightResult> {
  const request = create(PreflightWorkerRequestSchema, {
    orgSlug,
    draft: workerDraftToProto(draft),
  });
  const responseBytes = await getPodService().preflight_worker_connect(
    toBinary(PreflightWorkerRequestSchema, request),
  );
  const response = fromBinary(
    PreflightWorkerResponseSchema,
    new Uint8Array(responseBytes),
  );
  return {
    issues: response.issues.map(preflightIssueFromProto),
    resolved_spec_json: response.resolvedSpecJson,
    options_revision: response.optionsRevision,
  };
}

export async function fillWorkerDraft(
  orgSlug: string,
  prompt: string,
  currentDraft?: WorkerSpecDraft,
): Promise<WorkerDraftFillResult> {
  const request = create(FillWorkerDraftRequestSchema, {
    orgSlug,
    prompt,
    currentDraft: currentDraft ? workerDraftToProto(currentDraft) : undefined,
  });
  const responseBytes = await getPodService().fill_worker_draft_connect(
    toBinary(FillWorkerDraftRequestSchema, request),
  );
  const response = fromBinary(
    FillWorkerDraftResponseSchema,
    new Uint8Array(responseBytes),
  );
  if (!response.draft) {
    throw new Error("worker draft response is missing draft");
  }
  return {
    draft: workerDraftFromProto(response.draft),
    issues: response.issues.map(preflightIssueFromProto),
  };
}

function preflightIssueFromProto(issue: {
  code: string;
  field: string;
  message: string;
  severity: string;
}): WorkerPreflightIssue {
  return {
    code: issue.code,
    field: issue.field,
    message: issue.message,
    severity: issue.severity,
  };
}

export { listWorkerCreateOptions } from "./podWorkerCreateOptionsConnect";
export { workerDraftFromProto, workerDraftToProto };
export type {
  WorkerCreateOptions,
  WorkerCreateOptionsFilter,
  WorkerConfigDocumentBinding,
  WorkerConfigDocumentRequirement,
  WorkerCredentialRequirement,
  WorkerDraftFillResult,
  WorkerPreflightIssue,
  WorkerPreflightResult,
  WorkerResourceRequest,
  WorkerSpecDraft,
  WorkerToolModelRequirement,
  WorkerTypeOption,
} from "./podWorkerCreationTypes";
