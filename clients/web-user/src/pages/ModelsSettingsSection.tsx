import type { ReactNode } from "react";
import { CircleAlertIcon, CpuIcon } from "lucide-react";
import { defaultModelConfig, useModelConfigs } from "@/hooks/useModelConfigs";

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
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

export function ModelsSettingsSection() {
  const { data: resources, isLoading, isError, error } = useModelConfigs();
  const defaultResource = defaultModelConfig(resources);

  return (
    <Section
      title="AI Resources"
      description="Available organization and personal resources for Worker launch."
    >
      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading…</p>
      ) : isError ? (
        <div className="flex items-start gap-2 border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          <CircleAlertIcon className="mt-0.5 size-4 shrink-0" />
          <span>{error instanceof Error ? error.message : "Failed to load AI resources."}</span>
        </div>
      ) : (
        <ul className="divide-y border border-border">
          {(resources ?? []).map((resource) => (
            <li key={resource.id} className="flex items-center gap-3 px-3 py-3 text-sm">
              <CpuIcon className="size-4 shrink-0 text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <div className="truncate font-medium">{resource.name}</div>
                <div className="truncate text-muted-foreground">
                  {resource.provider_key}/{resource.model}
                  {resource.is_default ? " · default" : ""}
                </div>
              </div>
            </li>
          ))}
          {!resources?.length && (
            <li className="px-3 py-4 text-muted-foreground text-sm">
              No compatible AI resources are configured.
            </li>
          )}
        </ul>
      )}

      {!isLoading && !isError && defaultResource && (
        <p className="text-muted-foreground text-xs">
          Default for new Workers: {defaultResource.name}
        </p>
      )}
    </Section>
  );
}
