"use client";

import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import { useTranslations } from "next-intl";
import type { WorkflowData } from "@/lib/viewModels/workflow";
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
  const title = editWorkflow ? t("workflows.editWorkflow") : t("workflows.createWorkflow");

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
        <ResponsiveDialogHeader onClose={() => onOpenChange(false)}>
          <ResponsiveDialogTitle>{title}</ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        {open && (
          <div className="px-4 pb-4">
            <WorkflowCreateContent
            key={editWorkflow?.slug ?? "create"}
            editWorkflow={editWorkflow}
            showAiSection={!editWorkflow}
            onCreated={onCreated}
            onCancel={() => onOpenChange(false)}
            />
          </div>
        )}
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
