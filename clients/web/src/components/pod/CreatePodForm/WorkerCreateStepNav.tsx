"use client";

import { Button } from "@/components/ui/button";
import type { WorkerCreateStepId } from "./WorkerCreateStepper";

interface WorkerCreateStepNavProps {
  step: WorkerCreateStepId;
  canNext: boolean;
  onBack: () => void;
  onNext: () => void;
  t: (key: string) => string;
}

export function WorkerCreateStepNav({
  step,
  canNext,
  onBack,
  onNext,
  t,
}: WorkerCreateStepNavProps) {
  return (
    <div className="mt-6 flex items-center justify-between gap-3 border-t border-border pt-4">
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={onBack}
        disabled={step === 1}
        className={step === 1 ? "invisible" : ""}
      >
        {t("ide.createPod.stepNavBack")}
      </Button>
      {step < 4 && (
        <Button type="button" size="sm" onClick={onNext} disabled={!canNext}>
          {t("ide.createPod.stepNavNext")}
        </Button>
      )}
    </div>
  );
}
