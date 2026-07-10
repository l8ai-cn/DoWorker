import { useQuery } from "@tanstack/react-query";
import { listModelResources, type ModelConfig } from "@/lib/modelConfigsApi";

export function useModelConfigs(enabled = true) {
  return useQuery({
    queryKey: ["model-resources"],
    queryFn: listModelResources,
    enabled,
    staleTime: 30_000,
  });
}

export function defaultModelConfig(models: ModelConfig[] | undefined): ModelConfig | null {
  if (!models?.length) return null;
  return models.find((m) => m.is_default) ?? null;
}
