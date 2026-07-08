"use client";

import { useEffect, useState } from "react";
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

  return (
    <div>
      <label htmlFor="worker-vkey-select" className="block text-sm font-medium mb-2">
        {t("ide.createPod.modelBindingLabel")}
      </label>
      <select
        id="worker-vkey-select"
        className="w-full px-3 py-2 border border-border rounded-md bg-background"
        value={selectedVirtualKeyId ?? ""}
        onChange={(e) => onSelect(e.target.value ? Number(e.target.value) : null)}
        disabled={loading}
      >
        <option value="">{t("ide.createPod.modelBindingNone")}</option>
        {keys.map((k) => (
          <option key={k.id} value={k.id}>
            {k.name} ({k.key_prefix}…)
            {k.token_budget ? ` · ${k.token_budget.toLocaleString()} tok` : ""}
          </option>
        ))}
      </select>
      <p className="text-xs text-muted-foreground mt-1">
        {t("ide.createPod.modelBindingHint")}
      </p>
    </div>
  );
}
