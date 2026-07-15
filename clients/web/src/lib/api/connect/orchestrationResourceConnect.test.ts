import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  ExportResourceRequestSchema,
  ExportResourceResponseSchema,
  GetResourcePlanRequestSchema,
  ListResourcesRequestSchema,
  ListResourcesResponseSchema,
  PlanResourceResponseSchema,
  ResourceOperation,
  ResourcePlanSchema,
  ResourceSchema,
  SourceFormat,
  ValidateResourceRequestSchema,
  ValidateResourceResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";

const methods = {
  validateResourceConnect: vi.fn(),
  planResourceConnect: vi.fn(),
  getResourceConnect: vi.fn(),
  listResourcesConnect: vi.fn(),
  exportResourceConnect: vi.fn(),
  getResourcePlanConnect: vi.fn(),
};

vi.mock("@/lib/wasm-core", () => ({
  getOrchestrationResourceService: () => methods,
}));

import {
  exportResource,
  getResource,
  getResourcePlan,
  listResources,
  planResource,
  validateResource,
} from "./orchestrationResourceConnect";

const target = {
  apiVersion: "orchestration.do/v1",
  kind: "WorkerTemplate",
  namespace: "default",
  name: "reviewer",
};

describe("orchestration resource Connect adapter", () => {
  beforeEach(() => {
    Object.values(methods).forEach((method) => method.mockReset());
  });

  it("encodes YAML validation and planning through the wasm service", async () => {
    methods.validateResourceConnect.mockResolvedValue(toBinary(
      ValidateResourceResponseSchema,
      create(ValidateResourceResponseSchema, {
        operation: ResourceOperation.CREATE,
        canonicalJson: new TextEncoder().encode('{"kind":"WorkerTemplate"}'),
      }),
    ));
    methods.planResourceConnect.mockResolvedValue(toBinary(
      PlanResourceResponseSchema,
      create(PlanResourceResponseSchema, {
        operation: ResourceOperation.CREATE,
        plan: { planId: "plan-1" },
      }),
    ));

    const document = {
      format: SourceFormat.YAML,
      content: "kind: WorkerTemplate\n",
    };
    await expect(validateResource("acme", document)).resolves.toMatchObject({
      operation: ResourceOperation.CREATE,
    });
    await expect(planResource("acme", document)).resolves.toMatchObject({
      plan: { planId: "plan-1" },
    });

    const request = fromBinary(
      ValidateResourceRequestSchema,
      methods.validateResourceConnect.mock.calls[0][0],
    );
    expect(request.orgSlug).toBe("acme");
    expect(request.source?.format).toBe(SourceFormat.YAML);
    expect(new TextDecoder().decode(request.source?.content)).toBe(
      "kind: WorkerTemplate\n",
    );
  });

  it("routes reads, export, and plan lookup without losing bigint values", async () => {
    const resource = create(ResourceSchema, { id: 42n, revision: 7n });
    const plan = create(ResourcePlanSchema, { planId: "plan-1" });
    methods.getResourceConnect.mockResolvedValue(toBinary(ResourceSchema, resource));
    methods.listResourcesConnect.mockResolvedValue(toBinary(
      ListResourcesResponseSchema,
      create(ListResourcesResponseSchema, { items: [resource], total: 1n }),
    ));
    methods.exportResourceConnect.mockResolvedValue(toBinary(
      ExportResourceResponseSchema,
      create(ExportResourceResponseSchema, {
        content: new TextEncoder().encode("kind: WorkerTemplate\n"),
      }),
    ));
    methods.getResourcePlanConnect.mockResolvedValue(
      toBinary(ResourcePlanSchema, plan),
    );

    await expect(getResource("acme", target)).resolves.toMatchObject({ id: 42n });
    await expect(listResources("acme", {
      kind: "WorkerTemplate",
      offset: 20,
      limit: 10,
    })).resolves.toMatchObject({ total: 1n, offset: 0, limit: 0 });
    const exported = await exportResource(
      "acme",
      target,
      SourceFormat.YAML,
      7n,
    );
    expect(new TextDecoder().decode(exported)).toBe("kind: WorkerTemplate\n");
    await expect(getResourcePlan("acme", "plan-1")).resolves.toMatchObject({
      planId: "plan-1",
    });

    const listRequest = fromBinary(
      ListResourcesRequestSchema,
      methods.listResourcesConnect.mock.calls[0][0],
    );
    expect(listRequest).toMatchObject({
      orgSlug: "acme",
      kind: "WorkerTemplate",
      offset: 20,
      limit: 10,
    });
    const exportRequest = fromBinary(
      ExportResourceRequestSchema,
      methods.exportResourceConnect.mock.calls[0][0],
    );
    expect(exportRequest).toMatchObject({
      orgSlug: "acme",
      format: SourceFormat.YAML,
      revision: 7n,
    });
    const planRequest = fromBinary(
      GetResourcePlanRequestSchema,
      methods.getResourcePlanConnect.mock.calls[0][0],
    );
    expect(planRequest).toMatchObject({ orgSlug: "acme", planId: "plan-1" });
  });
});
