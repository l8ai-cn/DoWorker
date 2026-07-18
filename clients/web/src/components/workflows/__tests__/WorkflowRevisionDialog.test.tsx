import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { WorkflowRevisionDialog } from "../WorkflowRevisionDialog";

const exportResource = vi.fn();

vi.mock("@/lib/api/facade/orchestrationResource", () => ({
  exportResource: (...args: unknown[]) => exportResource(...args),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/components/resource-editor/ResourceEditorShell", () => ({
  ResourceEditorShell: ({
    initialDraft,
    lockedIdentity,
    onApplied,
  }: {
    initialDraft: {
      metadata: { name: string };
      spec: { promptRef: { name: string } };
    };
    lockedIdentity: {
      kind: string;
      namespace: string;
      name: string;
    };
    onApplied: () => void;
  }) => (
    <button
      type="button"
      data-testid="revision-editor"
      data-name={initialDraft.metadata.name}
      data-prompt={initialDraft.spec.promptRef.name}
      data-locked-kind={lockedIdentity.kind}
      data-locked-namespace={lockedIdentity.namespace}
      data-locked-name={lockedIdentity.name}
      onClick={onApplied}
    >
      apply revision
    </button>
  ),
}));

vi.mock("@/components/ui/responsive-dialog", () => ({
  ResponsiveDialog: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
  ResponsiveDialogContent: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
  ResponsiveDialogHeader: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
  ResponsiveDialogTitle: ({ children }: { children: React.ReactNode }) => (
    <h2>{children}</h2>
  ),
}));

const exportedWorkflow = `apiVersion: agentsmesh.io/v1alpha1
kind: Workflow
metadata:
  name: nightly-review
  namespace: acme
  displayName: Nightly review
  labels: {}
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: delivery-review
  inputs: {}
  executionMode: direct
  sandboxStrategy: fresh
  sessionPersistence: false
  concurrencyPolicy: skip
  maxConcurrentRuns: 1
  maxRetainedRuns: 30
  timeoutMinutes: 60
  idleTimeoutSeconds: 30
`;

describe("WorkflowRevisionDialog", () => {
  it("loads the active declaration into the single resource draft", async () => {
    exportResource.mockResolvedValue(new TextEncoder().encode(exportedWorkflow));
    const onApplied = vi.fn();

    render(
      <WorkflowRevisionDialog
        open
        orgSlug="acme"
        workflowSlug="nightly-review"
        onOpenChange={() => {}}
        onApplied={onApplied}
      />,
    );

    const editor = await screen.findByTestId("revision-editor");
    expect(exportResource).toHaveBeenCalledWith(
      "acme",
      {
        apiVersion: "agentsmesh.io/v1alpha1",
        kind: "Workflow",
        namespace: "acme",
        name: "nightly-review",
      },
      expect.any(Number),
    );
    expect(editor).toHaveAttribute("data-name", "nightly-review");
    expect(editor).toHaveAttribute("data-prompt", "delivery-review");
    expect(editor).toHaveAttribute("data-locked-kind", "Workflow");
    expect(editor).toHaveAttribute("data-locked-namespace", "acme");
    expect(editor).toHaveAttribute("data-locked-name", "nightly-review");

    fireEvent.click(editor);
    expect(onApplied).toHaveBeenCalledTimes(1);
  });

  it("shows an explicit load failure without opening an empty editor", async () => {
    exportResource.mockRejectedValue(new Error("resource not found"));

    render(
      <WorkflowRevisionDialog
        open
        orgSlug="acme"
        workflowSlug="nightly-review"
        onOpenChange={() => {}}
        onApplied={() => {}}
      />,
    );

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(
        "workflows.revisionLoadFailed",
      );
    });
    expect(screen.queryByTestId("revision-editor")).not.toBeInTheDocument();
  });
});
