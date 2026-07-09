import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";
import { LoopCreateDialog } from "../LoopCreateDialog";
import type { LoopData } from "@/lib/viewModels/loop";

// --- store / data hook mocks ---------------------------------------------

const mockCreateLoop = vi.fn();
const mockUpdateLoop = vi.fn();
vi.mock("@/stores/loop", () => ({
  useLoopStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ createLoop: mockCreateLoop, updateLoop: mockUpdateLoop }),
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
const { mockUsePodCreationData } = vi.hoisted(() => ({
  mockUsePodCreationData: vi.fn(),
}));
vi.mock("@/components/pod/hooks", () => ({
  usePodCreationData: mockUsePodCreationData,
}));

vi.mock("@/components/ide/hooks", () => ({
  useConfigOptions: () => ({
    fields: [],
    loading: false,
    config: {},
    updateConfig: vi.fn(),
    resetConfig: vi.fn(),
  }),
}));

// --- EnvBundleService mock --------------------------------------------------
// useLoopEnvBundles calls listEnvBundles({kind:"credential"}) + listEnvBundles({kind:"runtime"})
// in parallel. The mock dispatches by kind so each query returns its own list.

const { mockListEnvBundles } = vi.hoisted(() => ({
  mockListEnvBundles: vi.fn(),
}));
vi.mock("@/lib/api/facade/envBundleConnect", () => ({
  listEnvBundles: mockListEnvBundles,
}));

// --- Stubs for visual/dialog/intl/toast deps -------------------------------

vi.mock("next-intl", () => ({
  // PromptInput / AdvancedOptions call useTranslations(namespace?) themselves.
  useTranslations: () => (key: string) => key,
  NextIntlClientProvider: ({ children }: { children: React.ReactNode }) => children,
}));

vi.mock("@/components/pod/CreatePodForm/PromptInput", () => ({
  PromptInput: ({
    value,
    onChange,
    placeholder,
  }: {
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
  AdvancedOptions: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="advanced-options">{children}</div>
  ),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}));

vi.mock("../LoopNlCreate", () => ({
  LoopNlCreate: () => null,
}));

vi.mock("@/components/ide/ConfigForm", () => ({
  ConfigForm: () => <div data-testid="config-form" />,
}));

vi.mock("@/components/ui/responsive-dialog", () => ({
  ResponsiveDialog: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="dialog">{children}</div>
  ),
  ResponsiveDialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResponsiveDialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResponsiveDialogTitle: ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>,
  ResponsiveDialogBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResponsiveDialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/components/pod/CreatePodForm/AdvancedOptions", () => ({
  AdvancedOptions: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// --- Test data --------------------------------------------------------------

const bundleWork = {
  id: BigInt(1),
  agentSlug: "claude-code",
  name: "Work",
  kind: "credential",
  kindPrimary: true,
  isActive: true,
  configuredFields: ["ANTHROPIC_API_KEY"],
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

const bundlePersonal = {
  id: BigInt(2),
  agentSlug: "claude-code",
  name: "Personal",
  kind: "credential",
  kindPrimary: false,
  isActive: true,
  configuredFields: ["ANTHROPIC_API_KEY"],
  createdAt: "x",
  updatedAt: "x",
};

const bundleDevPreferences = {
  id: BigInt(3),
  agentSlug: "claude-code",
  name: "dev-preferences",
  kind: "runtime",
  kindPrimary: false,
  isActive: true,
  configuredFields: ["ANTHROPIC_MODEL", "LOG_LEVEL"],
  createdAt: "x",
  updatedAt: "x",
};

const bundleProxyStaging = {
  id: BigInt(4),
  agentSlug: "claude-code",
  name: "proxy-staging",
  kind: "runtime",
  kindPrimary: false,
  isActive: true,
  configuredFields: ["HTTPS_PROXY"],
  createdAt: "x",
  updatedAt: "x",
};

function fillRequiredFields() {
  fireEvent.change(screen.getByPlaceholderText("daily-code-review"), {
    target: { value: "Nightly CI" },
  });
  const prompt = screen.getByPlaceholderText("loops.promptPlaceholder") as HTMLTextAreaElement;
  fireEvent.change(prompt, { target: { value: "run tests" } });
}

async function waitForBundlesLoaded() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0));
  });
}

function mockBundleList(creds: unknown[], runtimes: unknown[] = []) {
  mockListEnvBundles.mockImplementation(async (opts?: { kind?: string }) => {
    if (opts?.kind === "credential") return { items: creds, total: creds.length };
    if (opts?.kind === "runtime") return { items: runtimes, total: runtimes.length };
    return { items: [], total: 0 };
  });
}

/**
 * Interact with the custom (non-native) Select in `@/components/ui/select`:
 * options render as `role="option"` with `data-option-value` only while the
 * trigger is open. Clicking the trigger opens it, then the matching option is
 * clicked to fire `onValueChange`.
 */
function pickCustomSelect(trigger: HTMLElement, optionValue: string) {
  fireEvent.click(trigger);
  const option = screen
    .getAllByRole("option")
    .find((el) => el.getAttribute("data-option-value") === optionValue);
  if (!option) {
    throw new Error(`Select option not found: "${optionValue}"`);
  }
  fireEvent.click(option);
}

// ---------------------------------------------------------------------------

describe("LoopCreateDialog — EnvBundle binding", () => {
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
    mockBundleList([bundleWork], [bundleDevPreferences]);
    mockCreateLoop.mockResolvedValue({ loop: { slug: "nightly-ci" } });
    mockUpdateLoop.mockResolvedValue({ slug: "nightly-ci" });
  });

  it("disables saving when the loop agent has no compatible runner", async () => {
    mockUsePodCreationData.mockReturnValue({
      runners: [{ ...mockCompatibleRunner, available_agents: ["codex-cli"] }],
      repositories: [],
      loading: false,
      selectedRunner: null,
      setSelectedRunnerId: vi.fn(),
      availableAgents: [],
      agents: mockAvailableAgents,
      error: null,
    });
    const editLoop: LoopData = {
      id: 5,
      organization_id: 1,
      slug: "nightly",
      name: "Nightly",
      prompt_template: "run tests",
      agent_slug: "claude-code",
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

    render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} editLoop={editLoop} />
    );

    expect(screen.getByRole("button", { name: "common.save" })).toBeDisabled();
  });

  it("renders both credential single-select AND runtime multi-select after an agent is chosen", async () => {
    const { container } = render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} />
    );

    const agentSelect = container.querySelector("#worker-image-select") as HTMLElement;
    expect(agentSelect).toBeTruthy();
    pickCustomSelect(agentSelect, "claude-code");

    await waitForBundlesLoaded();

    // Credential picker is a labeled custom Select; runtime picker uses checkbox list.
    expect(screen.getByLabelText("ide.createPod.selectCredential")).toBeInTheDocument();
    expect(screen.getByText("ide.createPod.selectRuntimeBundles")).toBeInTheDocument();
    // The runtime bundle appears as a checkbox row.
    expect(screen.getByText("dev-preferences")).toBeInTheDocument();
  });

  it("submits used_env_bundles=[credName] when only a credential is selected", async () => {
    const { container } = render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} />
    );

    const agentSelect = container.querySelector("#worker-image-select") as HTMLElement;
    pickCustomSelect(agentSelect, "claude-code");
    await waitForBundlesLoaded();

    fillRequiredFields();

    pickCustomSelect(screen.getByLabelText("ide.createPod.selectCredential"), "Work");

    const createBtn = screen.getByRole("button", { name: "loops.createLoop" });
    await act(async () => {
      fireEvent.click(createBtn);
    });

    await waitFor(() => expect(mockCreateLoop).toHaveBeenCalledTimes(1));
    const payload = mockCreateLoop.mock.calls[0][0];
    expect(payload.used_env_bundles).toEqual(["Work"]);
  });

  it("submits used_env_bundles=[] when no bundle is selected (default auth + no runtime)", async () => {
    const { container } = render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} />
    );

    const agentSelect = container.querySelector("#worker-image-select") as HTMLElement;
    pickCustomSelect(agentSelect, "claude-code");
    await waitForBundlesLoaded();

    fillRequiredFields();
    // Force-clear the credential select (auto-default may have set a primary).
    pickCustomSelect(screen.getByLabelText("ide.createPod.selectCredential"), "");

    const createBtn = screen.getByRole("button", { name: "loops.createLoop" });
    await act(async () => {
      fireEvent.click(createBtn);
    });

    await waitFor(() => expect(mockCreateLoop).toHaveBeenCalledTimes(1));
    const payload = mockCreateLoop.mock.calls[0][0];
    expect(payload.used_env_bundles).toEqual([]);
  });

  it("merges credential first then runtime bundles when both are selected", async () => {
    mockBundleList([bundleWork], [bundleDevPreferences, bundleProxyStaging]);
    const { container } = render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} />
    );

    const agentSelect = container.querySelector("#worker-image-select") as HTMLElement;
    pickCustomSelect(agentSelect, "claude-code");
    await waitForBundlesLoaded();
    fillRequiredFields();

    // Pick a credential.
    pickCustomSelect(screen.getByLabelText("ide.createPod.selectCredential"), "Work");
    // Pick two runtimes in a specific order.
    fireEvent.click(screen.getByRole("checkbox", { name: /proxy-staging/i }));
    fireEvent.click(screen.getByRole("checkbox", { name: /dev-preferences/i }));

    const createBtn = screen.getByRole("button", { name: "loops.createLoop" });
    await act(async () => {
      fireEvent.click(createBtn);
    });

    await waitFor(() => expect(mockCreateLoop).toHaveBeenCalledTimes(1));
    const payload = mockCreateLoop.mock.calls[0][0];
    // Credential first, then runtime in selection order.
    expect(payload.used_env_bundles).toEqual(["Work", "proxy-staging", "dev-preferences"]);
  });

  it("edit mode: reconciles used_env_bundles back into credential + runtime state", async () => {
    mockBundleList([bundleWork, bundlePersonal], [bundleDevPreferences]);
    const editLoop: LoopData = {
      id: 5,
      organization_id: 1,
      slug: "nightly",
      name: "Nightly",
      permission_mode: "bypassPermissions",
      prompt_template: "run tests",
      agent_slug: "claude-code",
      used_env_bundles: ["Work", "dev-preferences"],
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

    render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} editLoop={editLoop} />
    );

    await waitForBundlesLoaded();

    // Custom Select trigger shows the selected credential's label.
    const credSelect = screen.getByLabelText("ide.createPod.selectCredential");
    expect(credSelect).toHaveTextContent("Work");

    const runtimeCheckbox = screen.getByRole("checkbox", {
      name: /dev-preferences/i,
    }) as HTMLInputElement;
    expect(runtimeCheckbox.checked).toBe(true);
  });

  it("edit mode: updating bundle picks survives the round-trip to updateLoop", async () => {
    mockBundleList([bundleWork, bundlePersonal], [bundleDevPreferences]);
    const editLoop: LoopData = {
      id: 5,
      organization_id: 1,
      slug: "nightly",
      name: "Nightly",
      permission_mode: "bypassPermissions",
      prompt_template: "run tests",
      agent_slug: "claude-code",
      used_env_bundles: ["Work"],
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

    render(
      <LoopCreateDialog open onOpenChange={() => {}} onCreated={() => {}} editLoop={editLoop} />
    );

    await waitForBundlesLoaded();

    // Swap credential Work → Personal.
    pickCustomSelect(screen.getByLabelText("ide.createPod.selectCredential"), "Personal");
    // Add a runtime bundle.
    fireEvent.click(screen.getByRole("checkbox", { name: /dev-preferences/i }));

    const saveBtn = screen.getByRole("button", { name: "common.save" });
    await act(async () => {
      fireEvent.click(saveBtn);
    });

    await waitFor(() => expect(mockUpdateLoop).toHaveBeenCalledTimes(1));
    const [slug, payload] = mockUpdateLoop.mock.calls[0];
    expect(slug).toBe("nightly");
    expect(payload.used_env_bundles).toEqual(["Personal", "dev-preferences"]);
  });
});
