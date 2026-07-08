import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createModelConfig,
  deleteModelConfig,
  listModelConfigs,
  type CreateModelConfigInput,
  type ModelConfig,
} from "@/lib/modelConfigsApi";

export function useModelConfigs(enabled = true) {
  return useQuery({
    queryKey: ["model-configs"],
    queryFn: listModelConfigs,
    enabled,
    staleTime: 30_000,
  });
}

export function useModelConfigMutations() {
  const queryClient = useQueryClient();
  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["model-configs"] });

  return {
    create: async (input: CreateModelConfigInput) => {
      const m = await createModelConfig(input);
      await invalidate();
      return m;
    },
    remove: async (id: number) => {
      await deleteModelConfig(id);
      await invalidate();
    },
  };
}

export function defaultModelConfig(models: ModelConfig[] | undefined): ModelConfig | null {
  if (!models?.length) return null;
  return models.find((m) => m.is_default) ?? models[0] ?? null;
}
