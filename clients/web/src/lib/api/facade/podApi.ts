import { readCurrentOrg } from "@/stores/auth";
import { createPod, terminatePod, type CreatePodInput } from "../connect/podConnect";
import {
  fillWorkerDraft,
  listWorkerCreateOptions,
  preflightWorker,
  type WorkerCreateOptionsFilter,
  type WorkerSpecDraft,
} from "../connect/podWorkerCreationConnect";

function orgSlug(): string {
  return readCurrentOrg()?.slug ?? "";
}

export const podApi = {
  create: async (data: CreatePodInput) => {
    return createPod(orgSlug(), data);
  },
  terminate: async (podKey: string) => {
    await terminatePod(orgSlug(), podKey);
  },
  listWorkerCreateOptions: async (filter: WorkerCreateOptionsFilter = {}) => {
    return listWorkerCreateOptions(orgSlug(), filter);
  },
  preflightWorker: async (draft: WorkerSpecDraft) => {
    return preflightWorker(orgSlug(), draft);
  },
  fillWorkerDraft: async (prompt: string, currentDraft?: WorkerSpecDraft) => {
    return fillWorkerDraft(orgSlug(), prompt, currentDraft);
  },
};
