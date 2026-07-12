import { BadgeCheck, Building2, Gauge, ShieldCheck } from "lucide-react";

import type {
  InstallationPlan,
  MarketplaceListingDetail,
} from "@/lib/marketplace/acquire-api";
import { formatMarketplaceCredits } from "@/lib/marketplace/presentation";

interface Props {
  listing: MarketplaceListingDetail;
  organizationName: string;
  plan: InstallationPlan;
}

export function MarketplaceAcquireSummary({
  listing,
  organizationName,
  plan,
}: Props) {
  const credits = formatMarketplaceCredits({
    mode: "per_install",
    estimated_credits_micro: plan.plan.estimated_credits_micro,
  }) ?? "以实际结算为准";

  return (
    <section className="space-y-5" aria-labelledby="confirm-title">
      <div>
        <p className="text-sm font-medium text-primary">启用确认</p>
        <h2 id="confirm-title" className="mt-1 text-2xl font-semibold text-foreground">
          确认权限与额度
        </h2>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          计划将在 {new Date(plan.plan.expires_at).toLocaleTimeString("zh-CN")} 前有效。
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <SummaryItem icon={Building2} label="目标组织" value={organizationName} />
        <SummaryItem icon={Gauge} label="预计启用额度" value={credits} />
        <SummaryItem icon={BadgeCheck} label="应用版本" value={`v${listing.version}`} />
        <SummaryItem icon={ShieldCheck} label="权限数量" value={`${plan.plan.required_permissions.length} 项`} />
      </div>
      <div className="rounded-lg border border-warning/30 bg-warning-bg p-4">
        <p className="text-sm font-medium text-foreground">
          此应用可以执行工作任务并修改授权范围内的仓库内容。
        </p>
        <p className="mt-1 text-sm leading-6 text-muted-foreground">
          权限只在「{organizationName}」内生效，启用前不会创建运行实例。
        </p>
      </div>
    </section>
  );
}

function SummaryItem({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Building2;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-lg border border-border bg-background p-4">
      <Icon className="h-4 w-4 text-primary" />
      <p className="mt-3 text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-sm font-medium text-foreground">{value}</p>
    </div>
  );
}
