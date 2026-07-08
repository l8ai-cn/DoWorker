export type DestroyPolicy = "manual" | "idle" | "completed";

export const destroyPolicyOptions: Array<{
  value: DestroyPolicy;
  labelKey: string;
  descriptionKey: string;
}> = [
  {
    value: "manual",
    labelKey: "ide.createPod.lifecyclePolicy.manual",
    descriptionKey: "ide.createPod.lifecyclePolicy.manualDesc",
  },
  {
    value: "idle",
    labelKey: "ide.createPod.lifecyclePolicy.idle",
    descriptionKey: "ide.createPod.lifecyclePolicy.idleDesc",
  },
  {
    value: "completed",
    labelKey: "ide.createPod.lifecyclePolicy.completed",
    descriptionKey: "ide.createPod.lifecyclePolicy.completedDesc",
  },
];

export const destroyAfterOptions = [
  { value: "30", labelKey: "ide.createPod.lifecycleAfter.30m" },
  { value: "120", labelKey: "ide.createPod.lifecycleAfter.2h" },
  { value: "480", labelKey: "ide.createPod.lifecycleAfter.8h" },
  { value: "1440", labelKey: "ide.createPod.lifecycleAfter.24h" },
];

export function lifecycleSummary(
  policy: DestroyPolicy,
  minutes: number,
  t: (key: string) => string,
): string {
  if (policy === "manual") return t("ide.createPod.lifecyclePolicy.manual");
  const afterLabel =
    destroyAfterOptions.find((o) => Number(o.value) === minutes)?.labelKey;
  const prefix =
    policy === "idle"
      ? t("ide.createPod.lifecycleSummary.idlePrefix")
      : t("ide.createPod.lifecycleSummary.completedPrefix");
  return afterLabel
    ? `${prefix} ${t(afterLabel)}`
    : `${prefix} ${minutes}m`;
}
