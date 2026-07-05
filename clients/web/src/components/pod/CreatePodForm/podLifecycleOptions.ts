export type DestroyPolicy = "manual" | "idle" | "completed";

export const destroyPolicyOptions: Array<{
  value: DestroyPolicy;
  label: string;
  description: string;
}> = [
  {
    value: "manual",
    label: "手动销毁",
    description: "Pod 保持可访问，直到用户手动终止。",
  },
  {
    value: "idle",
    label: "空闲后销毁",
    description: "适合临时调试和短任务，减少闲置资源占用。",
  },
  {
    value: "completed",
    label: "完成后销毁",
    description: "适合一次性任务，Agent 完成后进入清理流程。",
  },
];

export const destroyAfterOptions = [
  { value: "30", label: "30 分钟" },
  { value: "120", label: "2 小时" },
  { value: "480", label: "8 小时" },
  { value: "1440", label: "24 小时" },
];

export function lifecycleSummary(policy: DestroyPolicy, minutes: number): string {
  if (policy === "manual") return "手动销毁";
  const label = destroyAfterOptions.find((o) => Number(o.value) === minutes)?.label;
  return `${policy === "idle" ? "空闲" : "完成"}后 ${label ?? `${minutes} 分钟`} 销毁`;
}
