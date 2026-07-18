import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  ApplyExpertPlanRequestSchema,
  ApplyExpertPlanResponseSchema,
  ApplyPromptPlanRequestSchema,
  ApplyWorkerTemplatePlanResponseSchema,
  ApplyWorkflowPlanRequestSchema,
  ApplyWorkflowPlanResponseSchema,
  CreateGoalLoopFromPlanRequestSchema,
  CreateGoalLoopFromPlanResponseSchema,
  CreateWorkerFromPlanRequestSchema,
  CreateWorkerFromPlanResponseSchema,
  EnvironmentBundlePurpose,
  ExportResourceResponseSchema,
  GetResourceCapabilitiesRequestSchema,
  GetResourceCapabilitiesResponseSchema,
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
  getResourceCapabilitiesConnect: vi.fn(),
  listResourcesConnect: vi.fn(),
  exportResourceConnect: vi.fn(),
  getResourcePlanConnect: vi.fn(),
  applyBindingResourcePlanConnect: vi.fn(),
  applyWorkerTemplatePlanConnect: vi.fn(),
  createWorkerFromPlanConnect: vi.fn(),
  createGoalLoopFromPlanConnect: vi.fn(),
  applyPromptPlanConnect: vi.fn(),
  applyExpertPlanConnect: vi.fn(),
  applyWorkflowPlanConnect: vi.fn(),
};

vi.mock("@/lib/wasm-core", () => ({
  getOrchestrationResourceService: () => methods,
}));

import {
  applyBindingResourcePlan,
  applyExpertPlan,
  applyPromptPlan,
  applyWorkerTemplatePlan,
  applyWorkflowPlan,
  createGoalLoopFromPlan,
  createWorkerFromPlan,
  exportResource,
  getResource,
  getResourceCapabilities,
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

    const document = { format: SourceFormat.YAML, content: "kind: WorkerTemplate\n" };
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

  it("routes resource reads, export, and plan lookup without losing bigint values", async () => {
    const resource = create(ResourceSchema, { id: 42n, revision: 7n });
    const plan = create(ResourcePlanSchema, { planId: "plan-1" });
    methods.getResourceConnect.mockResolvedValue(toBinary(ResourceSchema, resource));
    methods.getResourceCapabilitiesConnect.mockResolvedValue(toBinary(
      GetResourceCapabilitiesResponseSchema,
      create(GetResourceCapabilitiesResponseSchema, {
        target,
        capabilities: {
          exists: true,
          canViewSource: true,
          canReference: true,
          canPlan: false,
        },
      }),
    ));
    methods.listResourcesConnect.mockResolvedValue(toBinary(
      ListResourcesResponseSchema,
      create(ListResourcesResponseSchema, {
        items: [resource],
        total: 1n,
        appliedEnvironmentBundleFilter: {
          purpose: EnvironmentBundlePurpose.CREDENTIAL,
          workerType: "do-agent",
          targetName: "DO_API_KEY",
        },
      }),
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
    await expect(getResourceCapabilities("acme", target)).resolves.toMatchObject({
      capabilities: {
        exists: true,
        canViewSource: true,
        canReference: true,
        canPlan: false,
      },
    });
    await expect(listResources("acme", {
      kind: "EnvironmentBundle",
      offset: 20,
      limit: 10,
      environmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType: "do-agent",
        targetName: "DO_API_KEY",
      },
    })).resolves.toMatchObject({
      total: 1n,
      offset: 0,
      limit: 0,
      appliedEnvironmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType: "do-agent",
        targetName: "DO_API_KEY",
      },
    });
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

    const planRequest = fromBinary(
      GetResourcePlanRequestSchema,
      methods.getResourcePlanConnect.mock.calls[0][0],
    );
    expect(planRequest).toMatchObject({ orgSlug: "acme", planId: "plan-1" });
    const capabilitiesRequest = fromBinary(
      GetResourceCapabilitiesRequestSchema,
      methods.getResourceCapabilitiesConnect.mock.calls[0][0],
    );
    expect(capabilitiesRequest).toMatchObject({
      orgSlug: "acme",
      target: {
        typeMeta: {
          apiVersion: target.apiVersion,
          kind: target.kind,
        },
        namespace: target.namespace,
        name: target.name,
      },
    });
    const listRequest = fromBinary(
      ListResourcesRequestSchema,
      methods.listResourcesConnect.mock.calls[0][0],
    );
    expect(listRequest).toMatchObject({
      orgSlug: "acme",
      kind: "EnvironmentBundle",
      offset: 20,
      limit: 10,
      environmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType: "do-agent",
        targetName: "DO_API_KEY",
      },
    });
  });

  it("keeps typed apply paths explicit", async () => {
    methods.applyBindingResourcePlanConnect.mockResolvedValue(toBinary(
      ResourceSchema,
      create(ResourceSchema, { id: 43n }),
    ));
    methods.applyWorkerTemplatePlanConnect.mockResolvedValue(toBinary(
      ApplyWorkerTemplatePlanResponseSchema,
      create(ApplyWorkerTemplatePlanResponseSchema, {
        resource: { id: 44n },
        workerSpecSnapshotId: 91n,
      }),
    ));
    methods.createWorkerFromPlanConnect.mockResolvedValue(toBinary(
      CreateWorkerFromPlanResponseSchema,
      create(CreateWorkerFromPlanResponseSchema, {
        resource: { id: 48n },
        launchId: 71n,
        podId: 73n,
        podKey: "7-standalone-12345678",
        workerSpecSnapshotId: 91n,
        resourceRevision: 3n,
        runnerId: 11n,
      }),
    ));
    methods.createGoalLoopFromPlanConnect.mockResolvedValue(toBinary(
      CreateGoalLoopFromPlanResponseSchema,
      create(CreateGoalLoopFromPlanResponseSchema, {
        resource: { id: 49n },
        goalLoopId: 83n,
        workerSpecSnapshotId: 93n,
        resourceRevision: 5n,
      }),
    ));
    methods.applyPromptPlanConnect.mockResolvedValue(toBinary(
      ResourceSchema,
      create(ResourceSchema, { id: 45n }),
    ));
    methods.applyExpertPlanConnect.mockResolvedValue(toBinary(
      ApplyExpertPlanResponseSchema,
      create(ApplyExpertPlanResponseSchema, {
        resource: { id: 46n },
        expertId: 81n,
        workerSpecSnapshotId: 91n,
        resourceRevision: 3n,
      }),
    ));
    methods.applyWorkflowPlanConnect.mockResolvedValue(toBinary(
      ApplyWorkflowPlanResponseSchema,
      create(ApplyWorkflowPlanResponseSchema, {
        resource: { id: 47n },
        workflowId: 82n,
        workerSpecSnapshotId: 92n,
        resourceRevision: 4n,
      }),
    ));

    await expect(applyBindingResourcePlan("acme", "binding-plan"))
      .resolves.toMatchObject({ id: 43n });
    await expect(applyWorkerTemplatePlan("acme", "worker-plan"))
      .resolves.toMatchObject({
        resource: { id: 44n },
        workerSpecSnapshotId: 91n,
      });
    await expect(createWorkerFromPlan("acme", "worker-run-plan"))
      .resolves.toMatchObject({
        resource: { id: 48n },
        launchId: 71n,
        podId: 73n,
        podKey: "7-standalone-12345678",
        workerSpecSnapshotId: 91n,
        resourceRevision: 3n,
        runnerId: 11n,
      });
    await expect(createGoalLoopFromPlan("acme", "goal-loop-plan"))
      .resolves.toMatchObject({
        resource: { id: 49n },
        goalLoopId: 83n,
        workerSpecSnapshotId: 93n,
        resourceRevision: 5n,
      });
    await expect(applyPromptPlan("acme", "prompt-plan"))
      .resolves.toMatchObject({ id: 45n });
    await expect(applyExpertPlan("acme", "expert-plan"))
      .resolves.toMatchObject({
        resource: { id: 46n },
        expertId: 81n,
        workerSpecSnapshotId: 91n,
        resourceRevision: 3n,
      });
    await expect(applyWorkflowPlan("acme", "workflow-plan"))
      .resolves.toMatchObject({
        resource: { id: 47n },
        workflowId: 82n,
        workerSpecSnapshotId: 92n,
        resourceRevision: 4n,
      });
    expect(methods.applyBindingResourcePlanConnect).toHaveBeenCalledTimes(1);
    expect(methods.applyWorkerTemplatePlanConnect).toHaveBeenCalledTimes(1);
    expect(methods.createWorkerFromPlanConnect).toHaveBeenCalledTimes(1);
    expect(methods.createGoalLoopFromPlanConnect).toHaveBeenCalledTimes(1);
    expect(methods.applyPromptPlanConnect).toHaveBeenCalledTimes(1);
    expect(methods.applyExpertPlanConnect).toHaveBeenCalledTimes(1);
    expect(methods.applyWorkflowPlanConnect).toHaveBeenCalledTimes(1);

    const promptRequest = fromBinary(
      ApplyPromptPlanRequestSchema,
      methods.applyPromptPlanConnect.mock.calls[0][0],
    );
    expect(promptRequest).toMatchObject({
      orgSlug: "acme",
      planId: "prompt-plan",
    });
    const workerRequest = fromBinary(
      CreateWorkerFromPlanRequestSchema,
      methods.createWorkerFromPlanConnect.mock.calls[0][0],
    );
    expect(workerRequest).toMatchObject({
      orgSlug: "acme",
      planId: "worker-run-plan",
    });
    const goalLoopRequest = fromBinary(
      CreateGoalLoopFromPlanRequestSchema,
      methods.createGoalLoopFromPlanConnect.mock.calls[0][0],
    );
    expect(goalLoopRequest).toMatchObject({
      orgSlug: "acme",
      planId: "goal-loop-plan",
    });
    const expertRequest = fromBinary(
      ApplyExpertPlanRequestSchema,
      methods.applyExpertPlanConnect.mock.calls[0][0],
    );
    expect(expertRequest).toMatchObject({
      orgSlug: "acme",
      planId: "expert-plan",
    });
    const workflowRequest = fromBinary(
      ApplyWorkflowPlanRequestSchema,
      methods.applyWorkflowPlanConnect.mock.calls[0][0],
    );
    expect(workflowRequest).toMatchObject({
      orgSlug: "acme",
      planId: "workflow-plan",
    });
  });
});
