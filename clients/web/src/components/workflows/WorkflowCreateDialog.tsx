"use client";

import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import type { WorkflowData } from "@/lib/viewModels/workflow";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";
import { WorkflowCreateContent } from "./WorkflowCreateContent";

interface WorkflowCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: (createdWorkflow?: WorkflowData) => void;
  editWorkflow?: WorkflowData;
}

/** Dialog shell around the linked AI + manual Workflow create flow. */
export function WorkflowCreateDialog({
  open,
  onOpenChange,
  onCreated,
  editWorkflow,
}: WorkflowCreateDialogProps) {
  const t = useTranslations();
  const params = useParams();
  const orgSlug = String(params.org ?? "");
  const title = editWorkflow ? t("workflows.editWorkflow") : t("workflows.createWorkflow");

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent
        className={editWorkflow
          ? "max-h-[90vh] max-w-2xl overflow-y-auto"
          : "max-h-[90vh] max-w-5xl overflow-y-auto"}
      >
        <ResponsiveDialogHeader onClose={() => onOpenChange(false)}>
          <ResponsiveDialogTitle>{title}</ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        {open && (
          <div className="px-4 pb-4">
            {editWorkflow ? (
              <WorkflowCreateContent
                key={editWorkflow.slug}
                editWorkflow={editWorkflow}
                showAiSection={false}
                onCreated={onCreated}
                onCancel={() => onOpenChange(false)}
              />
            ) : (
              <ResourceEditorShell
                orgSlug={orgSlug}
                kind="Workflow"
                onApplied={() => {
                  onOpenChange(false);
                  onCreated();
                }}
              />
            )}
          </div>
        )}
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
