import Link from "next/link";
import { Loader2, RefreshCw, Settings2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { MarketplaceModelResource } from "@/lib/marketplace-model-resources";

export function MarketplaceModelResourceField({
  resources,
  value,
  onChange,
  loading,
  error,
  incompatibleListing,
  onReload,
  settingsHref,
}: {
  resources: MarketplaceModelResource[];
  value: string;
  onChange: (value: string) => void;
  loading: boolean;
  error: boolean;
  incompatibleListing: boolean;
  onReload: () => void;
  settingsHref: string;
}) {
  if (incompatibleListing) {
    return (
      <div className="rounded-lg border border-danger/30 bg-danger-bg p-5 text-sm text-foreground">
        当前专家版本缺少兼容 Agent，请联系发布者修正后重新上架。
      </div>
    );
  }
  if (loading) {
    return (
      <div className="flex min-h-40 items-center justify-center gap-3 text-sm text-muted-foreground">
        <Loader2 className="h-5 w-5 animate-spin" />
        正在加载兼容模型
      </div>
    );
  }
  if (error) {
    return (
      <Button className="w-full gap-2" variant="outline" onClick={onReload}>
        <RefreshCw className="h-4 w-4" />
        重新加载模型
      </Button>
    );
  }
  if (resources.length === 0) {
    return (
      <Button asChild className="w-full gap-2" variant="outline">
        <Link href={settingsHref}>
          <Settings2 className="h-4 w-4" />
          配置兼容模型
        </Link>
      </Button>
    );
  }
  return (
    <select
      value={value}
      onChange={(event) => onChange(event.target.value)}
      className="h-12 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
      aria-label="选择运行模型"
    >
      <option value="">请选择运行模型</option>
      {resources.map((resource) => (
        <option key={resource.id} value={resource.id}>
          {resource.label}
        </option>
      ))}
    </select>
  );
}
