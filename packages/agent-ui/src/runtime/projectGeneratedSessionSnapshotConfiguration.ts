import type { SessionSnapshot } from "@do-worker/proto/agent_workbench/v2/session_pb";

import type { AgentConfigurationControl } from "../contracts";

export function projectConfiguration(
  snapshot: SessionSnapshot,
): AgentConfigurationControl[] | undefined {
  const configuration = snapshot.configuration;
  const capabilities = snapshot.capabilities;
  if (!configuration || !capabilities) return undefined;
  const controls: AgentConfigurationControl[] = [];
  appendControl(
    controls,
    "model",
    "Model",
    configuration.model,
    capabilities.models,
  );
  appendControl(
    controls,
    "permission_mode",
    "Permission mode",
    configuration.permissionMode,
    capabilities.permissionModes,
  );
  return controls.length > 0 ? controls : undefined;
}

function appendControl(
  controls: AgentConfigurationControl[],
  id: string,
  label: string,
  value: string | undefined,
  options: readonly string[],
): void {
  if (!value || !options.includes(value)) return;
  controls.push({
    id,
    label,
    value,
    options: options.map((option) => ({ value: option, label: option })),
  });
}
