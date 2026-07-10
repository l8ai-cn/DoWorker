"use client";

import { useCallback } from "react";
import { Server, Sliders, Star, X, GripVertical } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import type { EnvBundleSummary } from "@/lib/api";

interface Props {
  bundles: EnvBundleSummary[];
  selectedBundleNames: string[];
  onChange: (names: string[]) => void;
  loading?: boolean;
  error?: string | null;
  validationError?: string;
  t: (key: string) => string;
}

/**
 * Multi-select EnvBundle picker for Pod creation.
 *
 * Renders runtime bundles as an ordered checkbox list. Selection order matters:
 * later bundles override earlier ones on conflicting env keys.
 *
 * Empty selection emits no USE_ENV_BUNDLE directive.
 */
export function EnvBundleMultiSelect({
  bundles,
  selectedBundleNames,
  onChange,
  loading,
  error,
  validationError,
  t,
}: Props) {
  const toggle = useCallback(
    (name: string) => {
      if (selectedBundleNames.includes(name)) {
        onChange(selectedBundleNames.filter((n) => n !== name));
      } else {
        onChange([...selectedBundleNames, name]);
      }
    },
    [selectedBundleNames, onChange]
  );

  const remove = useCallback(
    (name: string) => onChange(selectedBundleNames.filter((n) => n !== name)),
    [selectedBundleNames, onChange]
  );

  const move = useCallback(
    (from: number, to: number) => {
      if (to < 0 || to >= selectedBundleNames.length) return;
      const next = [...selectedBundleNames];
      const [item] = next.splice(from, 1);
      next.splice(to, 0, item);
      onChange(next);
    },
    [selectedBundleNames, onChange]
  );

  if (loading) {
    return (
      <div>
        <label className="block text-sm font-medium mb-2">
          {t("ide.createPod.selectRuntimeBundles")}
        </label>
        <div className="flex items-center text-sm text-muted-foreground py-2">
          <Spinner size="sm" className="mr-2" />
          {t("common.loading")}
        </div>
      </div>
    );
  }

  return (
    <div>
      <label className="block text-sm font-medium mb-2">
        {t("ide.createPod.selectRuntimeBundles")}
      </label>

      {selectedBundleNames.length > 0 && (
        <div className="mb-2 surface-card bg-muted/30 p-2 space-y-1">
          <div className="text-xs text-muted-foreground px-1 pb-1">
            {t("ide.createPod.selectedOrderHint")}
          </div>
          {selectedBundleNames.map((name, idx) => {
            const b = bundles.find((x) => x.name === name);
            return (
              <div
                key={name}
                className="flex items-center gap-2 rounded bg-background px-2 py-1 border border-border"
              >
                <GripVertical className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                <span className="text-xs text-muted-foreground w-4 text-center shrink-0">
                  {idx + 1}
                </span>
                {b ? renderKindIcon(b.kind) : <Sliders className="w-4 h-4 text-muted-foreground shrink-0" />}
                <span className="text-sm flex-1 truncate" title={name}>
                  {name}
                </span>
                {b && (
                  <span className="text-[10px] uppercase tracking-wide text-muted-foreground shrink-0">
                    {kindLabel(b.kind, t)}
                  </span>
                )}
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground disabled:opacity-30 px-1"
                  disabled={idx === 0}
                  onClick={() => move(idx, idx - 1)}
                  title={t("ide.createPod.moveUp")}
                  aria-label={t("ide.createPod.moveUp")}
                >
                  ↑
                </button>
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground disabled:opacity-30 px-1"
                  disabled={idx === selectedBundleNames.length - 1}
                  onClick={() => move(idx, idx + 1)}
                  title={t("ide.createPod.moveDown")}
                  aria-label={t("ide.createPod.moveDown")}
                >
                  ↓
                </button>
                <button
                  type="button"
                  className="text-muted-foreground hover:text-destructive shrink-0"
                  onClick={() => remove(name)}
                  title={t("common.delete")}
                  aria-label={t("common.delete")}
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </div>
            );
          })}
        </div>
      )}

      {bundles.length === 0 ? (
        <p className="text-xs text-muted-foreground py-2">
          {t("ide.createPod.noRuntimeBundlesAvailableHint")}
        </p>
      ) : (
        <div className="surface-card max-h-48 overflow-y-auto">
          {bundles.map((b) => {
            const checked = selectedBundleNames.includes(b.name);
            return (
              <label
                key={b.id}
                className="flex items-center gap-2 px-2 py-1.5 border-b border-border last:border-b-0 motion-interactive hover:bg-surface-muted cursor-pointer"
              >
                <input
                  type="checkbox"
                  className="h-3.5 w-3.5"
                  checked={checked}
                  onChange={() => toggle(b.name)}
                />
                {renderKindIcon(b.kind)}
                <span className="text-sm flex-1 truncate" title={b.name}>
                  {b.name}
                </span>
                {b.kind_primary && (
                  <span className="inline-flex items-center text-[10px] text-primary shrink-0">
                    <Star className="w-2.5 h-2.5 mr-0.5" />
                    {t("settings.agentCredentials.default")}
                  </span>
                )}
                <span className="text-[10px] uppercase tracking-wide text-muted-foreground shrink-0">
                  {kindLabel(b.kind, t)}
                </span>
              </label>
            );
          })}
        </div>
      )}

      <p
        role={validationError || error ? "alert" : undefined}
        className={validationError || error ? "text-xs text-destructive mt-1" : "text-xs text-muted-foreground mt-1"}
      >
        {validationError || error || (selectedBundleNames.length === 0
          ? t("ide.createPod.noRuntimeBundleSelectedHint")
          : t("ide.createPod.multiBundleHint"))}
      </p>
    </div>
  );
}

function renderKindIcon(kind: string) {
  if (kind === "runtime") {
    return <Sliders className="w-4 h-4 text-muted-foreground shrink-0" />;
  }
  return <Server className="w-4 h-4 text-muted-foreground shrink-0" />;
}

function kindLabel(kind: string, t: (key: string) => string): string {
  if (kind === "runtime") return t("ide.createPod.bundleKind.runtime");
  return kind;
}

export default EnvBundleMultiSelect;
