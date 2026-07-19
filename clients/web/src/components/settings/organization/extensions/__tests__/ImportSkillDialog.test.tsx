import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ImportSkillDialog } from "../ImportSkillDialog";

const api = vi.hoisted(() => ({ import: vi.fn() }));

vi.mock("@/lib/api", () => ({ skillCatalogApi: api }));

const t = (key: string, params?: Record<string, string | number>) => (
  params ? `${key}:${JSON.stringify(params)}` : key
);

describe("ImportSkillDialog", () => {
  it("submits Pattern Designer as a compatible agent filter", async () => {
    api.import.mockResolvedValue({ skills: [], imported: 1 });
    render(
      <ImportSkillDialog t={t} open onOpenChange={vi.fn()} onImported={vi.fn()} />,
    );

    fireEvent.change(screen.getByPlaceholderText("https://github.com/owner/skills-repo"), {
      target: { value: "https://example.test/skills.git" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Pattern Designer" }));
    fireEvent.click(screen.getByRole("button", { name: "extensions.skillCatalog.import" }));

    await waitFor(() => expect(api.import).toHaveBeenCalledWith(expect.objectContaining({
      url: "https://example.test/skills.git",
      agent_filter: ["pattern-designer"],
    })));
  });
});
