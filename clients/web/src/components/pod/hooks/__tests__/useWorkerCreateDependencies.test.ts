import { renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { useWorkerCreateDependencies } from "../useWorkerCreateDependencies";

type WorkerTypeOption = WorkerCreateOptions["worker_types"][number];

const mockUseWorkerModelResources = vi.fn();

vi.mock("../useWorkerModelResources", () => ({
  useWorkerModelResources: (...args: unknown[]) => mockUseWorkerModelResources(...args),
}));
vi.mock("../useWorkerCreateEnvBundles", () => ({
  useWorkerCreateEnvBundles: () => ({
    runtime: { status: "ready", data: [] },
    credential: { status: "ready", data: [] },
    config: { status: "ready", data: [] },
  }),
}));
vi.mock("../useWorkerSkills", () => ({
  useWorkerSkills: () => ({
    loading: false,
    error: null,
    skills: [],
  }),
}));

describe("useWorkerCreateDependencies", () => {
  beforeEach(() => {
    mockUseWorkerModelResources.mockReset();
    mockUseWorkerModelResources.mockReturnValue({
      modelResources: [],
      toolModelResources: [],
      loadingModelResources: false,
      modelResourceError: null,
    });
  });

  it("passes the selected Worker Definition model requirement to the resource hook", () => {
    const selectedWorkerType: WorkerTypeOption = {
      slug: "definition-worker",
      name: "Definition Worker",
      description: "",
      schema_version: 1,
      config_schema: {},
      supported_interaction_modes: [],
      requires_model_resource: true,
      model_protocol_adapters: ["anthropic"],
      tool_model_requirements: [],
      credential_requirements: [],
      config_document_requirements: [],
      selectable: true,
      blocking_reason: "",
    };
    const useDependencies = useWorkerCreateDependencies as unknown as (
      workerType: WorkerTypeOption,
      repositoryId?: number,
    ) => unknown;

    renderHook(() => useDependencies(selectedWorkerType, 7));

    expect(mockUseWorkerModelResources).toHaveBeenCalledWith(
      "definition-worker",
      null,
      true,
      { required: true, protocolAdapters: ["anthropic"] },
    );
  });
});
