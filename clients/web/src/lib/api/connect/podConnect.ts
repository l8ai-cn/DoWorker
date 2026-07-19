// Connect-RPC adapter for proto.pod.v1.PodService.
//
// Encodes requests via @bufbuild/protobuf .toBinary(), passes the Uint8Array
// to the wasm bridge (binary in / binary out per conventions §2.5), decodes
// responses via .fromBinary(). No JSON intermediate.
//
// Returns snake_case web shapes (PodData, PodConnectionInfo) so call sites
// don't have to switch wire-camelCase off the proto generated types — same
// pattern as runnerConnect.ts during the dual-track migration window.

import {
  CreatePodRequestSchema,
  CreatePodResponseSchema,
} from "@proto/pod/v1/pod_pb";
import { create, toBinary, fromBinary } from "@bufbuild/protobuf";
// Shared proto->PodData projection. Aliased to the historical fromProtoPod
// name and re-exported for cross-file use (stores/pod, podProtoMap, facade).
import { podToCache as fromProtoPod } from "@/lib/api/projections";
import { getPodService } from "@/lib/wasm-core";
import type { PodData } from "@/lib/api/facade/pod";
import { workerDraftToProto } from "./podWorkerDraftProto";
import type { WorkerSpecDraft } from "./podWorkerCreationTypes";

export { fromProtoPod };

interface CreatePodBaseInput {
  ticket_slug?: string;
  cols?: number;
  rows?: number;
}

export interface FreshCreatePodInput extends CreatePodBaseInput {
  worker_spec: WorkerSpecDraft;
}

export interface ResumeCreatePodInput extends CreatePodBaseInput {
  source_pod_key: string;
  resume_agent_session?: boolean;
}

export type CreatePodInput = FreshCreatePodInput | ResumeCreatePodInput;

export async function createPod(
  orgSlug: string,
  input: CreatePodInput,
): Promise<{ pod: PodData; warning?: string }> {
  const base = {
    orgSlug,
    ticketSlug: input.ticket_slug,
    cols: input.cols ?? 0,
    rows: input.rows ?? 0,
  };
  const req = create(
    CreatePodRequestSchema,
    "worker_spec" in input
      ? { ...base, workerSpec: workerDraftToProto(input.worker_spec) }
      : {
          ...base,
          sourcePodKey: input.source_pod_key,
          resumeAgentSession: input.resume_agent_session,
        },
  );
  const bytes = toBinary(CreatePodRequestSchema, req);
  const respBytes = await getPodService().create_pod_connect(bytes);
  const resp = fromBinary(CreatePodResponseSchema, new Uint8Array(respBytes));
  return { pod: fromProtoPod(resp.pod!), warning: resp.warning };
}

export async function wakePod(
  orgSlug: string,
  sourcePodKey: string,
): Promise<{ pod: PodData; warning?: string }> {
  return createPod(orgSlug, {
    source_pod_key: sourcePodKey,
    resume_agent_session: true,
  });
}

export {
  getMobileAccessDescriptor,
  getPodConnection,
  listPodsByTicket,
  sendPodPrompt,
  terminatePod,
  updatePodAlias,
  updatePodPerpetual,
  updatePodPreviewConfig,
  type MobileAccessDescriptor,
  type PodConnectionInfo,
} from "./podControlConnect";

export { getPod, listPods, listPodsRaw } from "./podQueryConnect";
