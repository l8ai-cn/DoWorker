"use client";

import { Button } from "@/components/ui/button";
import { FileJson, Star, Check, Edit2, Trash2, Plus } from "lucide-react";
import type { ConfigFile } from "@/lib/api";
import type { ConfigFileBundleViewModel } from "./types";

interface Props {
  bundles: ConfigFileBundleViewModel[];
  fileSpecs: ConfigFile[];
  onSetDefault: (id: number) => Promise<void>;
  onClearDefault: () => Promise<void>;
  onEdit: (b: ConfigFileBundleViewModel) => void;
  onDelete: (id: number) => Promise<void>;
  onAdd: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
}

function previewJson(raw?: string): string {
  if (!raw) return "";
  const oneLine = raw.replace(/\s+/g, " ").trim();
  return oneLine.length > 120 ? `${oneLine.slice(0, 117)}…` : oneLine;
}

export function ConfigFilesSection({
  bundles,
  fileSpecs,
  onSetDefault,
  onClearDefault,
  onEdit,
  onDelete,
  onAdd,
  t,
}: Props) {
  if (fileSpecs.length === 0) return null;

  const hasDefault = bundles.some((b) => b.is_default);
  const targetLabel = fileSpecs.map((f) => f.id).join(", ");

  return (
    <div className="surface-card p-6">
      <div className="flex items-center gap-2 mb-4">
        <FileJson className="w-5 h-5 text-muted-foreground" />
        <h3 className="text-lg font-semibold">{t("settings.agentConfig.configFiles.title")}</h3>
      </div>
      <p className="text-sm text-muted-foreground mb-4">
        {t("settings.agentConfig.configFiles.description", { files: targetLabel })}
      </p>

      <div className="space-y-2">
        {bundles.length === 0 && (
          <div className="text-sm text-muted-foreground py-2">
            {t("settings.agentConfig.configFiles.empty")}
          </div>
        )}

        {bundles.map((b) => (
          <div
            key={b.id}
            className="flex items-center justify-between p-3 surface-card motion-interactive hover:bg-surface-muted"
          >
            <div className="flex items-center gap-3 min-w-0 flex-1">
              <FileJson className="w-4 h-4 text-muted-foreground shrink-0" />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium truncate">{b.name}</span>
                  {b.is_default && (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs bg-primary/10 text-primary shrink-0">
                      <Star className="w-3 h-3 mr-0.5" />
                      {t("settings.agentCredentials.default")}
                    </span>
                  )}
                </div>
                <div className="text-xs text-muted-foreground font-mono mt-0.5 truncate">
                  {previewJson(b.json_content) || t("settings.agentConfig.configFiles.noJson")}
                </div>
              </div>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              {!b.is_default && (
                <Button variant="ghost" size="sm" onClick={() => onSetDefault(b.id)}>
                  <Check className="w-4 h-4" />
                </Button>
              )}
              <Button variant="ghost" size="sm" onClick={() => onEdit(b)}>
                <Edit2 className="w-4 h-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onDelete(b.id)}
                className="text-destructive hover:text-destructive"
              >
                <Trash2 className="w-4 h-4" />
              </Button>
            </div>
          </div>
        ))}

        <div className="flex items-center gap-2 mt-2">
          <Button variant="outline" size="sm" onClick={onAdd}>
            <Plus className="w-4 h-4 mr-1" />
            {t("settings.agentConfig.configFiles.add")}
          </Button>
          {hasDefault && (
            <Button variant="ghost" size="sm" onClick={onClearDefault}>
              {t("settings.agentConfig.configFiles.clearDefault")}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
