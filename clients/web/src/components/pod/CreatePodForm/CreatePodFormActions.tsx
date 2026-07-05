"use client";

import { Button } from "@/components/ui/button";

interface Props {
  onCancel?: () => void;
  onCreate: () => void;
  disabled: boolean;
  loading: boolean;
  t: (key: string) => string;
}

export function CreatePodFormActions({
  onCancel,
  onCreate,
  disabled,
  loading,
  t,
}: Props) {
  return (
    <div className="flex flex-col-reverse sm:flex-row justify-end gap-3 mt-6">
      {onCancel && (
        <Button variant="outline" onClick={onCancel} className="w-full sm:w-auto">
          {t("ide.createPod.cancel")}
        </Button>
      )}
      <Button onClick={onCreate} disabled={disabled} className="w-full sm:w-auto">
        {loading ? t("ide.createPod.creating") : t("ide.createPod.create")}
      </Button>
    </div>
  );
}
