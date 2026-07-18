import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { WorkflowCreateDialog } from "../WorkflowCreateDialog";

vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme" }),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/components/resource-editor/ResourceEditorShell", () => ({
  ResourceEditorShell: ({
    kind,
    orgSlug,
    onApplied,
  }: {
    kind: string;
    orgSlug: string;
    onApplied: () => void;
  }) => (
    <button
      type="button"
      data-testid="resource-editor"
      data-kind={kind}
      data-org={orgSlug}
      onClick={onApplied}
    >
      apply resource
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

describe("WorkflowCreateDialog", () => {
  it("creates Workflow declarations through the resource editor", () => {
    const onOpenChange = vi.fn();
    const onCreated = vi.fn();

    render(
      <WorkflowCreateDialog
        open
        onOpenChange={onOpenChange}
        onCreated={onCreated}
      />,
    );

    const editor = screen.getByTestId("resource-editor");
    expect(editor).toHaveAttribute("data-kind", "Workflow");
    expect(editor).toHaveAttribute("data-org", "acme");

    fireEvent.click(editor);

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onCreated).toHaveBeenCalledTimes(1);
  });
});
