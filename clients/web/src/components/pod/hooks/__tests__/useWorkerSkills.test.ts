import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockListMarketSkills = vi.fn();
const mockListRepoSkills = vi.fn();

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "test-org" }),
}));

vi.mock("@/lib/api/facade/marketExtension", () => ({
  listMarketSkills: (...args: unknown[]) => mockListMarketSkills(...args),
}));

vi.mock("@/lib/api/facade/repoSkillExtension", () => ({
  listRepoSkills: (...args: unknown[]) => mockListRepoSkills(...args),
}));

vi.mock("@/stores/podCreation", () => ({
  usePodCreationStore: () => ({ lastSkillSlugs: [] }),
}));

import { useWorkerSkills } from "../useWorkerSkills";

describe("useWorkerSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("loads catalog skills when no repository is selected", async () => {
    mockListMarketSkills.mockResolvedValue({
      items: [
        { id: 11, slug: "pdf-tool", is_active: true },
        { id: 12, slug: "disabled", is_active: false },
      ],
    });

    const { result } = renderHook(() => useWorkerSkills(null));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockListMarketSkills).toHaveBeenCalledWith("test-org");
    expect(mockListRepoSkills).not.toHaveBeenCalled();
    expect(result.current.skills).toEqual([
      { id: 11, slug: "pdf-tool", scope: "org" },
    ]);
  });

  it("uses catalog IDs for repository-installed skills", async () => {
    mockListRepoSkills.mockImplementation(
      (_org: string, _repo: number, opts: { scope: string }) =>
        Promise.resolve({
          items: opts.scope === "org"
            ? [{ id: 91, market_item_id: 21, slug: "pdf-tool", scope: "org", is_enabled: true }]
            : [],
        }),
    );

    const { result } = renderHook(() => useWorkerSkills(7));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.skills).toEqual([
      { id: 21, slug: "pdf-tool", scope: "org" },
    ]);
  });
});
