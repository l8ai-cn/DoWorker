import { useCallback, useState } from "react";
import type { CustomEnvEntry } from "@/components/settings/AgentCredentialsSettings/credentialForms/types";
import { hasInvalidCustomEnvKey } from "@/components/settings/CustomEnvSection";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import { requiresModelResource } from "./useWorkerModelResources";
import type { FormValidationErrors } from "./useCreatePodFormTypes";

export function useCreatePodValidation(params: {
  selectedAgent: string | null;
  selectedRepository: number | null;
  selectedBranch: string;
  customEnv: CustomEnvEntry[];
  bundleLoadError: string | null;
  selectedAgentSlug: string;
  modelResourceError: string | null;
  loadingModelResources: boolean;
  selectedModelResourceId: number | null;
  selectedModelResource?: EffectiveResource;
}) {
  const [validationErrors, setValidationErrors] = useState<FormValidationErrors>({});

  const validate = useCallback((): boolean => {
    const errors: FormValidationErrors = {};
    if (!params.selectedAgent) errors.agent = "Please select an agent";
    if (params.selectedRepository && !params.selectedBranch.trim()) {
      errors.branch = "Branch name is recommended when using a repository";
    }
    if (params.selectedBranch.trim() && !/^[a-zA-Z0-9._/-]+$/.test(params.selectedBranch)) {
      errors.branch = "Branch name contains invalid characters";
    }
    if (hasInvalidCustomEnvKey(params.customEnv, new Set())) {
      errors.env = "One or more environment variable names are invalid";
    }
    if (params.bundleLoadError) errors.runtimeBundles = params.bundleLoadError;
    if (requiresModelResource(params.selectedAgentSlug)) {
      if (params.modelResourceError) {
        errors.modelResource = params.modelResourceError;
      } else if (params.loadingModelResources) {
        errors.modelResource = "Model resources are still loading";
      } else if (!params.selectedModelResourceId) {
        errors.modelResource = "Please select a model resource";
      } else if (!params.selectedModelResource) {
        errors.modelResource = "Selected model resource is no longer available";
      }
    }
    setValidationErrors(errors);
    return Object.values(errors).every((value) => !value);
  }, [params]);

  return { validationErrors, setValidationErrors, validate };
}
