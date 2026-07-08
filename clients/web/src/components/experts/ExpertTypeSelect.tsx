"use client";

import { useTranslations } from "next-intl";

// Expert type (类型) is a free-form string on the backend; the UI offers a
// curated set of categories. The selected key is sent verbatim as expert_type.
export const EXPERT_TYPES = [
  "general",
  "coding",
  "analysis",
  "writing",
  "research",
  "automation",
] as const;

const SELECT_CLASS =
  "h-9 w-full rounded-md bg-surface-raised px-3 text-sm ring-1 ring-border/35 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35";

interface Props {
  value: string;
  onChange: (value: string) => void;
  id?: string;
}

export function ExpertTypeSelect({ value, onChange, id }: Props) {
  const t = useTranslations("experts.create");

  return (
    <select
      id={id}
      className={SELECT_CLASS}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      <option value="">{t("typePlaceholder")}</option>
      {EXPERT_TYPES.map((type) => (
        <option key={type} value={type}>
          {t(`typeOptions.${type}`)}
        </option>
      ))}
    </select>
  );
}
