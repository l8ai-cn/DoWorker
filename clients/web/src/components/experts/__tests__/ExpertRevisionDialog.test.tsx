import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ExpertRevisionDialog } from "../ExpertRevisionDialog";

const exportResource = vi.fn();
const getResourceCapabilities = vi.fn();

vi.mock("@/lib/api/facade/orchestrationResource", () => ({
  exportResource: (...args: unknown[]) => exportResource(...args),
  getResourceCapabilities: (...args: unknown[]) =>
    getResourceCapabilities(...args),
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
      spec: { workerTemplateRef: { name: string } };
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
      data-worker-template={initialDraft.spec.workerTemplateRef.name}
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

const exportedExpert = `apiVersion: agentsmesh.io/v1alpha1
kind: Expert
metadata:
  name: release-reviewer
  namespace: acme
  displayName: Release reviewer
  labels: {}
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: release-review
  description: Reviews release candidates
  category: engineering
  releaseNotes: Initial release
`;

describe("ExpertRevisionDialog", () => {
  it("loads the active declaration into a locked resource draft", async () => {
    getResourceCapabilities.mockResolvedValue({
      capabilities: { canPlan: true, canViewSource: true },
    });
    exportResource.mockResolvedValue(new TextEncoder().encode(exportedExpert));
    const onApplied = vi.fn();

    render(
      <ExpertRevisionDialog
        open
        orgSlug="acme"
        expertSlug="release-reviewer"
        onOpenChange={() => {}}
        onApplied={onApplied}
      />,
    );

    const editor = await screen.findByTestId("revision-editor");
    expect(exportResource).toHaveBeenCalledWith(
      "acme",
      {
        apiVersion: "agentsmesh.io/v1alpha1",
        kind: "Expert",
        namespace: "acme",
        name: "release-reviewer",
      },
      expect.any(Number),
    );
    expect(editor).toHaveAttribute("data-name", "release-reviewer");
    expect(editor).toHaveAttribute(
      "data-worker-template",
      "codex-reviewer",
    );
    expect(editor).toHaveAttribute("data-locked-kind", "Expert");
    expect(editor).toHaveAttribute("data-locked-namespace", "acme");
    expect(editor).toHaveAttribute("data-locked-name", "release-reviewer");

    fireEvent.click(editor);
    expect(onApplied).toHaveBeenCalledTimes(1);
  });

  it("shows an explicit load failure without opening an empty editor", async () => {
    getResourceCapabilities.mockResolvedValue({
      capabilities: { canPlan: true, canViewSource: true },
    });
    exportResource.mockRejectedValue(new Error("resource not found"));

    render(
      <ExpertRevisionDialog
        open
        orgSlug="acme"
        expertSlug="release-reviewer"
        onOpenChange={() => {}}
        onApplied={() => {}}
      />,
    );

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(
        "experts.revisionLoadFailed",
      );
    });
    expect(screen.queryByTestId("revision-editor")).not.toBeInTheDocument();
  });

  it("does not export or open an editor when revision planning is denied", async () => {
    getResourceCapabilities.mockResolvedValue({
      capabilities: { canPlan: false, canViewSource: true },
    });

    render(
      <ExpertRevisionDialog
        open
        orgSlug="acme"
        expertSlug="release-reviewer"
        onOpenChange={() => {}}
        onApplied={() => {}}
      />,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "experts.revisionPermissionDenied",
    );
    expect(exportResource).not.toHaveBeenCalled();
    expect(screen.queryByTestId("revision-editor")).not.toBeInTheDocument();
  });

  it("does not export when source viewing is denied", async () => {
    getResourceCapabilities.mockResolvedValue({
      capabilities: { canPlan: true, canViewSource: false },
    });

    render(
      <ExpertRevisionDialog
        open
        orgSlug="acme"
        expertSlug="release-reviewer"
        onOpenChange={() => {}}
        onApplied={() => {}}
      />,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "experts.revisionPermissionDenied",
    );
    expect(exportResource).not.toHaveBeenCalled();
    expect(screen.queryByTestId("revision-editor")).not.toBeInTheDocument();
  });

  it("rejects an exported declaration with a different identity", async () => {
    getResourceCapabilities.mockResolvedValue({
      capabilities: { canPlan: true, canViewSource: true },
    });
    exportResource.mockResolvedValue(new TextEncoder().encode(
      exportedExpert.replace("name: release-reviewer", "name: other-expert"),
    ));

    render(
      <ExpertRevisionDialog
        open
        orgSlug="acme"
        expertSlug="release-reviewer"
        onOpenChange={() => {}}
        onApplied={() => {}}
      />,
    );

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(
        "experts.revisionLoadFailed",
      );
    });
    expect(screen.queryByTestId("revision-editor")).not.toBeInTheDocument();
  });
});
