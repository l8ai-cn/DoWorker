import { render } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { vi } from "vitest";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useAvailableAgents, type AvailableAgent } from "@/hooks/useAvailableAgents";
import { useDirectorySessions } from "@/hooks/useDirectorySessions";
import { useHostFilesystem } from "@/hooks/useHostFilesystem";
import { useHosts, type Host } from "@/hooks/useHosts";
import { defaultModelConfig, useModelConfigs } from "@/hooks/useModelConfigs";
import type { ModelConfig } from "@/lib/modelConfigsApi";
import { useRunnerHealthRegistration } from "@/hooks/RunnerHealthProvider";
import { CapabilitiesProvider } from "@/lib/CapabilitiesContext";
import type { ServerInfo } from "@/lib/capabilities";
import { authenticatedFetch } from "@/lib/identity";
import { setDoWorkerHostConfig } from "@/lib/host";
import { setPendingInitialPrompt } from "@/store/chatStore";
import { NewChatLandingScreen, resetLandingDraft } from "./NewChatDialog";

export const authenticatedFetchMock = vi.mocked(authenticatedFetch);
export const setPendingInitialPromptMock = vi.mocked(setPendingInitialPrompt);
const useHostsMock = vi.mocked(useHosts);
const useAvailableAgentsMock = vi.mocked(useAvailableAgents);
const useModelConfigsMock = vi.mocked(useModelConfigs);
const defaultModelConfigMock = vi.mocked(defaultModelConfig);
export const useHostFilesystemMock = vi.mocked(useHostFilesystem);
export const useDirectorySessionsMock = vi.mocked(useDirectorySessions);
export const useRunnerHealthMock = vi.mocked(useRunnerHealthRegistration);
export const RECENT_KEY = "do-worker:recent-workspaces";

export function host(status: "online" | "offline" = "online", i = 1): Host {
  return { host_id: `host_${i}`, name: `machine-${i}`, owner: "me", status };
}

export function mockHosts(hosts: Host[]) {
  useHostsMock.mockReturnValue({ data: hosts } as ReturnType<typeof useHosts>);
}

export function mockAgents(agents: AvailableAgent[]) {
  useAvailableAgentsMock.mockReturnValue({
    data: agents.map((agent) => ({
      workerTypeSlug: agent.workerTypeSlug ?? agent.id,
      supportedModes: agent.supportedModes ?? ["acp", "pty"],
      requiresModelResource: agent.requiresModelResource ?? false,
      ...agent,
    })),
  } as ReturnType<typeof useAvailableAgents>);
}

export function mockModelConfigs(models: ModelConfig[]) {
  useModelConfigsMock.mockReturnValue({
    data: models,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useModelConfigs>);
  defaultModelConfigMock.mockImplementation((rows) => rows?.find((row) => row.is_default) ?? null);
}

export function mockModelConfigError(message: string) {
  useModelConfigsMock.mockReturnValue({
    data: undefined,
    isLoading: false,
    isError: true,
    error: new Error(message),
  } as ReturnType<typeof useModelConfigs>);
  defaultModelConfigMock.mockReturnValue(null);
}

export function setupLandingMocks() {
  authenticatedFetchMock.mockReset();
  setPendingInitialPromptMock.mockReset();
  useHostsMock.mockReset();
  useAvailableAgentsMock.mockReset();
  useModelConfigsMock.mockReset();
  defaultModelConfigMock.mockReset();
  useHostFilesystemMock.mockReset();
  useDirectorySessionsMock.mockReset();
  useRunnerHealthMock.mockReset();
  resetLandingDraft();
  localStorage.clear();
  setDoWorkerHostConfig({});
  localStorage.setItem(RECENT_KEY, JSON.stringify({ host_1: ["/Users/corey/repo"] }));
  mockHosts([host()]);
  mockAgents([
    { id: "a1", name: "claude-native-ui", display_name: "Claude Code", description: null, harness: "claude-native", skills: [] },
    { id: "a2", name: "codex-native-ui", display_name: "Codex", description: null, harness: "codex-native", skills: [] },
  ]);
  mockModelConfigs([]);
  useDirectorySessionsMock.mockReturnValue({ data: [] } as ReturnType<typeof useDirectorySessions>);
  useRunnerHealthMock.mockReturnValue(new Map<string, boolean>());
  useHostFilesystemMock.mockReturnValue({
    data: undefined,
    isLoading: false,
    error: null,
    isPlaceholderData: false,
  } as ReturnType<typeof useHostFilesystem>);
}

export function renderLanding(infoOverrides: Partial<ServerInfo> = {}, route = "/") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const info: ServerInfo = {
    accounts_enabled: false, login_url: null, needs_setup: false, databricks_features: false,
    managed_sandboxes_enabled: false, sandbox_provider: null, server_version: null,
    smart_routing_enabled: false, ...infoOverrides,
  };
  return render(
    <QueryClientProvider client={client}>
      <CapabilitiesProvider info={info}>
        <TooltipProvider><MemoryRouter initialEntries={[route]}><NewChatLandingScreen /></MemoryRouter></TooltipProvider>
      </CapabilitiesProvider>
    </QueryClientProvider>,
  );
}
