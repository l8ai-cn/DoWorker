"use client";

import { useTranslations } from "next-intl";

interface DocStepHeaderProps {
  step: number;
  titleKey: string;
}

export function DocStepHeader({ step, titleKey }: DocStepHeaderProps) {
  const t = useTranslations();

  return (
    <div className="flex items-center gap-3 mb-4">
      <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-bold">
        {step}
      </div>
      <h2 className="text-xl font-semibold text-foreground">{t(titleKey)}</h2>
    </div>
  );
}
