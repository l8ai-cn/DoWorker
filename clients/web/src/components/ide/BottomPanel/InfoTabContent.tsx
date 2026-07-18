"use client";

import React from "react";

import { cn } from "@/lib/utils";
import type { PodData } from "@/lib/api/facade/pod";
import { getPodDisplayName } from "@/lib/pod-display-name";
import { getPodStatusInfo } from "@/stores/mesh";
import { usePods } from "@/stores/pod";
import { PodInfoGrid } from "./PodInfoGrid";
import { RelatedPodsList } from "./RelatedPodsList";
import { Terminal } from "lucide-react";

function getRelatedPods(pods: PodData[], pod: PodData | null): PodData[] {
  if (!pod?.ticket?.id) return [];
  return pods.filter(
    (p) => p.ticket?.id === pod.ticket?.id && p.pod_key !== pod.pod_key
  );
}

interface InfoTabContentProps {
  selectedPodKey: string | null;
  pod: PodData | null;
  orgSlug: string;
  t: (key: string, params?: Record<string, string | number>) => string;
}

export function InfoTabContent({
  selectedPodKey,
  pod,
  orgSlug,
  t,
}: InfoTabContentProps) {
  const pods = usePods();
  const relatedPods = getRelatedPods(pods, pod);

  if (!selectedPodKey) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
        <Terminal className="w-8 h-8 mb-2 opacity-50" />
        <span className="text-xs">{t("ide.bottomPanel.selectPodFirst")}</span>
      </div>
    );
  }

  if (!pod) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
        <Terminal className="w-8 h-8 mb-2 opacity-50" />
        <span className="text-xs">{t("ide.bottomPanel.infoTab.notFound")}</span>
      </div>
    );
  }

  const statusInfo = getPodStatusInfo(pod.status);

  return (
    <div className="h-full overflow-auto space-y-3">
      {/* Pod Name & Status */}
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium truncate">
          {getPodDisplayName(pod, 40)}
        </span>
        <span
          className={cn(
            "px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap",
            statusInfo.color,
            statusInfo.bgColor
          )}
        >
          {statusInfo.label}
        </span>
      </div>

      <PodInfoGrid pod={pod} orgSlug={orgSlug} t={t} />

      {/* Related Pods */}
      {relatedPods.length > 0 && (
        <RelatedPodsList relatedPods={relatedPods} t={t} />
      )}
    </div>
  );
}

export default InfoTabContent;
