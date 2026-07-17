import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import { ResourceReferenceField } from "./ResourceReferenceField";

const catalog = {
  loading: false,
  error: null,
  byKind: {
    WorkerTemplate: [
      {
        name: "codex-workbench",
        displayName: "Codex Workbench",
        revision: 3,
      },
    ],
  },
};

describe("ResourceReferenceField", () => {
  it("selects a known resource through the platform select control", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <ResourceReferenceField
        id="worker-template"
        label="Worker template"
        kind="WorkerTemplate"
        catalog={catalog}
        required
        onChange={onChange}
      />,
    );

    await user.click(screen.getByRole("combobox", {
      name: "Worker template",
    }));
    await user.click(screen.getByRole("option", {
      name: "Codex Workbench codex-workbench",
    }));

    expect(onChange).toHaveBeenCalledWith({
      kind: "WorkerTemplate",
      name: "codex-workbench",
      revision: undefined,
    });
  });

  it("allows an optional resource reference to be cleared", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <ResourceReferenceField
        id="prompt"
        label="Prompt"
        kind="Prompt"
        value={{ kind: "Prompt", name: "review-prompt" }}
        catalog={{
          loading: false,
          error: null,
          byKind: {
            Prompt: [
              {
                name: "review-prompt",
                displayName: "Review prompt",
                revision: 1,
              },
            ],
          },
        }}
        onChange={onChange}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "Prompt" }));
    await user.click(screen.getByRole("option", { name: "None configured." }));

    expect(onChange).toHaveBeenCalledWith(undefined);
  });
});
