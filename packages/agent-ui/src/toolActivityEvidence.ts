import type { AgentToolActivityItem } from "./agentToolContracts";

export function toolActivityRawEvidence(item: AgentToolActivityItem) {
  return JSON.stringify(
    {
      identity: item.identity,
      inputValue: item.inputValue,
      results: item.results,
      title: item.title,
    },
    (_key, value) => (typeof value === "bigint" ? value.toString() : value),
    2,
  );
}

export function cleanToolEvidence(value?: string) {
  const clean = value?.trim();
  return !clean || clean === "{}" || clean === "null" ? undefined : clean;
}
