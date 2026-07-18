import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@/test/test-utils";
import { createResourceDraft } from "./resource-draft-factory";
import { ResourceEditorShell } from "./ResourceEditorShell";

const api = vi.hoisted(() => ({
  listResources: vi.fn(),
  planResource: vi.fn(),
  validateResource: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => ({
  ...api,
}));

vi.mock("./use-resource-reference-options", () => ({
  useResourceReferenceOptions: () => ({
    loading: false,
    error: null,
    errorsByKind: {},
    byKind: {},
  }),
}));

describe("ResourceEditorShell revision identity", () => {
  beforeEach(() => {
    Object.values(api).forEach((method) => method.mockReset());
    api.listResources.mockResolvedValue({ items: [] });
  });

  it("disables the resource name for a locked revision", () => {
    const draft = createResourceDraft("Workflow", "acme");
    draft.metadata.name = "nightly-review";

    render(
      <ResourceEditorShell
        orgSlug="acme"
        kind="Workflow"
        initialDraft={draft}
        lockedIdentity={{
          apiVersion: draft.apiVersion,
          kind: draft.kind,
          namespace: draft.metadata.namespace,
          name: draft.metadata.name,
        }}
      />,
    );

    expect(screen.getByLabelText(/Resource name/)).toBeDisabled();
  });

  it("blocks planning YAML that changes the locked resource name", async () => {
    const user = userEvent.setup();
    const draft = createResourceDraft("Workflow", "acme");
    draft.metadata.name = "nightly-review";

    render(
      <ResourceEditorShell
        orgSlug="acme"
        kind="Workflow"
        initialDraft={draft}
        lockedIdentity={{
          apiVersion: draft.apiVersion,
          kind: draft.kind,
          namespace: draft.metadata.namespace,
          name: draft.metadata.name,
        }}
      />,
    );

    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, {
      target: {
        value: (editor as HTMLTextAreaElement).value.replace(
          "name: nightly-review",
          "name: another-workflow",
        ),
      },
    });
    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findAllByText(
      "Resource identity cannot change when creating a revision.",
    )).not.toHaveLength(0);
    expect(api.planResource).not.toHaveBeenCalled();
  });
});
