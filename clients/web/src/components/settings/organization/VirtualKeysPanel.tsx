"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  type ModelConfig,
  type VirtualKey,
  createVirtualKey,
  listVirtualKeys,
  revokeVirtualKey,
} from "@/lib/api/quotaApi";

const inputCls =
  "w-full rounded-md border border-border bg-background px-3 py-1.5 text-sm";

export function VirtualKeysPanel({ models }: { models: ModelConfig[] }) {
  const [keys, setKeys] = useState<VirtualKey[]>([]);
  const [name, setName] = useState("");
  const [modelId, setModelId] = useState<number | "">("");
  const [budget, setBudget] = useState<string>("");
  const [newToken, setNewToken] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      setKeys(await listVirtualKeys());
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to load virtual keys");
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void listVirtualKeys(controller.signal)
      .then(setKeys)
      .catch((e) => {
        if (!controller.signal.aborted) {
          toast.error(e instanceof Error ? e.message : "Failed to load virtual keys");
        }
      });
    return () => controller.abort();
  }, []);

  const onCreate = async () => {
    if (!name.trim() || modelId === "") {
      toast.error("Name and model are required");
      return;
    }
    try {
      const res = await createVirtualKey({
        name: name.trim(),
        ai_model_id: Number(modelId),
        token_budget: budget ? Number(budget) : undefined,
      });
      setNewToken(res.token);
      setName("");
      setBudget("");
      setModelId("");
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to create key");
    }
  };

  const onRevoke = async (id: number) => {
    try {
      await revokeVirtualKey(id);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to revoke key");
    }
  };

  return (
    <div className="surface-card space-y-4 p-6">
      <div>
        <h2 className="text-lg font-semibold">Virtual API Keys</h2>
        <p className="text-sm text-muted-foreground">
          Issue a token-budgeted handle over one of your model credentials. Bind it to a Worker to attribute usage.
        </p>
      </div>

      {newToken && (
        <div className="rounded-md border border-primary/40 bg-primary/5 p-3 text-sm">
          <p className="mb-1 font-medium">Copy this token now — it is shown only once:</p>
          <code className="break-all">{newToken}</code>
        </div>
      )}

      <div className="grid grid-cols-1 gap-2 sm:grid-cols-4">
        <input
          className={inputCls}
          placeholder="Key name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <select
          className={inputCls}
          value={modelId}
          onChange={(e) => setModelId(e.target.value ? Number(e.target.value) : "")}
        >
          <option value="">Select model…</option>
          {models.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name} ({m.model})
            </option>
          ))}
        </select>
        <input
          className={inputCls}
          type="number"
          placeholder="Token budget (optional)"
          value={budget}
          onChange={(e) => setBudget(e.target.value)}
        />
        <Button onClick={onCreate}>Create key</Button>
      </div>

      <div className="divide-y divide-border">
        {keys.length === 0 && (
          <p className="py-3 text-sm text-muted-foreground">No virtual keys yet.</p>
        )}
        {keys.map((k) => (
          <div key={k.id} className="flex items-center justify-between py-2 text-sm">
            <div>
              <span className="font-medium">{k.name}</span>{" "}
              <code className="text-muted-foreground">{k.key_prefix}…</code>
              <span className="ml-2 text-xs text-muted-foreground">
                {k.token_budget ? `${k.token_budget.toLocaleString()} tokens` : "unlimited"} · {k.status}
              </span>
            </div>
            {k.status === "active" && (
              <Button variant="outline" size="sm" onClick={() => onRevoke(k.id)}>
                Revoke
              </Button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
