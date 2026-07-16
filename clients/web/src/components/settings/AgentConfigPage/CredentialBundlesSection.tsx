"use client";

import { Check, Edit2, KeyRound, Plus, Star, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { CredentialProfileViewModel } from "../_shared/credentialViewModel";
import { getConfiguredKeys } from "../_shared/credentialViewModel";

interface CredentialBundlesSectionProps {
  bundles: CredentialProfileViewModel[];
  onSetDefault: (id: number) => Promise<void>;
  onClearDefault: () => Promise<void>;
  onEdit: (bundle: CredentialProfileViewModel) => void;
  onDelete: (id: number) => Promise<void>;
  onAdd: () => void;
  t: (key: string) => string;
}

export function CredentialBundlesSection({
  bundles,
  onSetDefault,
  onClearDefault,
  onEdit,
  onDelete,
  onAdd,
  t,
}: CredentialBundlesSectionProps) {
  const hasDefault = bundles.some((bundle) => bundle.is_default);
  return (
    <div className="surface-card p-6">
      <div className="mb-4 flex items-center gap-2">
        <KeyRound className="h-5 w-5 text-muted-foreground" />
        <h3 className="text-lg font-semibold">
          {t("settings.agentConfig.credentialBundles.title")}
        </h3>
      </div>
      <p className="mb-4 text-sm text-muted-foreground">
        {t("settings.agentConfig.credentialBundles.description")}
      </p>
      <div className="space-y-2">
        {bundles.length === 0 && (
          <p className="py-2 text-sm text-muted-foreground">
            {t("settings.agentConfig.credentialBundles.empty")}
          </p>
        )}
        {bundles.map((bundle) => (
          <div
            key={bundle.id}
            className="surface-card motion-interactive flex items-center justify-between p-3 hover:bg-surface-muted"
          >
            <div className="flex min-w-0 flex-1 items-center gap-3">
              <KeyRound className="h-4 w-4 shrink-0 text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="truncate font-medium">{bundle.name}</span>
                  {bundle.is_default && (
                    <span className="inline-flex shrink-0 items-center rounded bg-primary/10 px-1.5 py-0.5 text-xs text-primary">
                      <Star className="mr-0.5 h-3 w-3" />
                      {t("settings.agentCredentials.default")}
                    </span>
                  )}
                </div>
                {bundle.description && (
                  <p className="truncate text-xs text-muted-foreground">{bundle.description}</p>
                )}
                <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
                  {getConfiguredKeys(bundle).join(", ") || t("settings.agentConfig.credentialBundles.noKeys")}
                </p>
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-1">
              {!bundle.is_default && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onSetDefault(bundle.id)}
                  title={t("settings.agentCredentials.setAsDefault")}
                >
                  <Check className="h-4 w-4" />
                </Button>
              )}
              <Button variant="ghost" size="sm" onClick={() => onEdit(bundle)}>
                <Edit2 className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onDelete(bundle.id)}
                className="text-destructive hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        ))}
        <div className="mt-2 flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onAdd}>
            <Plus className="mr-1 h-4 w-4" />
            {t("settings.agentConfig.credentialBundles.add")}
          </Button>
          {hasDefault && (
            <Button variant="ghost" size="sm" onClick={onClearDefault}>
              {t("settings.agentConfig.credentialBundles.clearDefault")}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
