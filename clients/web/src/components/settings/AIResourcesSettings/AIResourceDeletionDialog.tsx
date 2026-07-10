"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { AIResourceDeletionTarget } from "./types";

interface AIResourceDeletionDialogProps {
  target: AIResourceDeletionTarget | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: (target: AIResourceDeletionTarget) => Promise<boolean>;
}

export function AIResourceDeletionDialog({ target, onOpenChange, onConfirm }: AIResourceDeletionDialogProps) {
  const t = useTranslations();
  const [deleting, setDeleting] = useState(false);
  const [failed, setFailed] = useState(false);

  const changeOpen = (open: boolean) => {
    if (!open) setFailed(false);
    onOpenChange(open);
  };

  const confirm = async () => {
    if (!target) return;
    setDeleting(true);
    setFailed(false);
    try {
      if (await onConfirm(target)) changeOpen(false);
      else setFailed(true);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <AlertDialog open={Boolean(target)} onOpenChange={changeOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t("settings.aiResources.deleteConfirm.title")}</AlertDialogTitle>
          {target && <AlertDialogDescription>{t(`settings.aiResources.deleteConfirm.${target.kind}Description`, { name: target.name })}</AlertDialogDescription>}
          {failed && <p role="alert" className="text-sm text-destructive">{t("settings.aiResources.deleteConfirm.error")}</p>}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
          <AlertDialogAction disabled={deleting} onClick={() => void confirm()} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
            {t("settings.aiResources.deleteConfirm.action")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
