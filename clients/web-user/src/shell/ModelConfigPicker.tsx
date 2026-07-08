import { ChevronDownIcon, CpuIcon } from "lucide-react";
import { useEffect, useMemo } from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { defaultModelConfig, useModelConfigs } from "@/hooks/useModelConfigs";
import type { ModelConfig } from "@/lib/modelConfigsApi";
import { cn } from "@/lib/utils";

/**
 * Worker model knobs as a bare menu fragment, for embedding inside the
 * agent picker's Do-agent submenu (no dropdown shell of its own).
 * Shares the same `model_config_id` / `token_budget` state the create
 * request sends, so the flyout pick and the inline picker stay in sync.
 */
export function WorkerModelMenuOptions({
  selectedId,
  onSelect,
  tokenBudget,
  onTokenBudgetChange,
}: {
  selectedId: number | null;
  onSelect: (id: number | null, model: ModelConfig | null) => void;
  tokenBudget: number | null;
  onTokenBudgetChange: (n: number | null) => void;
}) {
  const { data: models } = useModelConfigs();
  if (!models?.length) {
    return (
      <DropdownMenuLabel className="px-2 py-1.5 text-[11px] font-normal text-muted-foreground">
        No models configured — add one in Settings → Models.
      </DropdownMenuLabel>
    );
  }
  const effective = selectedId ?? defaultModelConfig(models)?.id ?? null;
  return (
    <>
      <DropdownMenuLabel className="px-2 pb-0.5 pt-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        Model
      </DropdownMenuLabel>
      <DropdownMenuRadioGroup
        value={effective != null ? String(effective) : ""}
        onValueChange={(v) => {
          const id = Number(v);
          onSelect(id || null, models.find((m) => m.id === id) ?? null);
        }}
      >
        {models.map((m) => (
          <DropdownMenuRadioItem
            key={m.id}
            value={String(m.id)}
            className="rounded-sm py-1 pl-2 text-xs"
            data-testid={`new-chat-landing-worker-model-${m.id}`}
          >
            <span className="truncate">
              {m.name}
              {m.is_default ? " ★" : ""}
            </span>
          </DropdownMenuRadioItem>
        ))}
      </DropdownMenuRadioGroup>
      <DropdownMenuSeparator />
      <div className="px-2 py-1.5" onKeyDown={(e) => e.stopPropagation()}>
        <label className="text-[11px] font-medium text-muted-foreground">
          Token cap (optional)
        </label>
        <Input
          type="number"
          min={0}
          placeholder="Unlimited"
          className="mt-1 h-8 text-xs"
          value={tokenBudget ?? ""}
          onChange={(e) => {
            const v = e.target.value.trim();
            onTokenBudgetChange(v === "" ? null : Number(v));
          }}
          data-testid="new-chat-landing-worker-token-budget"
        />
      </div>
    </>
  );
}

interface ModelConfigPickerProps {
  selectedId: number | null;
  onSelect: (id: number | null, model: ModelConfig | null) => void;
  tokenBudget: number | null;
  onTokenBudgetChange: (n: number | null) => void;
  className?: string;
  disabled?: boolean;
}

/** Compact model + token cap picker for the landing composer (Do-agent Workers). */
export function ModelConfigPicker({
  selectedId,
  onSelect,
  tokenBudget,
  onTokenBudgetChange,
  className,
  disabled = false,
}: ModelConfigPickerProps) {
  const { data: models, isLoading } = useModelConfigs();

  useEffect(() => {
    if (selectedId != null || !models?.length) return;
    const d = defaultModelConfig(models);
    if (d) onSelect(d.id, d);
  }, [models, selectedId, onSelect]);

  const selected = useMemo(
    () => models?.find((m) => m.id === selectedId) ?? defaultModelConfig(models),
    [models, selectedId],
  );

  if (isLoading) {
    return (
      <span className={cn("text-xs text-muted-foreground", className)} data-testid="model-config-loading">
        …
      </span>
    );
  }
  if (!models?.length) {
    return null;
  }

  const label = selected ? selected.name : "Model";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={disabled}
          data-testid="new-chat-landing-model-select"
          className={cn(
            "h-8 max-w-[10.5rem] gap-1 px-2 font-normal text-muted-foreground hover:text-foreground focus-visible:border-transparent focus-visible:ring-0",
            className,
          )}
        >
          <CpuIcon className="size-3.5 shrink-0 text-primary" />
          <span className="truncate text-xs text-foreground">{label}</span>
          <ChevronDownIcon className="size-3.5 shrink-0 opacity-60" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-64 p-1">
        <DropdownMenuRadioGroup
          value={selectedId != null ? String(selectedId) : ""}
          onValueChange={(v) => {
            const id = Number(v);
            if (!id) {
              onSelect(null, null);
              return;
            }
            onSelect(id, models.find((m) => m.id === id) ?? null);
          }}
        >
          {models.map((m) => (
            <DropdownMenuRadioItem
              key={m.id}
              value={String(m.id)}
              className="text-13"
              data-testid={`new-chat-landing-model-${m.id}`}
            >
              <span className="truncate">
                {m.name}
                {m.is_default ? " ★" : ""}
              </span>
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
        <DropdownMenuSeparator />
        <div className="px-2 py-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">
            Token cap (optional)
          </label>
          <Input
            type="number"
            min={0}
            placeholder="Unlimited"
            className="mt-1 h-8 text-xs"
            value={tokenBudget ?? ""}
            onChange={(e) => {
              const v = e.target.value.trim();
              onTokenBudgetChange(v === "" ? null : Number(v));
            }}
            data-testid="new-chat-landing-token-budget"
          />
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
