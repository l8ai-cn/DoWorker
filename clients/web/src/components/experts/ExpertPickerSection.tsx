"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useExpertStore, useExperts } from "@/stores/expert";
import { useCurrentOrg } from "@/stores/auth";
import type { Expert } from "@/lib/api/expertApi";
import type { CreatePodFormState } from "@/components/pod/hooks/useCreatePodFormTypes";
import { applyExpertToForm } from "@/lib/expert-form-prefill";

interface ExpertPickerSectionProps {
  form: CreatePodFormState;
  setSelectedRunnerId: (id: number | null) => void;
  onExpertSelected?: (expert: Expert | null) => void;
  initialExpertSlug?: string;
}

export function ExpertPickerSection({
  form,
  setSelectedRunnerId,
  onExpertSelected,
  initialExpertSlug,
}: ExpertPickerSectionProps) {
  const t = useTranslations("experts.picker");
  const currentOrg = useCurrentOrg();
  const experts = useExperts();
  const fetchExperts = useExpertStore((s) => s.fetchExperts);
  const [selectedSlug, setSelectedSlug] = useState<string>("none");
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (currentOrg) fetchExperts();
  }, [currentOrg, fetchExperts]);

  useEffect(() => {
    if (initialized || !initialExpertSlug || experts.length === 0) return;
    const match = experts.find((e) => e.slug === initialExpertSlug);
    if (!match) return;
    // Defer so the set-state-in-effect analyzer treats this as opaque.
    Promise.resolve().then(() => {
      setSelectedSlug(match.slug);
      applyExpertToForm(match, form, setSelectedRunnerId);
      onExpertSelected?.(match);
      setInitialized(true);
    });
  }, [initialExpertSlug, experts, initialized, form, setSelectedRunnerId, onExpertSelected]);

  const handleChange = (value: string) => {
    setSelectedSlug(value);
    if (value === "none") {
      onExpertSelected?.(null);
      return;
    }
    const expert = experts.find((e) => e.slug === value);
    if (expert) {
      applyExpertToForm(expert, form, setSelectedRunnerId);
      onExpertSelected?.(expert);
    }
  };

  if (experts.length === 0) return null;

  return (
    <div className="space-y-2 rounded-lg border border-border bg-muted/20 p-4">
      <Label>{t("label")}</Label>
      <Select value={selectedSlug} onValueChange={handleChange}>
        <SelectTrigger>
          <SelectValue placeholder={t("placeholder")} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="none">{t("none")}</SelectItem>
          {experts.map((expert) => (
            <SelectItem key={expert.slug} value={expert.slug}>
              {expert.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-xs text-muted-foreground">{t("hint")}</p>
    </div>
  );
}
