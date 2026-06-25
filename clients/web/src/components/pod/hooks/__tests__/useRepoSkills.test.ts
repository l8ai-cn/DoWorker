import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { InstalledSkill } from "@/lib/viewModels/extension";

const storeState = {
  lastSkillSlugs: [] as string[],
};

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "test-org" }),
}));

vi.mock("@/stores/podCreation", () => ({
  usePodCreationStore: () => storeState,
}));

const mockListRepoSkills = vi.fn();
vi.mock("@/lib/api/facade/repoSkillExtension", () => ({
  listRepoSkills: (...args: unknown[]) => mockListRepoSkills(...args),
}));

import { useRepoSkills } from "../useCreatePodFormEffects";

function skill(slug: string, enabled = true): InstalledSkill {
  return { slug, is_enabled: enabled } as InstalledSkill;
}

describe("useRepoSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    storeState.lastSkillSlugs = [];
  });

  it("loads enabled org + user skills and drops disabled ones", async () => {
    mockListRepoSkills.mockImplementation((_o, _r, opts: { scope?: string }) =>
      Promise.resolve({
        items: opts.scope === "org" ? [skill("pdf-tool"), skill("off", false)] : [skill("commit-helper")],
      }),
    );

    const { result } = renderHook(() => useRepoSkills(7));

    await waitFor(() => expect(result.current.loadingSkills).toBe(false));
    expect(result.current.repoSkills.map((s) => s.slug)).toEqual(["pdf-tool", "commit-helper"]);
    expect(result.current.selectedSkillSlugs).toEqual([]);
  });

  it("restores persisted selection filtered to still-installed skills", async () => {
    storeState.lastSkillSlugs = ["pdf-tool", "stale-removed"];
    mockListRepoSkills.mockImplementation((_o, _r, opts: { scope?: string }) =>
      Promise.resolve({ items: opts.scope === "org" ? [skill("pdf-tool")] : [] }),
    );

    const { result } = renderHook(() => useRepoSkills(7));

    await waitFor(() => expect(result.current.loadingSkills).toBe(false));
    expect(result.current.selectedSkillSlugs).toEqual(["pdf-tool"]);
  });

  it("clears skills and selection when no repository is selected", async () => {
    storeState.lastSkillSlugs = ["pdf-tool"];
    const { result } = renderHook(() => useRepoSkills(null));

    await waitFor(() => expect(result.current.repoSkills).toEqual([]));
    expect(result.current.selectedSkillSlugs).toEqual([]);
    expect(mockListRepoSkills).not.toHaveBeenCalled();
  });
});
