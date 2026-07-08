"use client";

import { useEffect, useState } from "react";
import { RefreshCw, Save } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  getKnowledgeBase,
  syncKnowledgeBase,
  updateKnowledgeBase,
  type KnowledgeBase,
} from "@/lib/api/facade/knowledgeBaseApi";
import { SourceConfigFields } from "./SourceConfigFields";
import {
  SOURCE_LABELS,
  SYNC_STATUS_LABELS,
  buildSourceConfigJson,
  emptySourceConfig,
  isExternalSource,
  parseSourceConfigJson,
  syncStatusVariant,
  type SourceConfigForm,
} from "./sourceConfig";

interface KnowledgeBaseSourceSettingsProps {
  orgSlug: string;
  kb: KnowledgeBase;
  onUpdated: (kb: KnowledgeBase) => void;
}

export function KnowledgeBaseSourceSettings({
  orgSlug,
  kb,
  onUpdated,
}: KnowledgeBaseSourceSettingsProps) {
  const external = isExternalSource(kb.source_type);
  const [sourceConfig, setSourceConfig] = useState<SourceConfigForm>({});
  const [saving, setSaving] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!external) return;
    const parsed = parseSourceConfigJson(kb.source_config_json);
    setSourceConfig({ ...emptySourceConfig(kb.source_type), ...parsed });
  }, [external, kb.source_config_json, kb.source_type]);

  if (!external) return null;

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const existing = parseSourceConfigJson(kb.source_config_json);
      const updated = await updateKnowledgeBase(orgSlug, kb.slug, {
        sourceConfigJson: buildSourceConfigJson(kb.source_type, sourceConfig, existing),
      });
      setMessage("数据源配置已保存");
      onUpdated(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存配置失败");
    } finally {
      setSaving(false);
    }
  };

  const handleSync = async () => {
    setSyncing(true);
    setError(null);
    setMessage(null);
    try {
      const updated = await syncKnowledgeBase(orgSlug, kb.slug);
      setMessage("同步完成");
      onUpdated(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "同步失败");
      try {
        onUpdated(await getKnowledgeBase(orgSlug, kb.slug));
      } catch {
        // ignore refresh failure
      }
    } finally {
      setSyncing(false);
    }
  };

  return (
    <div className="border-b border-border px-6 py-4">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm font-medium">外部同步</span>
          <Badge variant="secondary">{SOURCE_LABELS[kb.source_type] ?? kb.source_type}</Badge>
          {kb.sync_status && (
            <Badge variant={syncStatusVariant(kb.sync_status)}>
              {SYNC_STATUS_LABELS[kb.sync_status] ?? kb.sync_status}
            </Badge>
          )}
          {kb.last_synced_at && (
            <span className="text-xs text-muted-foreground">上次同步：{kb.last_synced_at}</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleSync} disabled={syncing || saving}>
            <RefreshCw className={`mr-1 h-3.5 w-3.5 ${syncing ? "animate-spin" : ""}`} />
            立即同步
          </Button>
          <Button size="sm" onClick={handleSave} loading={saving} disabled={syncing}>
            <Save className="mr-1 h-3.5 w-3.5" />
            保存配置
          </Button>
        </div>
      </div>
      {kb.sync_error && <p className="mb-3 text-sm text-destructive">{kb.sync_error}</p>}
      {error && <p className="mb-3 text-sm text-destructive">{error}</p>}
      {message && <p className="mb-3 text-sm text-muted-foreground">{message}</p>}
      <SourceConfigFields
        sourceType={kb.source_type}
        value={sourceConfig}
        onChange={setSourceConfig}
        idPrefix={`kb-settings-${kb.slug}`}
      />
    </div>
  );
}
