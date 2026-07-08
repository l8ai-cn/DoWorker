"use client";

import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { listVirtualKeys, type VirtualKey } from "@/lib/api/quotaApi";

interface WorkerModelBindingSelectProps {
  selectedVirtualKeyId: number | null;
  onSelect: (id: number | null) => void;
  t: (key: string) => string;
}

// Binds a Worker to a platform-issued virtual API key so its token usage is
// attributed to that key's quota. "None" falls back to the agent's own
// credentials with no per-Worker attribution. Direct model-config binding is
// an admin concept and intentionally omitted from the create form.
export function WorkerModelBindingSelect({
  selectedVirtualKeyId,
  onSelect,
  t,
}: WorkerModelBindingSelectProps) {
  const [keys, setKeys] = useState<VirtualKey[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listVirtualKeys()
      .then((all) => {
        if (!cancelled) setKeys(all.filter((k) => k.status === "active"));
      })
      .catch(() => {
        if (!cancelled) setKeys([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!loading && keys.length === 0) return null;

  const selectedKey = keys.find((k) => k.id === selectedVirtualKeyId);
  const formatKeyLabel = (k: VirtualKey) =>
    `${k.name} (${k.key_prefix}…)${k.token_budget ? ` · ${k.token_budget.toLocaleString()} tok` : ""}`;

  return (
    <div>
      <label htmlFor="worker-vkey-select" className="block text-sm font-medium mb-2">
        {t("ide.createPod.modelBindingLabel")}
      </label>
      <Select
        value={selectedVirtualKeyId ? String(selectedVirtualKeyId) : ""}
        onValueChange={(value) => onSelect(value ? Number(value) : null)}
        disabled={loading}
      >
        <SelectTrigger id="worker-vkey-select">
          <span className={cn(!selectedVirtualKeyId && "text-muted-foreground")}>
            {loading
              ? t("common.loading")
              : selectedKey
                ? formatKeyLabel(selectedKey)
                : t("ide.createPod.modelBindingNone")}
          </span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("ide.createPod.modelBindingNone")}</SelectItem>
          {keys.map((k) => (
            <SelectItem key={k.id} value={String(k.id)}>
              {formatKeyLabel(k)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-xs text-muted-foreground mt-1">
        {t("ide.createPod.modelBindingHint")}
      </p>
    </div>
  );
}
