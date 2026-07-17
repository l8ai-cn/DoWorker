"use client";

import React from "react";

export const AUTOMATION_LEVELS = ["interactive", "auto_edit", "autonomous"] as const;
export type AutomationLevel = (typeof AUTOMATION_LEVELS)[number];

interface AutomationLevelSelectProps {
  value: string;
  onChange: (level: AutomationLevel) => void;
  supportedModes?: string[];
  t: (key: string) => string;
}

// Unified cross-agent permission/automation tier. The backend adapter maps the
// selected level onto each agent's native permission mechanism at create time.
export function AutomationLevelSelect({
  value,
  onChange,
  supportedModes,
  t,
}: AutomationLevelSelectProps) {
  const active = AUTOMATION_LEVELS.includes(value as AutomationLevel)
    ? (value as AutomationLevel)
    : "autonomous";

  return (
    <div>
      <label className="block text-sm font-medium mb-1.5">
        {t("ide.createPod.automationLevel.label")}
      </label>
      <div className="flex gap-2">
        {AUTOMATION_LEVELS.map((level) => (
          <button
            key={level}
            type="button"
            onClick={() => onChange(level)}
            data-testid={`automation-level-${level}`}
            aria-pressed={active === level}
            className={`flex-1 px-3 py-2 text-sm rounded-md border transition-colors ${
              active === level
                ? "border-primary bg-primary/10 text-primary font-medium"
                : "border-border bg-background text-muted-foreground hover:bg-muted"
            }`}
          >
            {t(`ide.createPod.automationLevel.${level}`)}
          </button>
        ))}
      </div>
      <p className="mt-1.5 text-xs text-muted-foreground">
        {t(
          active === "autonomous" &&
            supportedModes !== undefined &&
            !supportedModes.includes("acp")
            ? "ide.createPod.automationLevel.autonomousPtyHint"
            : `ide.createPod.automationLevel.${active}Hint`,
        )}
      </p>
    </div>
  );
}
