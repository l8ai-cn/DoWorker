import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  KnowledgeBase,
  KnowledgeMountSelection,
} from "@/lib/api/facade/knowledgeBaseApi";
import { KnowledgeBaseMountSelect } from "../KnowledgeBaseMountSelect";

const mockListKnowledgeBases = vi.fn<() => Promise<KnowledgeBase[]>>();

vi.mock("@/lib/api/facade/knowledgeBaseApi", () => ({
  listKnowledgeBases: () => mockListKnowledgeBases(),
}));
vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme" }),
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("KnowledgeBaseMountSelect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListKnowledgeBases.mockResolvedValue([
      knowledgeBase(1, "docs-a"),
      knowledgeBase(2, "docs-b"),
    ]);
  });

  it("changes only the selected ID-backed mount", async () => {
    const onChange = vi.fn();
    const selected: KnowledgeMountSelection[] = [
      { id: 1, slug: "", mode: "ro" },
      { id: 2, slug: "", mode: "ro" },
    ];

    render(
      <KnowledgeBaseMountSelect
        selectedMounts={selected}
        onChange={onChange}
      />,
    );

    await waitFor(() => expect(mockListKnowledgeBases).toHaveBeenCalled());
    fireEvent.click(
      screen.getAllByTitle("ide.createPod.knowledgeModeToggle")[0],
    );

    expect(onChange).toHaveBeenCalledWith([
      { id: 1, slug: "", mode: "rw" },
      { id: 2, slug: "", mode: "ro" },
    ]);
  });

  it("removes a stale ID even when it is absent from the loaded catalog", async () => {
    const onChange = vi.fn();
    render(
      <KnowledgeBaseMountSelect
        selectedMounts={[{ id: 99, slug: "", mode: "ro" }]}
        onChange={onChange}
      />,
    );

    await waitFor(() => expect(mockListKnowledgeBases).toHaveBeenCalled());
    fireEvent.click(
      screen.getByRole("button", {
        name: "ide.createPod.removeKnowledgeBase",
      }),
    );

    expect(onChange).toHaveBeenCalledWith([]);
  });

  it("uses selection language and a localized loading error", async () => {
    mockListKnowledgeBases.mockRejectedValueOnce(new Error("backend detail"));
    render(
      <KnowledgeBaseMountSelect
        selectedMounts={[{ id: 1, slug: "", mode: "ro" }]}
        onChange={vi.fn()}
      />,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ide.createPod.knowledgeBasesLoadFailed",
    );
    expect(screen.getByText("ide.createPod.manageKnowledgeBases")).toBeInTheDocument();
    expect(screen.getByText("ide.createPod.knowledgeModeReadOnly")).toBeInTheDocument();
  });
});

function knowledgeBase(id: number, slug: string): KnowledgeBase {
  return {
    id,
    slug,
    name: slug,
    description: "",
    http_clone_url: "",
    default_branch: "main",
    source_type: "git",
    sync_status: "ready",
    created_at: "",
    updated_at: "",
  };
}
