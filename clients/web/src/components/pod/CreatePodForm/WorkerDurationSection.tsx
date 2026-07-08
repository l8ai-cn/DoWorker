"use client";

import type { CreatePodFormState } from "../hooks";
import { WorkerDurationPreset, type WorkerDurationKind } from "./WorkerDurationPreset";

interface WorkerDurationSectionProps {
  form: CreatePodFormState;
  t: (key: string) => string;
}

export function WorkerDurationSection({ form, t }: WorkerDurationSectionProps) {
  const applyDuration = (kind: WorkerDurationKind) => {
    if (kind === "long") {
      form.setPerpetual(true);
      form.setDestroyPolicy("manual");
      return;
    }
    form.setPerpetual(false);
    form.setDestroyPolicy("idle");
    form.setDestroyAfterMinutes(120);
  };

  return (
    <WorkerDurationPreset
      perpetual={form.perpetual}
      destroyPolicy={form.destroyPolicy}
      onSelect={applyDuration}
      t={t}
    />
  );
}
