import type { MarketplaceResourceType } from "./catalog-api";

export const marketplaceTypeLabels: Record<MarketplaceResourceType, string> = {
  application: "应用",
  skill: "Skill",
  mcp_connector: "系统连接",
  resource: "资源",
};

export function formatMarketplaceCredits(
  quota: { mode: string; estimated_credits_micro: string } | undefined,
): string | null {
  if (!quota || quota.mode !== "per_install") return null;
  try {
    const credits = BigInt(quota.estimated_credits_micro);
    const zero = BigInt(0);
    const scale = BigInt(1_000_000);
    if (credits <= zero) return null;
    const whole = credits / scale;
    const decimal = (credits % scale)
      .toString()
      .padStart(6, "0")
      .replace(/0+$/, "");
    return decimal ? `${whole}.${decimal} 市场额度` : `${whole} 市场额度`;
  } catch {
    return null;
  }
}
