"use client";

import { useEffect, useEffectEvent, useRef, useState } from "react";
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
  const appliedInitialExpertRef = useRef<string | null>(null);
  const selectionChangedRef = useRef(false);
  const currentOrgSlug = currentOrg?.slug;

  const applyInitialExpert = useEffectEvent((expert: Expert) => {
    applyExpertToForm(expert, form, setSelectedRunnerId);
    onExpertSelected?.(expert);
  });

  useEffect(() => {
    if (!currentOrgSlug) return;
    let cancelled = false;

    async function loadExperts() {
      const loadedExperts = await fetchExperts();
      if (cancelled || !initialExpertSlug || selectionChangedRef.current) return;
      const initializationKey = `${currentOrgSlug}:${initialExpertSlug}`;
      if (appliedInitialExpertRef.current === initializationKey) return;
      const match = loadedExperts.find((expert) => expert.slug === initialExpertSlug);
      if (!match) return;
      appliedInitialExpertRef.current = initializationKey;
      setSelectedSlug(match.slug);
      applyInitialExpert(match);
    }

    void loadExperts();
    return () => {
      cancelled = true;
    };
  }, [currentOrgSlug, fetchExperts, initialExpertSlug]);

  const handleChange = (value: string) => {
    selectionChangedRef.current = true;
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
