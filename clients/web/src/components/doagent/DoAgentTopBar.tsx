"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { Terminal } from "lucide-react";
import { useAcpSessionField } from "@/stores/acpSession";
import { usePodTitle } from "@/hooks/usePodTitle";
import { getShortPodKey } from "@/lib/pod-display-name";

interface DoAgentTopBarProps {
  podKey: string;
}

export function DoAgentTopBar({ podKey }: DoAgentTopBarProps) {
  const t = useTranslations("doagent");
  const params = useParams();
  const org = typeof params.org === "string" ? params.org : "";
  const title = usePodTitle(podKey);
  const sessionState = useAcpSessionField(podKey, (s) => s.state);

  return (
    <div className="flex h-9 items-center justify-between border-b border-border px-3">
      <div className="flex min-w-0 items-center gap-2">
        <span className="text-sm font-medium">{t("title")}</span>
        <span className="truncate text-xs text-muted-foreground">{title}</span>
        <code className="text-[10px] text-muted-foreground">{getShortPodKey(podKey)}</code>
        <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase">{sessionState}</span>
      </div>
      {org && (
        <Link
          href={`/${org}/workspace?pod=${encodeURIComponent(podKey)}`}
          className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          title={t("openWorkspace")}
        >
          <Terminal className="h-3 w-3" />
          {t("openWorkspace")}
        </Link>
      )}
    </div>
  );
}
