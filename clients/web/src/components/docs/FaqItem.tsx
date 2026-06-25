"use client";

import { useTranslations } from "next-intl";

interface FaqItemProps {
  questionKey: string;
  answerKey: string;
}

export function FaqItem({ questionKey, answerKey }: FaqItemProps) {
  const t = useTranslations();

  return (
    <details className="surface-card p-4 motion-interactive">
      <summary className="font-medium cursor-pointer">{t(questionKey)}</summary>
      <p className="text-sm text-muted-foreground mt-3">{t(answerKey)}</p>
    </details>
  );
}
