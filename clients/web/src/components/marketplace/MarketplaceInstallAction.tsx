"use client";

import { useEffect, useState } from "react";
import { ArrowRight, Loader2, RefreshCw, Settings2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { fetchFirstOrgSlug } from "@/lib/light-auth";
import { updateLightSessionOrgSlug } from "@/lib/light-session";
import { installMarketplaceApplication } from "@/lib/marketplace-install";
import {
  listMarketplaceModelResources,
  type MarketplaceModelResource,
} from "@/lib/marketplace-model-resources";

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
  const [targetOrgSlug, setTargetOrgSlug] = useState(props.orgSlug);
  const [resources, setResources] = useState<MarketplaceModelResource[]>([]);
  const [selectedResourceID, setSelectedResourceID] = useState("");
  const [loadingResources, setLoadingResources] = useState(true);
  const [resourceError, setResourceError] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    async function loadResources() {
      setLoadingResources(true);
      setResourceError(false);
      setSelectedResourceID("");
      try {
        const resolvedOrgSlug = props.orgSlug || (await fetchFirstOrgSlug());
        if (cancelled) return;
        setTargetOrgSlug(resolvedOrgSlug);
        if (!resolvedOrgSlug) {
          setResources([]);
          return;
        }
        updateLightSessionOrgSlug(resolvedOrgSlug);
        const items = await listMarketplaceModelResources(
          resolvedOrgSlug,
          props.agentSlug,
        );
        if (!cancelled) setResources(items);
      } catch {
        if (!cancelled) {
          setResources([]);
          setResourceError(true);
        }
      } finally {
        if (!cancelled) setLoadingResources(false);
      }
    }
    void loadResources();
    return () => {
      cancelled = true;
    };
  }, [props.agentSlug, props.orgSlug, reloadKey]);

  async function install() {
    if (!targetOrgSlug || !selectedResourceID) return;
    setInstalling(true);
    try {
      const result = await installMarketplaceApplication(
        targetOrgSlug,
        props.applicationSlug,
        Number(selectedResourceID),
      );
      props.onInstalled(
        targetOrgSlug,
        result.expert.slug,
        result.already_installed,
      );
    } catch {
      toast.error("启用失败，请稍后重试");
    } finally {
      setInstalling(false);
    }
  }

  if (!loadingResources && !targetOrgSlug) {
    return (
      <Button className="w-full gap-2" onClick={props.onNeedsOrganization}>
        创建组织后启用
        <ArrowRight className="h-4 w-4" />
      </Button>
    );
  }
  if (resourceError) {
    return (
      <Button
        className="w-full gap-2"
        variant="outline"
        onClick={() => setReloadKey((value) => value + 1)}
      >
        <RefreshCw className="h-4 w-4" />
        重新加载模型
      </Button>
    );
  }
  if (!loadingResources && resources.length === 0) {
    return (
      <Button
        className="w-full gap-2"
        variant="outline"
        onClick={() =>
          targetOrgSlug && props.onConfigureResources(targetOrgSlug)
        }
      >
        <Settings2 className="h-4 w-4" />
        配置兼容模型
      </Button>
    );
  }

  const selectedLabel = resources.find(
    (resource) => String(resource.id) === selectedResourceID,
  )?.label;
  return (
    <div className="space-y-2">
      <Select
        value={selectedResourceID}
        onValueChange={setSelectedResourceID}
        disabled={loadingResources || installing}
      >
        <SelectTrigger aria-label="选择运行模型" className="h-10">
          <span className={selectedLabel ? "" : "text-muted-foreground"}>
            {loadingResources ? "正在加载模型" : selectedLabel || "选择运行模型"}
          </span>
        </SelectTrigger>
        <SelectContent>
          {resources.map((resource) => (
            <SelectItem key={resource.id} value={String(resource.id)}>
              {resource.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button
        className="w-full gap-2"
        onClick={install}
        disabled={installing || loadingResources || !selectedResourceID}
      >
        {installing ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
        {installing ? "正在启用" : "立即启用"}
        {!installing ? <ArrowRight className="h-4 w-4" /> : null}
      </Button>
    </div>
  );
}
