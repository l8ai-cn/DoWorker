import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CatalogSkill } from "@/lib/api";
import { CatalogSkillList } from "../CatalogSkillList";

const t = (key: string) => key;

function skill(id: number, slug: string, tags: string[]): CatalogSkill {
  return {
    id,
    organization_id: 1,
    slug,
    display_name: slug,
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

const skills = [
  skill(1, "video-editing", ["editing", "video"]),
  skill(2, "video-captioning", ["video"]),
  skill(3, "audio-mixing", ["audio"]),
  skill(4, "general-review", []),
];

function renderList(overrides: Partial<React.ComponentProps<typeof CatalogSkillList>> = {}) {
  const props: React.ComponentProps<typeof CatalogSkillList> = {
    t,
    loading: false,
    skills,
    syncingSlug: null,
    savingSlugs: new Set(),
    saveErrorSlugs: new Set(),
    onSync: vi.fn(),
    onDelete: vi.fn(),
    onImport: vi.fn(),
    onEditTags: vi.fn(),
    onUpdateTags: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
  return { ...render(<CatalogSkillList {...props} />), props };
}

describe("CatalogSkillList", () => {
  it("shows loading, empty, and no-filter-results states", () => {
    const { rerender, props } = renderList({ loading: true, skills: [] });
    expect(screen.getByText("extensions.loading")).toBeVisible();

    rerender(<CatalogSkillList {...props} loading={false} skills={[]} />);
    expect(screen.getByText("extensions.skillCatalog.noSkills")).toBeVisible();

    rerender(<CatalogSkillList {...props} loading={false} skills={skills} />);
    fireEvent.click(screen.getByRole("button", { name: "audio" }));
    rerender(<CatalogSkillList {...props} loading={false} skills={skills.slice(0, 2)} />);
    expect(screen.getByText("extensions.skillCatalog.noFilterResults")).toBeVisible();
  });

  it("filters by any selected tag", () => {
    renderList();

    fireEvent.click(screen.getByRole("button", { name: "video" }));
    fireEvent.click(screen.getByRole("button", { name: "audio" }));

    expect(screen.getAllByText("video-editing")).toHaveLength(2);
    expect(screen.getAllByText("video-captioning")).toHaveLength(2);
    expect(screen.getAllByText("audio-mixing")).toHaveLength(2);
    expect(screen.queryByText("general-review")).not.toBeInTheDocument();
  });

  it("groups multi-tag skills in every tag group and includes untagged", () => {
    renderList();

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.groupByTag",
    }));

    const editing = screen.getByRole("region", {
      name: "extensions.skillCatalog.tagGroup: editing",
    });
    const video = screen.getByRole("region", {
      name: "extensions.skillCatalog.tagGroup: video",
    });
    const untagged = screen.getByRole("region", { name: "extensions.skillCatalog.untagged" });
    expect(within(editing).getAllByText("video-editing")).toHaveLength(2);
    expect(within(video).getAllByText("video-editing")).toHaveLength(2);
    expect(within(untagged).getAllByText("general-review")).toHaveLength(2);
  });

  it("keeps a real 未标签 tag separate from the untagged group", () => {
    const collisionT = (key: string) => {
      if (key === "extensions.skillCatalog.untagged") return "未标签";
      if (key === "extensions.skillCatalog.tagGroup") return "标签";
      return key;
    };
    renderList({
      t: collisionT,
      skills: [
        skill(5, "named-untagged", ["未标签"]),
        skill(6, "actually-untagged", []),
      ],
    });

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.groupByTag",
    }));

    const tagged = screen.getByRole("region", { name: "标签: 未标签" });
    const untagged = screen.getByRole("region", { name: "未标签" });
    expect(within(tagged).getAllByText("named-untagged")).toHaveLength(2);
    expect(within(tagged).queryByText("actually-untagged")).not.toBeInTheDocument();
    expect(within(untagged).getAllByText("actually-untagged")).toHaveLength(2);
    expect(within(untagged).queryByText("named-untagged")).not.toBeInTheDocument();
  });

  it("edits tags with Enter, delete, save, and disabled saving state", async () => {
    const onUpdateTags = vi.fn().mockReturnValue(new Promise<void>(() => undefined));
    const { rerender, props } = renderList({ onUpdateTags });

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: video-editing",
    }));
    const input = screen.getByLabelText("extensions.skillCatalog.tagInput");
    fireEvent.change(input, { target: { value: "Curated" } });
    fireEvent.keyDown(input, { key: "Enter" });
    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.removeTag: editing",
    }));
    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.saveTags" }));

    await waitFor(() => expect(onUpdateTags).toHaveBeenCalledWith(
      "video-editing",
      ["curated", "video"],
    ));

    rerender(<CatalogSkillList {...props} savingSlugs={new Set(["video-editing"])} />);
    expect(screen.getByRole("button", { name: "extensions.skillCatalog.savingTags" })).toBeDisabled();
    expect(screen.getByLabelText("extensions.skillCatalog.tagInput")).toBeDisabled();
  });

  it("includes pending input when save is clicked directly", async () => {
    const onUpdateTags = vi.fn().mockResolvedValue(undefined);
    renderList({ onUpdateTags });
    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: video-editing",
    }));
    const input = screen.getByLabelText("extensions.skillCatalog.tagInput");
    fireEvent.change(input, { target: { value: " Curated " } });

    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.saveTags",
    }));

    await waitFor(() => expect(onUpdateTags).toHaveBeenCalledWith(
      "video-editing",
      ["curated", "editing", "video"],
    ));
  });

  it("limits tag input and disables adding after twenty tags", () => {
    renderList({
      skills: [skill(7, "tag-limit", Array.from({ length: 20 }, (_, index) => `tag-${index}`))],
    });
    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: tag-limit",
    }));

    const input = screen.getByLabelText("extensions.skillCatalog.tagInput");
    expect(input).toHaveAttribute("maxlength", "40");
    expect(input).toBeDisabled();
  });

  it("keeps the editor open with a save failure state", () => {
    renderList({ saveErrorSlugs: new Set(["video-editing"]) });
    fireEvent.click(screen.getByRole("button", {
      name: "extensions.skillCatalog.editTags: video-editing",
    }));
    expect(screen.getByText("extensions.skillCatalog.failedToSaveTags")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.cancelTags" }));
    expect(screen.queryByLabelText("extensions.skillCatalog.tagInput")).not.toBeInTheDocument();
  });
});
