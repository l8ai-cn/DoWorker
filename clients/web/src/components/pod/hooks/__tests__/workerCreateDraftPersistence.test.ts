import { beforeEach, describe, expect, it } from "vitest";
import type { WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import {
  clearWorkerCreateDraft,
  loadWorkerCreateDraft,
  persistWorkerCreateDraft,
} from "../workerCreateDraftPersistence";

describe("workerCreateDraftPersistence", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it("restores the wizard state without storing sensitive config values", () => {
    const draft = {
      worker_type_slug: "minimax-cli",
      type_config_values: {
        approval_mode: "never",
        api_key: "plaintext-secret",
        access_token: "another-secret",
      },
      skill_ids: [12],
    } as Partial<WorkerSpecDraft>;

    persistWorkerCreateDraft("test-org", {
      step: 3,
      fillPrompt: "Review this repository",
      draft,
    });

    expect(loadWorkerCreateDraft("test-org")).toEqual({
      step: 3,
      fillPrompt: "Review this repository",
      draft: {
        ...draft,
        type_config_values: { approval_mode: "never" },
      },
    });
    expect(sessionStorage.getItem("agentcloud.worker-create-draft.v1:test-org"))
      .not.toContain("plaintext-secret");
  });

  it("clears the draft after the flow is completed or cancelled", () => {
    persistWorkerCreateDraft("test-org", {
      step: 1,
      fillPrompt: "",
      draft: {},
    });

    clearWorkerCreateDraft("test-org");

    expect(loadWorkerCreateDraft("test-org")).toBeNull();
  });
});
