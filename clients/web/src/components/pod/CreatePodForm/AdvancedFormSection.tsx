"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { PodLifecycleSection } from "./PodLifecycleSection";
import type { CreatePodFormState } from "../hooks";

interface AdvancedFormSectionProps {
  form: CreatePodFormState;
}

export function AdvancedFormSection({
  form,
}: AdvancedFormSectionProps) {
  const t = useTranslations();

  return (
    <div className="space-y-4">
      <div>
        <label htmlFor="pod-alias" className="mb-1 block text-sm font-medium">
          {t("ide.createPod.alias")}
        </label>
        <Input
          id="pod-alias"
          value={form.alias}
          onChange={(e) => form.setAlias(e.target.value)}
          placeholder={t("ide.createPod.aliasPlaceholder")}
          maxLength={100}
        />
      </div>

      <PodLifecycleSection
        destroyPolicy={form.destroyPolicy}
        destroyAfterMinutes={form.destroyAfterMinutes}
        onPolicyChange={form.setDestroyPolicy}
        onAfterChange={form.setDestroyAfterMinutes}
      />
    </div>
  );
}
