import Link from "next/link";
import { ArrowRight, CalendarDays, CircleCheck, PackageOpen } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { MarketplaceOrganizationApplication } from "@/lib/marketplace/application-api";

const resourceTypeLabels = {
  application: "应用",
  skill: "技能",
  mcp_connector: "MCP 连接器",
  resource: "资源",
} as const;

const statusLabels = {
  installing: { label: "正在启用", variant: "warning" as const },
  verifying: { label: "正在验证", variant: "info" as const },
  active: { label: "可使用", variant: "success" as const },
} as const;

export function ApplicationCard({
  application,
  expertSlug,
  orgSlug,
}: {
  application: MarketplaceOrganizationApplication;
  expertSlug?: string;
  orgSlug: string;
}) {
  const status = statusLabels[application.status];
  const canStart = application.status === "active" && expertSlug;

  return (
    <article className="flex min-h-64 flex-col rounded-xl border border-border bg-card p-5 shadow-xs">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <PackageOpen className="h-4 w-4 shrink-0 text-primary" />
            <Badge variant="outline">{resourceTypeLabels[application.resource_type]}</Badge>
          </div>
          <h2 className="mt-4 text-lg font-semibold text-foreground">{application.display_name}</h2>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{application.tagline}</p>
        </div>
        <Badge variant={status.variant}>{status.label}</Badge>
      </div>

      {application.outcomes[0] ? (
        <p className="mt-5 border-l-2 border-primary/40 pl-3 text-sm text-foreground">
          可完成：{application.outcomes[0]}
        </p>
      ) : null}

      <div className="mt-auto space-y-4 pt-6">
        <p className="flex items-center gap-2 text-xs text-muted-foreground">
          <CalendarDays className="h-3.5 w-3.5" />
          启用时间：{formatInstalledAt(application.installed_at)}
        </p>
        {canStart ? (
          <Button asChild className="w-full gap-2">
            <Link href={`/${orgSlug}/experts/${expertSlug}`}>
              开始第一个任务
              <ArrowRight className="h-4 w-4" />
            </Link>
          </Button>
        ) : (
          <EnableDetails application={application} statusLabel={status.label} />
        )}
      </div>
    </article>
  );
}

function EnableDetails({
  application,
  statusLabel,
}: {
  application: MarketplaceOrganizationApplication;
  statusLabel: string;
}) {
  return (
    <details className="rounded-lg border border-border bg-surface-muted px-3 py-2 text-sm text-muted-foreground">
      <summary className="cursor-pointer font-medium text-foreground">查看启用详情</summary>
      <div className="mt-3 space-y-1.5">
        <p>当前状态：{statusLabel}</p>
        <p className="break-all">安装编号：{application.installation_id}</p>
        {application.runtime_ref ? <p>运行时：{application.runtime_ref}</p> : null}
        {!application.runtime_ref ? (
          <p className="flex items-center gap-1.5">
            <CircleCheck className="h-3.5 w-3.5" />
            此资源暂不提供可启动的专家任务入口。
          </p>
        ) : null}
      </div>
    </details>
  );
}

function formatInstalledAt(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "未提供";
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
