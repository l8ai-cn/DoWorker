import { useState } from "react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import type { WorkerTemplateConfigDocumentBinding } from "./resource-editor-types";
import { WorkerTemplateConfigDocumentBindingsField } from "./WorkerTemplateConfigDocumentBindingsField";

describe("WorkerTemplateConfigDocumentBindingsField", () => {
  it("keeps unresolved config references visible and read-only", () => {
    render(
      <WorkerTemplateConfigDocumentBindingsField
        requirements={[]}
        value={[
          {
            documentId: "settings",
            configBundleRef: {
              kind: "EnvironmentBundle",
              name: "existing-settings",
              revision: 3,
            },
          },
        ]}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByRole("textbox", { name: "settings" }))
      .toHaveValue("existing-settings");
    expect(screen.getByRole("textbox", { name: "settings" }))
      .toHaveAttribute("readonly");
    expect(screen.getByRole("spinbutton", { name: /settings/ }))
      .toHaveValue(3);
    expect(screen.getByRole("spinbutton", { name: /settings/ }))
      .toHaveAttribute("readonly");
  });

  it("binds each declared document and removes bindings outside the catalog", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<BindingsHarness onChange={onChange} />);

    expect(screen.getByText("json · /workspace/settings.json"))
      .toBeInTheDocument();
    expect(screen.getByText("toml · /workspace/auth.toml"))
      .toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: /settings/ }))
      .toHaveAttribute("aria-required", "true");
    expect(screen.getByRole("combobox", { name: /auth/ }))
      .toHaveAttribute("aria-required", "false");

    await user.click(screen.getByRole("combobox", { name: /settings/ }));
    await user.click(screen.getByRole("option", {
      name: /Settings bundle settings-bundle/,
    }));

    expect(onChange).toHaveBeenLastCalledWith([
      {
        documentId: "settings",
        configBundleRef: {
          kind: "EnvironmentBundle",
          name: "settings-bundle",
        },
      },
      {
        documentId: "auth",
        configBundleRef: {
          kind: "EnvironmentBundle",
          name: "auth-bundle",
          revision: 2,
        },
      },
    ]);

    await user.click(screen.getByRole("combobox", { name: /auth/ }));
    await user.click(screen.getByRole("option", { name: "None configured." }));

    expect(onChange).toHaveBeenLastCalledWith([
      {
        documentId: "settings",
        configBundleRef: {
          kind: "EnvironmentBundle",
          name: "settings-bundle",
        },
      },
    ]);
  });
});

function BindingsHarness({
  onChange,
}: {
  onChange: (value: WorkerTemplateConfigDocumentBinding[]) => void;
}) {
  const [value, setValue] = useState<WorkerTemplateConfigDocumentBinding[]>([
    {
      documentId: "auth",
      configBundleRef: {
        kind: "EnvironmentBundle",
        name: "auth-bundle",
        revision: 2,
      },
    },
    {
      documentId: "obsolete",
      configBundleRef: {
        kind: "EnvironmentBundle",
        name: "obsolete-bundle",
      },
    },
  ]);
  return (
    <WorkerTemplateConfigDocumentBindingsField
      requirements={[
        {
          document_id: "settings",
          format: "json",
          target_path: "/workspace/settings.json",
          required: true,
        },
        {
          document_id: "auth",
          format: "toml",
          target_path: "/workspace/auth.toml",
          required: false,
        },
      ]}
      value={value}
      catalog={{
        loading: false,
        error: null,
        errorsByKind: {},
        byKind: {
          "EnvironmentBundle:config": [
            {
              name: "settings-bundle",
              displayName: "Settings bundle",
              revision: 1,
            },
            {
              name: "auth-bundle",
              displayName: "Auth bundle",
              revision: 2,
            },
          ],
        },
      }}
      onChange={(next) => {
        setValue(next);
        onChange(next);
      }}
    />
  );
}
