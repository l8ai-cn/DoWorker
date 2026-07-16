import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { WorkflowCreateDialog } from "../WorkflowCreateDialog";
import type { EffectiveResource, ProviderDefinition } from "@/lib/api/facade/aiResource";
import type { WorkflowData } from "@/lib/viewModels/workflow";

const mockCreateWorkflow = vi.fn();
const mockUpdateWorkflow = vi.fn();
vi.mock("@/stores/workflow", () => ({
  useWorkflowStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ createWorkflow: mockCreateWorkflow, updateWorkflow: mockUpdateWorkflow }),
}));
vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "acme" }),
}));
vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme" }),
}));
vi.mock("@/components/resource-editor/ResourceEditorShell", () => ({
  ResourceEditorShell: ({
    kind,
    orgSlug,
    onApplied,
  }: {
    kind: string;
    orgSlug: string;
    onApplied: () => void;
  }) => (
    <button
      type="button"
      data-testid="resource-editor"
      data-kind={kind}
      data-org={orgSlug}
      onClick={onApplied}
    >
      apply resource
    </button>
  ),
}));

const mockAvailableAgents = [
  { name: "Claude Code", slug: "claude-code", is_builtin: true, is_active: true },
];
const mockCompatibleRunner = {
  id: 1,
  node_id: "runner-claude",
  current_pods: 0,
  max_concurrent_pods: 5,
  status: "online" as const,
  available_agents: ["claude-code"],
  is_enabled: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};
const { mockUsePodCreationData } = vi.hoisted(() => ({ mockUsePodCreationData: vi.fn() }));
vi.mock("@/components/pod/hooks", () => ({ usePodCreationData: mockUsePodCreationData }));
vi.mock("@/components/ide/hooks", () => ({
  useConfigOptions: () => ({
    fields: [],
    loading: false,
    config: {},
    updateConfig: vi.fn(),
    resetConfig: vi.fn(),
  }),
}));

const { mockListEnvBundles } = vi.hoisted(() => ({ mockListEnvBundles: vi.fn() }));
vi.mock("@/lib/api/facade/envBundleConnect", () => ({ listEnvBundles: mockListEnvBundles }));

const { mockGetCatalog, mockListOrganizationEffectiveResources } = vi.hoisted(() => ({
  mockGetCatalog: vi.fn(),
  mockListOrganizationEffectiveResources: vi.fn(),
}));
vi.mock("@/lib/api/facade/aiResourceConnect", () => ({
  getCatalog: mockGetCatalog,
  listOrganizationEffectiveResources: mockListOrganizationEffectiveResources,
  listPersonalEffectiveResources: vi.fn(),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
  NextIntlClientProvider: ({ children }: { children: React.ReactNode }) => children,
}));
vi.mock("@/components/pod/CreatePodForm/PromptInput", () => ({
  PromptInput: ({ value, onChange, placeholder }: {
    value: string;
    onChange: (v: string) => void;
    placeholder?: string;
  }) => (
    <textarea
      data-testid="prompt-input"
      value={value}
      placeholder={placeholder}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));
vi.mock("@/components/pod/CreatePodForm/AdvancedOptions", () => ({
  AdvancedOptions: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() } }));
vi.mock("../WorkflowNlCreate", () => ({ WorkflowNlCreate: () => null }));
vi.mock("@/components/ide/ConfigForm", () => ({ ConfigForm: () => <div data-testid="config-form" /> }));
vi.mock("@/components/ui/responsive-dialog", () => ({
  ResponsiveDialog: ({ children }: { children: React.ReactNode }) => <div data-testid="dialog">{children}</div>,
  ResponsiveDialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResponsiveDialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResponsiveDialogTitle: ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>,
  ResponsiveDialogBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResponsiveDialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

const runtimeBundle = {
  id: BigInt(3),
  agentSlug: "claude-code",
  name: "dev-preferences",
  kind: "runtime",
  kindPrimary: false,
  isActive: true,
  configuredFields: ["LOG_LEVEL"],
  createdAt: "x",
  updatedAt: "x",
};

const anthropicProvider: ProviderDefinition = {
  key: "anthropic",
  displayName: "Anthropic",
  modalities: ["chat"],
  credentialFields: [],
  defaultBaseUrl: "https://api.anthropic.com",
  protocolAdapter: "anthropic",
  supportsCustomEndpoint: true,
  supportsModelDiscovery: false,
};
const claudeResource: EffectiveResource = {
  selectable: true,
  blockingReason: "",
  connection: {
    id: 7,
    ownerScope: "organization",
    identifier: "team-anthropic",
    providerKey: "anthropic",
    name: "Team Anthropic",
    baseUrl: "https://api.anthropic.com",
    configuredFields: ["api_key"],
    status: "validated",
    isEnabled: true,
    validationError: "",
    canManage: true,
    resources: [],
  },
  resource: {
    id: 42,
    providerConnectionId: 7,
    identifier: "claude-sonnet",
    modelId: "claude-sonnet-4",
    displayName: "Claude Sonnet",
    modalities: ["chat"],
    capabilities: ["text-generation"],
    defaultModalities: [],
    status: "validated",
    isEnabled: true,
    validationError: "",
  },
};

async function waitForAsyncEffects() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0));
  });
}

function mockResources(resources: EffectiveResource[] = [claudeResource]) {
  mockGetCatalog.mockResolvedValue([anthropicProvider]);
  mockListOrganizationEffectiveResources.mockResolvedValue(resources);
}

describe("WorkflowCreateDialog — model resources and runtime bundles", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUsePodCreationData.mockReturnValue({
      runners: [mockCompatibleRunner],
      repositories: [],
      loading: false,
      selectedRunner: null,
      setSelectedRunnerId: vi.fn(),
      availableAgents: mockAvailableAgents,
      agents: mockAvailableAgents,
      error: null,
    });
    mockListEnvBundles.mockResolvedValue({ items: [runtimeBundle], total: 1 });
    mockResources();
    mockCreateWorkflow.mockResolvedValue({ workflow: { slug: "nightly-ci" } });
    mockUpdateWorkflow.mockResolvedValue({ slug: "nightly-ci" });
  });

  it("uses the resource-native Workflow editor for create mode", () => {
    const onOpenChange = vi.fn();
    const onCreated = vi.fn();
    render(
      <WorkflowCreateDialog
        open
        onOpenChange={onOpenChange}
        onCreated={onCreated}
      />,
    );

    const editor = screen.getByTestId("resource-editor");
    expect(editor).toHaveAttribute("data-kind", "Workflow");
    expect(editor).toHaveAttribute("data-org", "acme");
    editor.click();

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onCreated).toHaveBeenCalledTimes(1);
  });

  it("edit mode preserves saved model_resource_id and runtime bundles", async () => {
    const editWorkflow: WorkflowData = {
      id: 5,
      organization_id: 1,
      slug: "nightly",
      name: "Nightly",
      permission_mode: "bypassPermissions",
      prompt_template: "run tests",
      agent_slug: "claude-code",
      model_resource_id: 42,
      used_env_bundles: ["dev-preferences"],
      execution_mode: "autopilot",
      status: "enabled",
      sandbox_strategy: "persistent",
      session_persistence: true,
      concurrency_policy: "skip",
      max_concurrent_runs: 1,
      max_retained_runs: 0,
      timeout_minutes: 60,
      created_by_id: 1,
      total_runs: 0,
      successful_runs: 0,
      failed_runs: 0,
      active_run_count: 0,
      autopilot_config: {},
      created_at: "x",
      updated_at: "x",
    };

    render(<WorkflowCreateDialog open onOpenChange={() => {}} onCreated={() => {}} editWorkflow={editWorkflow} />);
    await waitForAsyncEffects();

    expect(screen.getByLabelText("ide.createPod.selectModelResource")).toHaveTextContent("Claude Sonnet");
    expect((screen.getByRole("checkbox", { name: /dev-preferences/i }) as HTMLInputElement).checked).toBe(true);
  });
});
