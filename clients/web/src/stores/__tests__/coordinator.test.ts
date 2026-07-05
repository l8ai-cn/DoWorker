import { beforeEach, describe, expect, it, vi } from "vitest";
import { coordinatorApi, type CoordinatorRunResult } from "@/lib/api/coordinatorApi";
import { useCoordinatorStore } from "@/stores/coordinator";

vi.mock("@/lib/api/coordinatorApi", () => ({
  coordinatorApi: {
    listProjects: vi.fn(),
    createProject: vi.fn(),
    updateProject: vi.fn(),
    deleteProject: vi.fn(),
    listExecutions: vi.fn(),
    runNow: vi.fn(),
  },
}));

const api = vi.mocked(coordinatorApi);

describe("useCoordinatorStore", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    useCoordinatorStore.setState({
      projects: [],
      executions: {},
      runResults: {},
      loading: false,
      error: null,
    });
  });

  it("stores the latest run result when a project is triggered", async () => {
    const result: CoordinatorRunResult = {
      project_id: 7,
      scanned: 3,
      candidates: 2,
      claimed: 1,
      dispatched: 1,
      skipped: 1,
      errors: [],
    };
    api.runNow.mockResolvedValue(result);
    api.listExecutions.mockResolvedValue([]);

    await useCoordinatorStore.getState().runNow(7);

    expect(useCoordinatorStore.getState().runResults[7]).toEqual(result);
  });
});
