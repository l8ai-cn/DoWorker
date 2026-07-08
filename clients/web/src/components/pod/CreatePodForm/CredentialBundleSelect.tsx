"use client";

import { Key, Star } from "lucide-react";
import { cn } from "@/lib/utils";
import { Spinner } from "@/components/ui/spinner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import type { EnvBundleSummary } from "@/lib/api";

interface Props {
  bundles: EnvBundleSummary[];
  selectedBundleName: string;
  onSelect: (name: string) => void;
  loading?: boolean;
  t: (key: string) => string;
}

/**
 * Single-select picker for credential-kind EnvBundles.
 *
 * Always offers "use Agent default authentication" as the first option
 * (empty string value) so users can deliberately opt out of bundle
 * injection — same semantics as the per-agent Settings page.
 *
 * Caller must filter `bundles` to `kind === 'credential'`.
 */
export function CredentialBundleSelect({
  bundles,
  selectedBundleName,
  onSelect,
  loading,
  t,
}: Props) {
  if (loading) {
    return (
      <div>
        <label className="block text-sm font-medium mb-2">
          {t("ide.createPod.selectCredential")}
        </label>
        <div className="flex items-center text-sm text-muted-foreground py-2">
          <Spinner size="sm" className="mr-2" />
          {t("common.loading")}
        </div>
      </div>
    );
  }

  const selectedBundle = bundles.find((b) => b.name === selectedBundleName);
  const triggerLabel = selectedBundle
    ? `${selectedBundle.name}${selectedBundle.kind_primary ? ` (${t("settings.agentCredentials.default")})` : ""}`
    : t("ide.createPod.useAgentDefaultAuth");

  return (
    <div>
      <label
        htmlFor="credential-bundle-select"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.selectCredential")}
      </label>
      <Select value={selectedBundleName} onValueChange={onSelect}>
        <SelectTrigger id="credential-bundle-select">
          <span className={cn(!selectedBundleName && "text-muted-foreground")}>
            {triggerLabel}
          </span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("ide.createPod.useAgentDefaultAuth")}</SelectItem>
          {bundles.map((b) => (
            <SelectItem key={b.id} value={b.name}>
              {b.name}
              {b.kind_primary ? ` (${t("settings.agentCredentials.default")})` : ""}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
        {selectedBundleName ? (
          <>
            <Key className="w-3 h-3 shrink-0" />
            {t("ide.createPod.credentialSelectedHint")}
          </>
        ) : (
          <>
            {bundles.some((b) => b.kind_primary) && (
              <Star className="w-3 h-3 shrink-0 text-primary" />
            )}
            {t("ide.createPod.noCredentialHint")}
          </>
        )}
      </p>
    </div>
  );
}

export default CredentialBundleSelect;
