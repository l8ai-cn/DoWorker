import { useState, type ReactNode } from "react";
import { Trash2Icon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useIsAdmin } from "@/hooks/useIsAdmin";
import { defaultModelConfig, useModelConfigMutations, useModelConfigs } from "@/hooks/useModelConfigs";

function Section({ title, description, children }: { title: string; description?: string; children: ReactNode }) {
  return (
    <section className="mx-auto max-w-xl space-y-4">
      <div>
        <h1 className="font-semibold text-lg">{title}</h1>
        {description && <p className="mt-1 text-muted-foreground text-sm">{description}</p>}
      </div>
      {children}
    </section>
  );
}

/** Settings → Models: maintain the org/user model pool for Worker launch. */
export function ModelsSettingsSection() {
  const isAdmin = useIsAdmin();
  const { data: models, isLoading } = useModelConfigs();
  const { create, remove } = useModelConfigMutations();
  const [name, setName] = useState("MiniMax");
  const [provider, setProvider] = useState("minimax");
  const [model, setModel] = useState("MiniMax-M3");
  const [baseUrl, setBaseUrl] = useState("https://api.minimax.chat/anthropic");
  const [apiKey, setApiKey] = useState("");
  const [tokenBudget, setTokenBudget] = useState("");
  const [scope, setScope] = useState<"org" | "user">(isAdmin ? "org" : "user");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const onAdd = async () => {
    setError(null);
    if (!name.trim() || !apiKey.trim()) {
      setError("Name and API key are required.");
      return;
    }
    setSaving(true);
    try {
      await create({
        name: name.trim(),
        provider_type: provider.trim(),
        model: model.trim(),
        base_url: baseUrl.trim(),
        credentials: { api_key: apiKey.trim() },
        is_default: !models?.length,
        token_budget: tokenBudget.trim() ? Number(tokenBudget) : null,
        scope,
      });
      setApiKey("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Section
      title="Models"
      description="Configure provider credentials and default models. Workers mount a model at launch so agents start with a working API key."
    >
      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading…</p>
      ) : (
        <ul className="divide-y rounded-lg border border-border">
          {(models ?? []).map((m) => (
            <li key={m.id} className="flex items-center justify-between gap-2 px-3 py-2 text-sm">
              <div>
                <span className="font-medium">{m.name}</span>
                <span className="ml-2 text-muted-foreground">
                  {m.provider_type}/{m.model}
                  {m.is_default ? " · default" : ""}
                  {m.scope === "org" ? " · org" : " · personal"}
                </span>
              </div>
              <Button type="button" variant="ghost" size="icon-sm" aria-label="Delete" onClick={() => remove(m.id)}>
                <Trash2Icon className="size-4" />
              </Button>
            </li>
          ))}
          {!models?.length && (
            <li className="px-3 py-4 text-muted-foreground text-sm">No models yet — add one below.</li>
          )}
        </ul>
      )}

      <div className="space-y-3 rounded-lg border border-border p-4">
        <h2 className="font-medium text-sm">Add model</h2>
        <div className="grid gap-2 sm:grid-cols-2">
          <Input placeholder="Display name" value={name} onChange={(e) => setName(e.target.value)} />
          <Input placeholder="Provider (minimax, anthropic, openai)" value={provider} onChange={(e) => setProvider(e.target.value)} />
          <Input placeholder="Model id" value={model} onChange={(e) => setModel(e.target.value)} />
          <Input placeholder="Base URL" value={baseUrl} onChange={(e) => setBaseUrl(e.target.value)} />
          <Input type="password" placeholder="API key" value={apiKey} onChange={(e) => setApiKey(e.target.value)} className="sm:col-span-2" />
          <Input type="number" placeholder="Default token budget (optional)" value={tokenBudget} onChange={(e) => setTokenBudget(e.target.value)} />
          {isAdmin && (
            <select
              className="rounded-md border border-border bg-background px-3 py-2 text-sm"
              value={scope}
              onChange={(e) => setScope(e.target.value as "org" | "user")}
            >
              <option value="org">Org shared</option>
              <option value="user">Personal</option>
            </select>
          )}
        </div>
        {error && <p className="text-destructive text-sm">{error}</p>}
        <Button type="button" disabled={saving} onClick={() => void onAdd()}>
          {saving ? "Saving…" : "Save model"}
        </Button>
        {defaultModelConfig(models) && (
          <p className="text-muted-foreground text-xs">
            Default for new Workers: {defaultModelConfig(models)?.name}
          </p>
        )}
      </div>
    </Section>
  );
}
