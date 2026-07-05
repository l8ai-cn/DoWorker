"use client";

import { buttonVariants } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { SSOConfig } from "@/lib/api/sso";

interface SSODeleteDialogProps {
  config: SSOConfig | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function SSODeleteDialog({ config, onOpenChange, onConfirm }: SSODeleteDialogProps) {
  return (
    <AlertDialog open={!!config} onOpenChange={(open) => !open && onOpenChange(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>删除 SSO 配置</AlertDialogTitle>
          <AlertDialogDescription>
            确定要删除 SSO 配置 &quot;{config?.name}&quot; ({config?.domain}) 吗？此操作无法撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className={buttonVariants({ variant: "destructive" })}
          >
            删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
