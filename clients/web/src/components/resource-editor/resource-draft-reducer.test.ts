import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  PlanResourceResponseSchema,
  ResourceOperation,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import {
  createResourceDraftState,
  resourceDraftCanApply,
  resourceDraftReducer,
} from "./resource-draft-reducer";
import { createWorkerTemplateDraft } from "./worker-template-draft";

describe("resource draft reducer", () => {
  it("invalidates a ready plan whenever the form draft changes", () => {
    const initial = createResourceDraftState(createWorkerTemplateDraft("acme"));
    const planning = resourceDraftReducer(initial, {
      type: "plan_loading",
      requestId: "plan-request",
      version: 0,
    });
    const ready = resourceDraftReducer(planning, {
      type: "plan_succeeded",
      requestId: "plan-request",
      version: 0,
      response: create(PlanResourceResponseSchema, {
        operation: ResourceOperation.CREATE,
        plan: { planId: "plan-1", expiresAt: "2099-01-01T00:00:00Z" },
      }),
    });

    expect(resourceDraftCanApply(ready)).toBe(true);
    const changed = resourceDraftReducer(ready, {
      type: "replace_draft",
      draft: {
        ...ready.draft,
        metadata: { ...ready.draft.metadata, displayName: "Reviewer" },
      },
    });
    expect(changed.version).toBe(1);
    expect(changed.plan.status).toBe("idle");
    expect(resourceDraftCanApply(changed)).toBe(false);
  });

  it("expires a ready plan before apply", () => {
    const initial = createResourceDraftState(createWorkerTemplateDraft("acme"));
    const planning = resourceDraftReducer(initial, {
      type: "plan_loading",
      requestId: "plan-request",
      version: 0,
    });
    const ready = resourceDraftReducer(planning, {
      type: "plan_succeeded",
      requestId: "plan-request",
      version: 0,
      response: create(PlanResourceResponseSchema, {
        plan: { planId: "plan-1", expiresAt: "2099-01-01T00:00:00Z" },
      }),
    });

    const expired = resourceDraftReducer(ready, { type: "plan_expired" });
    expect(expired.plan.status).toBe("expired");
    expect(resourceDraftCanApply(expired)).toBe(false);
  });

  it("ignores a plan response for an older draft version", () => {
    const initial = createResourceDraftState(createWorkerTemplateDraft("acme"));
    const planning = resourceDraftReducer(initial, {
      type: "plan_loading",
      requestId: "old-request",
      version: 0,
    });
    const changed = resourceDraftReducer(planning, {
      type: "replace_draft",
      draft: {
        ...planning.draft,
        metadata: { ...planning.draft.metadata, name: "new-reviewer" },
      },
    });
    const stale = resourceDraftReducer(changed, {
      type: "plan_succeeded",
      requestId: "old-request",
      version: 0,
      response: create(PlanResourceResponseSchema, {
        plan: { planId: "stale-plan" },
      }),
    });

    expect(stale).toBe(changed);
  });

  it("keeps invalid YAML text without applying the previous typed draft", () => {
    const initial = createResourceDraftState(createWorkerTemplateDraft("acme"));
    const changed = resourceDraftReducer(initial, {
      type: "source_changed",
      text: "password: secret-value\n  broken",
    });
    const invalid = resourceDraftReducer(changed, {
      type: "source_invalid",
      error: "YAML syntax error at line 2",
      version: changed.version,
    });

    expect(invalid.source.text).toContain("secret-value");
    expect(invalid.source.error).toBe("YAML syntax error at line 2");
    expect(invalid.source.dirty).toBe(true);
    expect(invalid.draft).toBe(initial.draft);
    expect(resourceDraftCanApply(invalid)).toBe(false);
  });

  it("replaces the typed draft after YAML parses successfully", () => {
    const initial = createResourceDraftState(createWorkerTemplateDraft("acme"));
    const changed = resourceDraftReducer(initial, {
      type: "source_changed",
      text: "kind: WorkerTemplate",
    });
    const parsedDraft = {
      ...initial.draft,
      metadata: { ...initial.draft.metadata, name: "yaml-reviewer" },
    };
    const parsed = resourceDraftReducer(changed, {
      type: "source_parsed",
      draft: parsedDraft,
      version: changed.version,
    });

    expect(parsed.draft.metadata.name).toBe("yaml-reviewer");
    expect(parsed.source.error).toBeNull();
    expect(parsed.source.dirty).toBe(false);
  });
});
