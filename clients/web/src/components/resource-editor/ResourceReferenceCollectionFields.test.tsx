import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ResourceReferenceListField } from "./ResourceReferenceListField";
import { ResourceReferenceMapField } from "./ResourceReferenceMapField";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

const blockedCatalog = {
  loading: true,
  error: null,
  errorsByKind: {},
  byKind: {
    Prompt: [{ name: "release-review", displayName: "Release", revision: 1 }],
  },
};

describe("resource reference collection fields", () => {
  it("keeps list structure unchanged while its catalog is blocked", () => {
    const onChange = vi.fn();
    render(
      <ResourceReferenceListField
        id="prompts"
        label="Prompts"
        kind="Prompt"
        value={[{ kind: "Prompt", name: "release-review" }]}
        catalog={blockedCatalog}
        onChange={onChange}
      />,
    );

    const add = screen.getByRole("button", { name: "collections.add Prompts" });
    const remove = screen.getByRole("button", {
      name: "collections.remove Prompts 1",
    });
    expect(add).toBeDisabled();
    expect(remove).toBeDisabled();
    fireEvent.click(add);
    fireEvent.click(remove);
    expect(onChange).not.toHaveBeenCalled();
  });

  it("keeps map keys and structure unchanged while its catalog is blocked", () => {
    const onChange = vi.fn();
    render(
      <ResourceReferenceMapField
        id="prompt-bindings"
        label="Prompt bindings"
        keyLabel="Binding"
        kind="Prompt"
        value={{ reviewer: { kind: "Prompt", name: "release-review" } }}
        catalog={blockedCatalog}
        onChange={onChange}
      />,
    );

    const key = screen.getByRole("textbox", { name: /Binding/ });
    expect(key).toBeDisabled();
    expect(screen.getByRole("button", {
      name: "collections.add Prompt bindings",
    })).toBeDisabled();
    expect(screen.getByRole("button", {
      name: "collections.remove Prompt bindings 1",
    })).toBeDisabled();
    fireEvent.change(key, { target: { value: "changed" } });
    expect(onChange).not.toHaveBeenCalled();
  });
});
