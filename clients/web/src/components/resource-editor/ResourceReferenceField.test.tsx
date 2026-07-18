import { describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";
import { render, screen } from "@/test/test-utils";
import { ResourceReferenceField } from "./ResourceReferenceField";

describe("ResourceReferenceField", () => {
  it("gives the revision input a stable field identity", () => {
    render(
      <ResourceReferenceField
        id="model-reference"
        label="Model binding"
        kind="Model"
        value={{ kind: "Model", name: "gpt-5", revision: 2 }}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByLabelText("Model binding Revision")).toHaveAttribute(
      "id",
      "model-reference-revision",
    );
  });

  it("exposes required references to the browser and assistive technology", () => {
    render(
      <ResourceReferenceField
        id="credential-reference"
        label="Credential"
        kind="EnvironmentBundle"
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        required
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByRole("combobox", { name: /Credential/ }))
      .toBeRequired();
    expect(screen.getByRole("combobox", { name: /Credential/ }))
      .toHaveAttribute("aria-required", "true");
  });

  it("uses a purpose-specific catalog without changing the stored resource kind", async () => {
    const user = userEvent.setup();
    render(
      <ResourceReferenceField
        id="config-reference"
        label="Config"
        kind="EnvironmentBundle"
        catalogKey="EnvironmentBundle:config"
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {
            "EnvironmentBundle:config": [{
              name: "do-agent-settings",
              displayName: "Do Agent settings",
              revision: 4,
            }],
            "EnvironmentBundle:credential:CURSOR_API_KEY": [{
              name: "cursor-secrets",
              displayName: "Cursor secrets",
              revision: 2,
            }],
          },
        }}
        onChange={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "Config" }));
    expect(screen.getByRole("option", {
      name: /Do Agent settings do-agent-settings/,
    })).toBeInTheDocument();
    expect(screen.queryByRole("option", {
      name: /Cursor secrets cursor-secrets/,
    })).not.toBeInTheDocument();
  });

  it("clears a pinned revision when the referenced identity changes", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <ResourceReferenceField
        id="model-reference"
        label="Model binding"
        kind="Model"
        value={{ kind: "Model", name: "gpt-5", revision: 4 }}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {
            Model: [
              { name: "gpt-5", displayName: "GPT-5", revision: 4 },
              {
                name: "claude-sonnet",
                displayName: "Claude Sonnet",
                revision: 2,
              },
            ],
          },
        }}
        onChange={onChange}
      />,
    );

    await user.click(screen.getByRole("combobox", {
      name: "Model binding",
    }));
    await user.click(screen.getByRole("option", {
      name: /Claude Sonnet claude-sonnet/,
    }));

    expect(onChange).toHaveBeenCalledWith({
      kind: "Model",
      name: "claude-sonnet",
      revision: undefined,
    });
  });

  it.each([
    {
      name: "loading",
      catalog: {
        loading: true,
        error: null,
        errorsByKind: {},
        byKind: {},
      },
    },
    {
      name: "permission error",
      catalog: {
        loading: false,
        error: "Permission denied",
        errorsByKind: {},
        byKind: {},
      },
    },
    {
      name: "unresolved value",
      catalog: {
        loading: false,
        error: null,
        errorsByKind: {},
        byKind: {
          Model: [{ name: "other-model", displayName: "", revision: 1 }],
        },
      },
    },
  ])("keeps the reference read-only while the catalog is $name", ({
    catalog,
  }) => {
    render(
      <ResourceReferenceField
        id="model-reference"
        label="Model binding"
        kind="Model"
        value={{ kind: "Model", name: "gpt-5", revision: 4 }}
        catalog={catalog}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByRole("combobox", { name: "Model binding" }))
      .toBeDisabled();
    expect(screen.getByLabelText("Model binding Revision"))
      .toHaveAttribute("readonly");
  });
});
