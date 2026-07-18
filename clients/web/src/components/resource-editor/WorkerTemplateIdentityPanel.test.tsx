import { describe, expect, it, vi } from "vitest";
import { render } from "@/test/test-utils";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { WorkerTemplateIdentityPanel } from "./WorkerTemplateIdentityPanel";

describe("WorkerTemplateIdentityPanel", () => {
  it("marks model binding as required for Worker types that need a model", () => {
    const { container } = render(
      <WorkerTemplateIdentityPanel
        draft={createWorkerTemplateDraft("acme")}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        modelRequired
        onChange={vi.fn()}
      />,
    );

    expect(
      container.querySelector('label[for="model-reference"]')?.textContent,
    ).toContain("*");
  });
});
