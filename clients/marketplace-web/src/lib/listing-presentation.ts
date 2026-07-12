import type { ResourceType } from "./marketplace-types";

export const resourceTypeLabels: Record<ResourceType, string> = {
  application: "应用",
  skill: "Skill",
  mcp_connector: "系统连接",
  resource: "资源",
};

export const acquireLabels: Record<ResourceType, string> = {
  application: "启用应用",
  skill: "安装 Skill",
  mcp_connector: "连接系统",
  resource: "申请使用",
};

export function formatQuotaSummary(
  quota: { mode: string; estimated_credits_micro: string } | undefined,
): string | null {
  if (!quota || quota.mode !== "per_install") return null;
  try {
    const micro = BigInt(quota.estimated_credits_micro);
    const scale = BigInt(1_000_000);
    if (micro <= BigInt(0)) return null;
    const whole = micro / scale;
    const fraction = (micro % scale)
      .toString()
      .padStart(6, "0")
      .replace(/0+$/, "");
    const credits = fraction ? `${whole}.${fraction}` : whole.toString();
    return `启用需 ${credits} 市场额度`;
  } catch {
    return null;
  }
}
