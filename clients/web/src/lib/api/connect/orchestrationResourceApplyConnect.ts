import { create, toBinary } from "@bufbuild/protobuf";
import {
  ApplyBindingResourcePlanRequestSchema,
  ApplyExpertPlanRequestSchema,
  ApplyExpertPlanResponseSchema,
  ApplyPromptPlanRequestSchema,
  ApplyWorkerTemplatePlanRequestSchema,
  ApplyWorkerTemplatePlanResponseSchema,
  ApplyWorkflowPlanRequestSchema,
  ApplyWorkflowPlanResponseSchema,
  CreateGoalLoopFromPlanRequestSchema,
  CreateGoalLoopFromPlanResponseSchema,
  CreateWorkerFromPlanRequestSchema,
  CreateWorkerFromPlanResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_apply_pb";
import { ResourceSchema } from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import { getOrchestrationResourceService } from "@/lib/wasm-core";
import { callOrchestrationResource } from "./orchestrationResourceBinaryCall";

export function applyBindingResourcePlan(orgSlug: string, planId: string) {
  const request = create(ApplyBindingResourcePlanRequestSchema, {
    orgSlug,
    planId,
  });
  return callOrchestrationResource(
    ResourceSchema,
    toBinary(ApplyBindingResourcePlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService()
      .applyBindingResourcePlanConnect(bytes),
  );
}

export function applyWorkerTemplatePlan(orgSlug: string, planId: string) {
  const request = create(ApplyWorkerTemplatePlanRequestSchema, {
    orgSlug,
    planId,
  });
  return callOrchestrationResource(
    ApplyWorkerTemplatePlanResponseSchema,
    toBinary(ApplyWorkerTemplatePlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService()
      .applyWorkerTemplatePlanConnect(bytes),
  );
}

export function createWorkerFromPlan(orgSlug: string, planId: string) {
  const request = create(CreateWorkerFromPlanRequestSchema, {
    orgSlug,
    planId,
  });
  return callOrchestrationResource(
    CreateWorkerFromPlanResponseSchema,
    toBinary(CreateWorkerFromPlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService()
      .createWorkerFromPlanConnect(bytes),
  );
}

export function createGoalLoopFromPlan(orgSlug: string, planId: string) {
  const request = create(CreateGoalLoopFromPlanRequestSchema, {
    orgSlug,
    planId,
  });
  return callOrchestrationResource(
    CreateGoalLoopFromPlanResponseSchema,
    toBinary(CreateGoalLoopFromPlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService()
      .createGoalLoopFromPlanConnect(bytes),
  );
}

export function applyPromptPlan(orgSlug: string, planId: string) {
  const request = create(ApplyPromptPlanRequestSchema, { orgSlug, planId });
  return callOrchestrationResource(
    ResourceSchema,
    toBinary(ApplyPromptPlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService().applyPromptPlanConnect(bytes),
  );
}

export function applyExpertPlan(orgSlug: string, planId: string) {
  const request = create(ApplyExpertPlanRequestSchema, { orgSlug, planId });
  return callOrchestrationResource(
    ApplyExpertPlanResponseSchema,
    toBinary(ApplyExpertPlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService().applyExpertPlanConnect(bytes),
  );
}

export function applyWorkflowPlan(orgSlug: string, planId: string) {
  const request = create(ApplyWorkflowPlanRequestSchema, { orgSlug, planId });
  return callOrchestrationResource(
    ApplyWorkflowPlanResponseSchema,
    toBinary(ApplyWorkflowPlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService()
      .applyWorkflowPlanConnect(bytes),
  );
}
