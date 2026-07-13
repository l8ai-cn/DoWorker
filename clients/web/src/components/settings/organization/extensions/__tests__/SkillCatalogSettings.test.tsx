import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
import type { CatalogSkill } from "@/lib/api";
import { SkillCatalogSettings } from "../SkillCatalogSettings";

const api = vi.hoisted(() => ({
  list: vi.fn(),
  update: vi.fn(),
  syncUpstream: vi.fn(),
  delete: vi.fn(),
}));

vi.mock("@/lib/api", () => ({ skillCatalogApi: api }));
vi.mock("@/stores/auth", () => ({ useCurrentOrg: () => ({ slug: "acme" }) }));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

const t = (key: string) => key;

function catalogSkill(id: number, slug: string, tags = ["video"]): CatalogSkill {
  return {
    id,
    organization_id: 1,
    slug,
    display_name: `Skill ${id}`,
    description: "",
    license: "",
    tags,
    is_active: true,
    git_repo_path: `am-skills/${slug}`,
    default_branch: "main",
    install_source: "gitops",
    content_sha: "",
    storage_key: "",
    package_size: 0,
    version: 1,
    created_at: "2026-07-14T00:00:00Z",
    updated_at: "2026-07-14T00:00:00Z",
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, reject, resolve };
}

describe("SkillCatalogSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.list.mockResolvedValue({ skills: [catalogSkill(1, "video-editing")], total: 1 });
  });

  it("loads every catalog page beyond the first 50 skills", async () => {
    const skills = Array.from({ length: 51 }, (_, index) => (
      catalogSkill(index + 1, `skill-${index + 1}`)
    ));
    api.list
      .mockResolvedValueOnce({ skills: skills.slice(0, 50), total: 51 })
      .mockResolvedValueOnce({ skills: skills.slice(50), total: 51 });

    render(<SkillCatalogSettings t={t} />);

    expect(await screen.findByText("Skill 51")).toBeVisible();
    expect(api.list).toHaveBeenNthCalledWith(1, 50, 0);
    expect(api.list).toHaveBeenNthCalledWith(2, 50, 50);
  });

  it("shows a load error and retries the full catalog load", async () => {
    api.list
      .mockRejectedValueOnce(new Error("catalog unavailable"))
      .mockResolvedValueOnce({ skills: [catalogSkill(2, "audio-mixing", ["audio"])], total: 1 });

    render(<SkillCatalogSettings t={t} />);

    expect(await screen.findByText("extensions.failedToLoadSkills")).toBeVisible();
    expect(screen.getByRole("button", {
      name: "extensions.skillCatalog.retry",
    })).toBeVisible();
    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.retry",
    }));

    expect(await screen.findByText("Skill 2")).toBeVisible();
    expect(screen.queryByText("extensions.failedToLoadSkills")).not.toBeInTheDocument();
    expect(api.list).toHaveBeenNthCalledWith(1, 50, 0);
    expect(api.list).toHaveBeenNthCalledWith(2, 50, 0);
    expect(toast.error).toHaveBeenCalledWith("catalog unavailable");
  });

  it("holds independent saves and blocks reentry for the same skill", async () => {
    const first = catalogSkill(1, "video-editing");
    const second = catalogSkill(2, "audio-mixing", ["audio"]);
    const firstUpdate = deferred<CatalogSkill>();
    const secondUpdate = deferred<CatalogSkill>();
    api.list.mockResolvedValue({ skills: [first, second], total: 2 });
    api.update.mockImplementation((slug: string) => (
      slug === first.slug ? firstUpdate.promise : secondUpdate.promise
    ));
    render(<SkillCatalogSettings t={t} />);
    await screen.findByText("Skill 1");

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: Skill 1",
    }));
    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.saveTags" }));
    expect(screen.getByRole("button", {
      name: "extensions.skillCatalog.savingTags",
    })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: Skill 2",
    }));
    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.saveTags" }));
    expect(screen.getAllByRole("button", {
      name: "extensions.skillCatalog.savingTags",
    })).toHaveLength(2);

    fireEvent.click(screen.getAllByRole("button", {
      name: "extensions.skillCatalog.savingTags",
    })[0]);
    expect(api.update).toHaveBeenCalledTimes(2);

    secondUpdate.resolve({ ...second, tags: ["audio", "curated"] });
    await waitFor(() => expect(screen.getAllByRole("button", {
      name: "extensions.skillCatalog.savingTags",
    })).toHaveLength(1));
    expect(screen.getByRole("button", {
      name: "extensions.skillCatalog.savingTags",
    })).toBeDisabled();

    firstUpdate.resolve({ ...first, tags: ["curated", "video"] });
    await waitFor(() => expect(screen.queryByRole("button", {
      name: "extensions.skillCatalog.savingTags",
    })).not.toBeInTheDocument());
  });

  it("updates only catalog tags and reports a failed save", async () => {
    api.update.mockRejectedValue(new Error("save failed"));
    render(<SkillCatalogSettings t={t} />);
    await screen.findByText("Skill 1");

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: Skill 1",
    }));
    const input = screen.getByLabelText("extensions.skillCatalog.tagInput");
    fireEvent.change(input, { target: { value: "curated" } });
    fireEvent.keyDown(input, { key: "Enter" });
    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.saveTags" }));

    await waitFor(() => expect(api.update).toHaveBeenCalledWith("video-editing", {
      tags: ["curated", "video"],
    }));
    expect(toast.error).toHaveBeenCalledWith("save failed");
    expect(screen.getByText("extensions.skillCatalog.failedToSaveTags")).toBeVisible();

    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.cancelTags" }));
    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: Skill 1",
    }));
    expect(screen.queryByText("extensions.skillCatalog.failedToSaveTags")).not.toBeInTheDocument();
  });
});
