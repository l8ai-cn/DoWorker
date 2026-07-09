"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  type TokenQuota,
  deleteTokenQuota,
  listTokenQuotas,
  upsertTokenQuota,
} from "@/lib/api/quotaApi";

const inputCls =
  "w-full rounded-md border border-border bg-background px-3 py-1.5 text-sm";

export function TokenQuotaPanel() {
  const [quotas, setQuotas] = useState<TokenQuota[]>([]);
  const [userId, setUserId] = useState("");
  const [model, setModel] = useState("");
  const [limit, setLimit] = useState("");

  const refresh = useCallback(async () => {
    try {
      setQuotas(await listTokenQuotas());
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to load quotas");
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void listTokenQuotas(controller.signal)
      .then(setQuotas)
      .catch((e) => {
        if (!controller.signal.aborted) {
          toast.error(e instanceof Error ? e.message : "Failed to load quotas");
        }
      });
    return () => controller.abort();
  }, []);

  const onSave = async () => {
    if (!limit) {
      toast.error("Limit is required");
      return;
    }
    try {
      await upsertTokenQuota({
        user_id: userId ? Number(userId) : null,
        model: model.trim() || null,
        limit_tokens: Number(limit),
      });
      setUserId("");
      setModel("");
      setLimit("");
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to save quota");
    }
  };

  return (
    <div className="surface-card space-y-4 p-6">
      <div>
        <h2 className="text-lg font-semibold">Token Quotas</h2>
        <p className="text-sm text-muted-foreground">
          Set token ceilings by organization (leave user blank), per user, and/or per model. Report-only — over-limit is flagged, not blocked.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-2 sm:grid-cols-4">
        <input
          className={inputCls}
          placeholder="User ID (blank = org-wide)"
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
        />
        <input
          className={inputCls}
          placeholder="Model (blank = all)"
          value={model}
          onChange={(e) => setModel(e.target.value)}
        />
        <input
          className={inputCls}
          type="number"
          placeholder="Limit tokens"
          value={limit}
          onChange={(e) => setLimit(e.target.value)}
        />
        <Button onClick={onSave}>Save quota</Button>
      </div>

      <div className="divide-y divide-border">
        {quotas.length === 0 && (
          <p className="py-3 text-sm text-muted-foreground">No quotas configured.</p>
        )}
        {quotas.map((q) => (
          <div key={q.id} className="flex items-center justify-between py-2 text-sm">
            <div>
              <span className="font-medium">
                {q.user_id ? `User ${q.user_id}` : "Organization"}
              </span>
              <span className="ml-2 text-muted-foreground">
                {q.model ? `· ${q.model}` : "· all models"} · {q.limit_tokens.toLocaleString()} tokens
              </span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={async () => {
                try {
                  await deleteTokenQuota(q.id);
                  await refresh();
                } catch (e) {
                  toast.error(e instanceof Error ? e.message : "Failed to delete quota");
                }
              }}
            >
              Delete
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
