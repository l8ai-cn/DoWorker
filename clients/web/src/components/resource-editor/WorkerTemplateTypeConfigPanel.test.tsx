import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { WorkerTemplateTypeConfigPanel } from "./WorkerTemplateTypeConfigPanel";

describe("WorkerTemplateTypeConfigPanel", () => {
  it("edits only definition-declared credential bundle references", () => {
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.typeConfig.secretRefs = {
      CURSOR_API_KEY: {
        kind: "EnvironmentBundle",
        name: "production-secrets",
        revision: 3,
      },
    };

    render(
      <WorkerTemplateTypeConfigPanel
        draft={draft}
        credentialRequirements={[{
          id: "cursor",
          source_kind: "credential_bundle",
          source_ref: "cursor",
          target_kind: "env",
          target_name: "CURSOR_API_KEY",
        }]}
        requiredCredentialFields={new Set()}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {
            "EnvironmentBundle:credential:CURSOR_API_KEY": [{
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
    expect(screen.getByRole("combobox", {
      name: "CURSOR_API_KEY",
    })).toHaveTextContent("Production secrets · production-secrets");
    expect(screen.getByLabelText("CURSOR_API_KEY Revision")).toHaveValue(3);
    expect(screen.queryByLabelText(/^Value$/)).not.toBeInTheDocument();
    expect(screen.queryByDisplayValue(/secret-value/i)).not.toBeInTheDocument();
  });
});
