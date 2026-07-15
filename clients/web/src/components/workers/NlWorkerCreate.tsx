"use client";

import { Sparkles, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { AlertMessage } from "@/components/ui/alert-message";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { AsyncState } from "@/components/pod/hooks/workerCreateDraft";
import { modelResourceLabel } from "@/components/pod/CreatePodForm/workerModelResources";

interface NlWorkerCreateProps {
  prompt: string;
  filling: boolean;
  generationModelResourceId: number;
  generationModels: AsyncState<EffectiveResource[]>;
  onPromptChange: (prompt: string) => void;
  onGenerationModelChange: (resourceId: number) => void;
  onFill: (prompt: string) => void;
}

export function NlWorkerCreate({
  prompt,
  filling,
  generationModelResourceId,
  generationModels,
  onPromptChange,
  onGenerationModelChange,
  onFill,
}: NlWorkerCreateProps) {
  const t = useTranslations();
  const trimmed = prompt.trim();
  const selectedGenerationModel = generationModels.status === "ready"
    ? generationModels.data.find(
      (item) => item.resource?.id === generationModelResourceId,
    )
    : undefined;
  const canFill = Boolean(trimmed && !filling && selectedGenerationModel);

  return (
    <section
      className="mb-6 rounded-lg border border-border bg-surface-raised p-4"
      data-testid="worker-fill-panel"
    >
      <div className="mb-3 flex items-center gap-2">
        <Sparkles className="h-4 w-4 text-primary" />
        <h2 className="text-sm font-medium">{t("workers.create.nl.title")}</h2>
      </div>
      <GenerationModelField
        state={generationModels}
        selected={selectedGenerationModel}
        onChange={onGenerationModelChange}
        t={t}
      />
      <Textarea
        value={prompt}
        onChange={(event) => onPromptChange(event.target.value)}
        placeholder={t("workers.create.nl.placeholder")}
        rows={3}
        maxLength={10000}
        disabled={filling}
        onKeyDown={(e) => {
          if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            e.preventDefault();
            if (canFill) onFill(trimmed);
          }
        }}
      />
      <div className="mt-3 flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {t("workers.create.nl.hint")}
        </p>
        <Button type="button" onClick={() => onFill(trimmed)} disabled={!canFill}>
          {filling && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {t("workers.create.nl.submit")}
        </Button>
      </div>
    </section>
  );
}

function GenerationModelField(props: {
  state: AsyncState<EffectiveResource[]>;
  selected?: EffectiveResource;
  onChange: (resourceId: number) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  const { state, selected, onChange, t } = props;
  return (
    <div className="mb-3">
      <label htmlFor="worker-generation-model" className="mb-2 block text-sm font-medium">
        {t("workers.create.nl.generationModel")}
      </label>
      {state.status === "idle" || state.status === "loading" ? (
        <div className="flex h-9 items-center gap-2 text-sm text-muted-foreground">
          <Spinner size="sm" />
          {t("workers.create.nl.loadingGenerationModels")}
        </div>
      ) : state.status === "error" ? (
        <AlertMessage type="error" message={state.error} />
      ) : state.data.length === 0 ? (
        <AlertMessage
          type="warning"
          message={t("workers.create.nl.noGenerationModels")}
        />
      ) : (
        <Select
          value={selected?.resource?.id ? String(selected.resource.id) : ""}
          onValueChange={(value) => onChange(Number(value))}
          disabled={state.data.length === 0}
        >
          <SelectTrigger
            id="worker-generation-model"
            aria-label={t("workers.create.nl.generationModel")}
          >
            <span className={selected ? undefined : "text-muted-foreground"}>
              {selected
                ? modelResourceLabel(selected)
                : t("workers.create.nl.generationModelPlaceholder")}
            </span>
          </SelectTrigger>
          <SelectContent>
            {state.data.map((resource) => (
              <SelectItem
                key={resource.resource!.id}
                value={String(resource.resource!.id)}
              >
                {modelResourceLabel(resource)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}
