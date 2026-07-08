"use client";

import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import { useTranslations } from "next-intl";
import type { LoopData } from "@/lib/viewModels/loop";
import { LoopCreateContent } from "./LoopCreateContent";

interface LoopCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: (createdLoop?: LoopData) => void;
  editLoop?: LoopData;
}

/** Dialog shell around the linked AI + manual Loop create flow. */
export function LoopCreateDialog({
  open,
  onOpenChange,
  onCreated,
  editLoop,
}: LoopCreateDialogProps) {
  const t = useTranslations();
  const title = editLoop ? t("loops.editLoop") : t("loops.createLoop");

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
        <ResponsiveDialogHeader onClose={() => onOpenChange(false)}>
          <ResponsiveDialogTitle>{title}</ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        {open && (
          <div className="px-4 pb-4">
            <LoopCreateContent
            key={editLoop?.slug ?? "create"}
            editLoop={editLoop}
            showAiSection={!editLoop}
            onCreated={onCreated}
            onCancel={() => onOpenChange(false)}
            />
          </div>
        )}
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
