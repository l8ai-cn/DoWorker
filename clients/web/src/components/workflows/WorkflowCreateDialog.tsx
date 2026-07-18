"use client";

import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";

interface WorkflowCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}

export function WorkflowCreateDialog({
  open,
  onOpenChange,
  onCreated,
}: WorkflowCreateDialogProps) {
  const t = useTranslations();
  const params = useParams();
  const orgSlug = String(params.org ?? "");

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent
        className="max-h-[90vh] max-w-5xl overflow-y-auto"
      >
        <ResponsiveDialogHeader onClose={() => onOpenChange(false)}>
          <ResponsiveDialogTitle>
            {t("workflows.createWorkflow")}
          </ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        {open && (
          <div className="px-4 pb-4">
            <ResourceEditorShell
              orgSlug={orgSlug}
              kind="Workflow"
              onApplied={() => {
                onOpenChange(false);
                onCreated();
              }}
            />
          </div>
        )}
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
