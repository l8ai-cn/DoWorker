"use client";

import { useState } from "react";
import { ArrowRight, Loader2, RefreshCw, Settings2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { installMarketplaceApplication } from "@/lib/marketplace-install";
import { MarketplaceToolModelFields } from "./MarketplaceToolModelFields";
import { useMarketplaceInstallResources } from "./useMarketplaceInstallResources";

interface MarketplaceInstallActionProps {
  applicationSlug: string;
  agentSlug: string;
  orgSlug: string | null;
  onInstalled: (orgSlug: string, expertSlug: string, existing: boolean) => void;
  onNeedsOrganization: () => void;
  onConfigureResources: (orgSlug: string) => void;
}

export function MarketplaceInstallAction(props: MarketplaceInstallActionProps) {
  const [installing, setInstalling] = useState(false);
  const runtime = useMarketplaceInstallResources(
    props.orgSlug,
    props.agentSlug,
  );

  async function install() {
    if (!runtime.orgSlug || !runtime.selectionComplete) return;
    setInstalling(true);
    try {
      const result = await installMarketplaceApplication(
        runtime.orgSlug,
        props.applicationSlug,
        Number(runtime.modelID),
        numericToolModelIDs(runtime.toolIDs),
      );
      props.onInstalled(
        runtime.orgSlug,
        result.expert.slug,
        result.already_installed,
      );
    } catch {
      toast.error("启用失败，请稍后重试");
    } finally {
      setInstalling(false);
    }
  }

  if (runtime.error) {
    return (
      <Button
        className="w-full gap-2"
        variant="outline"
        onClick={runtime.reload}
      >
        <RefreshCw className="h-4 w-4" />
        重新加载模型
      </Button>
    );
  }
  if (!runtime.loading && !runtime.orgSlug) {
    return (
      <Button className="w-full gap-2" onClick={props.onNeedsOrganization}>
        创建组织后启用
        <ArrowRight className="h-4 w-4" />
      </Button>
    );
  }
  if (!runtime.loading && runtime.missingCompatibleResource) {
    return (
      <Button
        className="w-full gap-2"
        variant="outline"
        onClick={() =>
          runtime.orgSlug && props.onConfigureResources(runtime.orgSlug)
        }
      >
        <Settings2 className="h-4 w-4" />
        配置兼容模型
      </Button>
    );
  }

  const selectedLabel = runtime.models.find(
    (resource) => String(resource.id) === runtime.modelID,
  )?.label;
  return (
    <div className="space-y-2">
      <Select
        value={runtime.modelID}
        onValueChange={runtime.setModelID}
        disabled={runtime.loading || installing}
      >
        <SelectTrigger aria-label="选择运行模型" className="h-10">
          <span className={selectedLabel ? "" : "text-muted-foreground"}>
            {runtime.loading ? "正在加载模型" : selectedLabel || "选择运行模型"}
          </span>
        </SelectTrigger>
        <SelectContent>
          {runtime.models.map((resource) => (
            <SelectItem key={resource.id} value={String(resource.id)}>
              {resource.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <MarketplaceToolModelFields
        groups={runtime.tools}
        values={runtime.toolIDs}
        onChange={runtime.setToolID}
        disabled={runtime.loading || installing}
      />
      <Button
        className="w-full gap-2"
        onClick={install}
        disabled={installing || runtime.loading || !runtime.selectionComplete}
      >
        {installing ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
        {installing ? "正在启用" : "立即启用"}
        {!installing ? <ArrowRight className="h-4 w-4" /> : null}
      </Button>
    </div>
  );
}

function numericToolModelIDs(
  values: Record<string, string>,
): Record<string, number> {
  return Object.fromEntries(
    Object.entries(values).map(([role, id]) => [role, Number(id)]),
  );
}
