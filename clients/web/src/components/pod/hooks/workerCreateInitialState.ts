import type { WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import type { WorkerCreateDraftState } from "./workerCreateDraft";

export function createInitialWorkerDraftState(
  initial?: Partial<WorkerSpecDraft>,
): WorkerCreateDraftState {
  return {
    instanceId: crypto.randomUUID(),
    step: 1,
    fillPrompt: "",
    generationModelResourceId: 0,
    draft: {
      model_resource_id: 0,
      tool_model_resource_ids: {},
      worker_type_slug: "",
      runtime_image_id: 0,
      placement_policy: "automatic",
      compute_target_id: 0,
      deployment_mode: "",
      resource_profile_id: 0,
      type_schema_version: 0,
      type_config_values: {},
      secret_refs: [],
      interaction_mode: "acp",
      automation_level: "autonomous",
      branch: "",
      skill_ids: [],
      knowledge_mounts: [],
      env_bundle_ids: [],
      instructions: "",
      initial_task: "",
      termination_policy: "manual",
      idle_timeout_minutes: 0,
      alias: "",
      options_revision: "",
      ...initial,
    },
    fill: { status: "idle" },
    fillRequestId: null,
    preflight: { status: "idle" },
    preflightRequestId: null,
    create: { status: "idle" },
  };
}
