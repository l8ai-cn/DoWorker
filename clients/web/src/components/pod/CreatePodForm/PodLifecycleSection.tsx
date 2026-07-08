"use client";

import { useTranslations } from "next-intl";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  destroyAfterOptions,
  destroyPolicyOptions,
  type DestroyPolicy,
} from "./podLifecycleOptions";

interface PodLifecycleSectionProps {
  destroyPolicy: DestroyPolicy;
  destroyAfterMinutes: number;
  onPolicyChange: (policy: DestroyPolicy) => void;
  onAfterChange: (minutes: number) => void;
}

export function PodLifecycleSection({
  destroyPolicy,
  destroyAfterMinutes,
  onPolicyChange,
  onAfterChange,
}: PodLifecycleSectionProps) {
  const t = useTranslations();
  const selected = destroyPolicyOptions.find((o) => o.value === destroyPolicy);

  return (
    <section className="rounded-lg border border-border bg-surface-muted/35 p-3">
      <div className="mb-3">
        <h3 className="text-sm font-medium text-foreground">
          {t("ide.createPod.lifecycleTitle")}
        </h3>
        <p className="text-xs leading-5 text-muted-foreground">
          {t("ide.createPod.lifecycleDescription")}
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            {t("ide.createPod.lifecyclePolicyLabel")}
          </label>
          <Select
            value={destroyPolicy}
            onValueChange={(value) => onPolicyChange(value as DestroyPolicy)}
          >
            <SelectTrigger>
              <SelectValue placeholder={t("ide.createPod.lifecyclePolicyPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {destroyPolicyOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {t(option.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {selected && (
            <p className="mt-1 text-xs text-muted-foreground">
              {t(selected.descriptionKey)}
            </p>
          )}
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            {t("ide.createPod.lifecycleAfterLabel")}
          </label>
          <Select
            value={String(destroyAfterMinutes)}
            onValueChange={(value) => onAfterChange(Number(value))}
            disabled={destroyPolicy === "manual"}
          >
            <SelectTrigger>
              <SelectValue placeholder={t("ide.createPod.lifecycleAfterPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {destroyAfterOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {t(option.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="mt-1 text-xs text-muted-foreground">
            {t("ide.createPod.lifecycleManualHint")}
          </p>
        </div>
      </div>
    </section>
  );
}
