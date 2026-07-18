import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import { WorkerTemplateCredentialBundleBindingsField } from "./WorkerTemplateCredentialBundleBindingsField";

describe("WorkerTemplateCredentialBundleBindingsField", () => {
  it("keeps unresolved credential references visible and read-only", () => {
    render(
      <WorkerTemplateCredentialBundleBindingsField
        requirements={[]}
        requiredFields={new Set()}
        value={{
          apiToken: {
            kind: "EnvironmentBundle",
            name: "existing-credentials",
            revision: 4,
          },
        }}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByRole("textbox", { name: "apiToken" }))
      .toHaveValue("existing-credentials");
    expect(screen.getByRole("textbox", { name: "apiToken" }))
      .toHaveAttribute("readonly");
    expect(screen.getByRole("spinbutton", { name: /apiToken/ }))
      .toHaveValue(4);
    expect(screen.getByRole("spinbutton", { name: /apiToken/ }))
      .toHaveAttribute("readonly");
  });
});
