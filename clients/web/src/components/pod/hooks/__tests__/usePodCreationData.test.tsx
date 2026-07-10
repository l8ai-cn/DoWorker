import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  listRunners: vi.fn(),
  listAgents: vi.fn(),
  fetchRepositories: vi.fn(),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "test-org" }),
}));

vi.mock("@/stores/repository", () => ({
  useRepositories: () => [],
  useRepositoryStore: (selector: (state: { fetchRepositories: () => void }) => unknown) =>
    selector({ fetchRepositories: mocks.fetchRepositories }),
}));

vi.mock("@/lib/api/facade/runnerConnect", () => ({
  listRunners: mocks.listRunners,
}));

vi.mock("@/lib/api/facade/agentConnect", () => ({
  listAgents: mocks.listAgents,
}));

import { usePodCreationData } from "../usePodCreationData";

describe("usePodCreationData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.listAgents.mockResolvedValue({ builtin_agents: [], custom_agents: [], agents: [] });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("surfaces runner loading failures instead of rendering an empty image list", async () => {
    mocks.listRunners.mockRejectedValue(new Error("runner service unavailable"));

    const { result } = renderHook(() => usePodCreationData(true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe("runner service unavailable");
    expect(result.current.availableAgents).toEqual([]);
    expect(console.error).toHaveBeenCalledWith(
      "Failed to load pod creation data:",
      expect.any(Error),
    );
  });
});
