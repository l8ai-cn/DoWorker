import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { WorkerTemplateTypeConfigPanel } from "./WorkerTemplateTypeConfigPanel";

describe("WorkerTemplateTypeConfigPanel", () => {
  it("edits secrets only as EnvironmentBundle references", () => {
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.typeConfig.secretRefs = {
      "api-token": {
        kind: "EnvironmentBundle",
        name: "production-secrets",
        revision: 3,
      },
    };

    render(
      <WorkerTemplateTypeConfigPanel
        draft={draft}
        catalog={{
          loading: false,
          error: null,
          byKind: {
            EnvironmentBundle: [{
              name: "production-secrets",
              displayName: "Production secrets",
              revision: 3,
            }],
          },
        }}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByText(
      "Only references to EnvironmentBundle resources are stored. " +
      "Secret values are never entered or displayed here.",
    )).toBeInTheDocument();
    expect(screen.getByRole("textbox", {
      name: /Configuration key/,
    })).toHaveValue("api-token");
    expect(screen.getByRole("combobox", {
      name: /Resource reference/,
    })).toHaveValue(
      "production-secrets",
    );
    expect(screen.getByLabelText("Resource reference Revision")).toHaveValue(3);
    expect(screen.queryByLabelText(/^Value$/)).not.toBeInTheDocument();
    expect(screen.queryByDisplayValue(/secret-value/i)).not.toBeInTheDocument();
  });
});
