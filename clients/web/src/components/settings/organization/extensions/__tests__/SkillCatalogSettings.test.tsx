import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
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
const catalogSkill = {
  id: 1,
  organization_id: 1,
  slug: "video-editing",
  display_name: "Video Editing",
  description: "",
  license: "",
  tags: ["video"],
  is_active: true,
  git_repo_path: "am-skills/video-editing",
  default_branch: "main",
  install_source: "gitops",
  content_sha: "",
  storage_key: "",
  package_size: 0,
  version: 1,
  created_at: "2026-07-14T00:00:00Z",
  updated_at: "2026-07-14T00:00:00Z",
};

describe("SkillCatalogSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.list.mockResolvedValue({ skills: [catalogSkill], total: 1 });
  });

  it("updates only catalog tags and reports a failed save", async () => {
    api.update.mockRejectedValue(new Error("save failed"));
    render(<SkillCatalogSettings t={t} />);
    await screen.findByText("Video Editing");

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: Video Editing",
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
      name: "extensions.skillCatalog.editTags: Video Editing",
    }));
    expect(screen.queryByText("extensions.skillCatalog.failedToSaveTags")).not.toBeInTheDocument();
  });
});
