import {
  create,
  toBinary,
} from "@bufbuild/protobuf";
import {
  ExportResourceRequestSchema,
  ExportResourceResponseSchema,
  GetResourcePlanRequestSchema,
  GetResourceRequestSchema,
  ListResourcesRequestSchema,
  ListResourcesResponseSchema,
  PlanResourceRequestSchema,
  PlanResourceResponseSchema,
  ResourcePlanSchema,
  ResourceSchema,
  SourceFormat,
  ValidateResourceRequestSchema,
  ValidateResourceResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { getOrchestrationResourceService } from "@/lib/wasm-core";
import { callOrchestrationResource as call } from "./orchestrationResourceBinaryCall";

export interface ResourceDocument {
  format: SourceFormat;
  content: string | Uint8Array;
}

export interface ResourceTargetInput {
  apiVersion: string;
  kind: string;
  namespace: string;
  name: string;
}

export interface ResourceListInput {
  kind?: string;
  offset?: number;
  limit?: number;
}

function source(document: ResourceDocument) {
  return {
    format: document.format,
    content: typeof document.content === "string"
      ? new TextEncoder().encode(document.content)
      : document.content,
  };
}

function target(input: ResourceTargetInput) {
  return {
    typeMeta: { apiVersion: input.apiVersion, kind: input.kind },
    namespace: input.namespace,
    name: input.name,
  };
}

export function validateResource(orgSlug: string, document: ResourceDocument) {
  const request = create(ValidateResourceRequestSchema, {
    orgSlug,
    source: source(document),
  });
  return call(
    ValidateResourceResponseSchema,
    toBinary(ValidateResourceRequestSchema, request),
    (bytes) => getOrchestrationResourceService().validateResourceConnect(bytes),
  );
}

export function planResource(orgSlug: string, document: ResourceDocument) {
  const request = create(PlanResourceRequestSchema, {
    orgSlug,
    source: source(document),
  });
  return call(
    PlanResourceResponseSchema,
    toBinary(PlanResourceRequestSchema, request),
    (bytes) => getOrchestrationResourceService().planResourceConnect(bytes),
  );
}

export function getResource(orgSlug: string, input: ResourceTargetInput) {
  const request = create(GetResourceRequestSchema, {
    orgSlug,
    target: target(input),
  });
  return call(
    ResourceSchema,
    toBinary(GetResourceRequestSchema, request),
    (bytes) => getOrchestrationResourceService().getResourceConnect(bytes),
  );
}

export function listResources(orgSlug: string, input: ResourceListInput = {}) {
  const request = create(ListResourcesRequestSchema, { orgSlug, ...input });
  return call(
    ListResourcesResponseSchema,
    toBinary(ListResourcesRequestSchema, request),
    (bytes) => getOrchestrationResourceService().listResourcesConnect(bytes),
  );
}

export async function exportResource(
  orgSlug: string,
  input: ResourceTargetInput,
  format: SourceFormat,
  revision?: bigint,
) {
  const request = create(ExportResourceRequestSchema, {
    orgSlug,
    target: target(input),
    format,
    revision,
  });
  const response = await call(
    ExportResourceResponseSchema,
    toBinary(ExportResourceRequestSchema, request),
    (bytes) => getOrchestrationResourceService().exportResourceConnect(bytes),
  );
  return response.content;
}

export function getResourcePlan(orgSlug: string, planId: string) {
  const request = create(GetResourcePlanRequestSchema, { orgSlug, planId });
  return call(
    ResourcePlanSchema,
    toBinary(GetResourcePlanRequestSchema, request),
    (bytes) => getOrchestrationResourceService().getResourcePlanConnect(bytes),
  );
}
