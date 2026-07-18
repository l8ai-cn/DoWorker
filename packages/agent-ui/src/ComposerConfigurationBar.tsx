import { useState } from "react";

import type {
  AgentSessionRuntime,
  AgentSessionSnapshot,
} from "./contracts";
import { ConfigurationSelect } from "./ConfigurationSelect";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";

export function ComposerConfigurationBar({
  onError,
  runtime,
  snapshot,
}: {
  onError: (error: unknown) => void;
  runtime: AgentSessionRuntime;
  snapshot: AgentSessionSnapshot;
}) {
  const [pending, setPending] = useState<string | null>(null);
  const text = useAgentWorkspaceText();
  if (!snapshot.capabilities.updateConfiguration) return null;

  return (snapshot.configuration ?? []).map((control) => {
    const label = text.configurationLabel(control.id, control.label);
    const options = control.options.map((option) => ({
      ...option,
      label: text.configurationOption(control.id, option.value, option.label),
    }));
    return (
      <ConfigurationSelect
        disabled={pending === control.id}
        key={control.id}
        label={label}
        onChange={(value) => {
          setPending(control.id);
          void runtime
            .updateConfiguration(
              snapshot.sessionId,
              crypto.randomUUID(),
              { [control.id]: value },
            )
            .catch(onError)
            .finally(() => setPending(null));
        }}
        options={options}
        optionsLabel={text.configurationOptions(label)}
        placeholder={label}
        value={control.value}
      />
    );
  });
}
