import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { CreatePodForm } from "../index";
import {
  mockSetPrompt,
  mockSetAlias,
  mockSetSelectedModelResourceId,
  defaultPodCreationData,
  defaultFormState,
  defaultConfigOptions,
  mockRunner,
  mockAgent,
  mockRepository,
  clearAllMocks,
  pickSelectOption,
} from "./test-utils";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";

vi.mock("../../hooks", () => ({
  usePodCreationData: vi.fn(() => defaultPodCreationData),
  useCreatePodForm: vi.fn(() => defaultFormState),
}));

vi.mock("@/components/ide/hooks", () => ({
  useConfigOptions: vi.fn(() => defaultConfigOptions),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/components/ide/ConfigForm", () => ({
  ConfigForm: () => <div data-testid="config-form">Config Form</div>,
}));

vi.mock("../KnowledgeBaseMountSelect", () => ({
  KnowledgeBaseMountSelect: () => <div data-testid="knowledge-base-mount-select" />,
}));

vi.mock("@/components/experts/ExpertPickerSection", () => ({
  ExpertPickerSection: () => <div data-testid="expert-picker-section" />,
}));

vi.mock("../WorkerRepositoryField", () => ({
  WorkerRepositoryField: ({
    value,
    onChange,
  }: {
    value: number | null;
    onChange: (v: number | null) => void;
  }) => (
    <div>
      <label htmlFor="repository-select">ide.createPod.selectRepository</label>
      <select
        id="repository-select"
        value={value ?? ""}
        onChange={(e) => onChange(e.target.value ? Number(e.target.value) : null)}
      >
        <option value="">ide.createPod.selectRepositoryPlaceholder</option>
        <option value="1">org/repo1</option>
        <option value="2">org/repo2</option>
      </select>
    </div>
  ),
}));

vi.mock("@/lib/terminal-size", () => ({
  estimateWorkspaceTerminalSize: () => ({ cols: 80, rows: 24 }),
}));

// Mock Collapsible to always render children (no collapse animation in tests)
vi.mock("@/components/ui/collapsible", () => ({
  Collapsible: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  CollapsibleTrigger: ({ children }: { children: React.ReactNode }) => <button>{children}</button>,
  CollapsibleContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/stores/podCreation", () => ({
  usePodCreationStore: () => ({
    lastAgentSlug: null,
    lastRepositoryId: null,
    lastCredentialName: "",
    lastRuntimeBundleNames: [],
    lastBranchName: null,
    lastSkillSlugs: [],
    setLastChoices: vi.fn(),
    clearLastChoices: vi.fn(),
    _hasHydrated: true,
    setHasHydrated: vi.fn(),
  }),
}));

import { usePodCreationData, useCreatePodForm } from "../../hooks";
import { useConfigOptions } from "@/components/ide/hooks";

const workspaceConfig = { scenario: "workspace" as const };

const openAIChatResource: EffectiveResource = {
  selectable: true,
  blockingReason: "",
  connection: {
    id: 1,
    ownerScope: "user",
    identifier: "openai-main",
    providerKey: "openai",
    name: "OpenAI",
    baseUrl: "https://api.openai.com/v1",
    configuredFields: ["api_key"],
    status: "valid",
    isEnabled: true,
    validationError: "",
    canManage: true,
    resources: [],
  },
  resource: {
    id: 42,
    providerConnectionId: 1,
    identifier: "gpt-5",
    modelId: "gpt-5",
    displayName: "GPT-5",
    modalities: ["chat"],
    capabilities: ["text-generation"],
    defaultModalities: ["chat"],
    status: "valid",
    isEnabled: true,
    validationError: "",
  },
};

describe("CreatePodForm - Agent Configuration", () => {
  beforeEach(() => {
    clearAllMocks();
    vi.clearAllMocks();
  });

  const setupAgentSelectedState = (overrides = {}) => {
    const mockSetSelectedRepository = vi.fn();
    const mockSetSelectedBranch = vi.fn();
    const mockSetSelectedRuntimeBundleNames = vi.fn();

    vi.mocked(usePodCreationData).mockReturnValue({
      ...defaultPodCreationData,
      runners: [mockRunner],
      repositories: [mockRepository, { ...mockRepository, id: 2, slug: "org/repo2" }],
      selectedRunner: mockRunner,
      availableAgents: [mockAgent],
    });

    vi.mocked(useCreatePodForm).mockReturnValue({
      ...defaultFormState,
      selectedAgent: "claude-code",
      envBundles: [
        { id: 3, agent_slug: "claude-code", name: "dev-preferences", kind: "runtime", kind_primary: false },
      ],
      modelResources: [openAIChatResource],
      setSelectedRepository: mockSetSelectedRepository,
      setSelectedBranch: mockSetSelectedBranch,
      setSelectedRuntimeBundleNames: mockSetSelectedRuntimeBundleNames,
      selectedAgentSlug: "claude-code",
      isValid: true,
      ...overrides,
    });

    return {
      mockSetSelectedRepository,
      mockSetSelectedBranch,
      mockSetSelectedRuntimeBundleNames,
    };
  };

  describe("AI model resource selection", () => {
    it("renders model resource select without default credential option", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.selectModelResource")).toBeInTheDocument();
      expect(screen.queryByLabelText("ide.createPod.selectCredential")).not.toBeInTheDocument();
    });

    it("lists selectable model resources", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      fireEvent.click(screen.getByLabelText("ide.createPod.selectModelResource"));
      const options = screen
        .getAllByRole("option")
        .map((el) => el.getAttribute("data-option-value") ?? "");
      expect(options).toContain("42");
    });

    it("calls setSelectedModelResourceId when a resource is picked", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      pickSelectOption("ide.createPod.selectModelResource", "42");
      expect(mockSetSelectedModelResourceId).toHaveBeenCalledWith(42);
    });

    it("shows no-resource blocking state", () => {
      setupAgentSelectedState({ modelResources: [] });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("ide.createPod.noModelResourcesAvailableHint")).toBeInTheDocument();
    });

    it("shows model resource load error", () => {
      setupAgentSelectedState({ modelResourceError: "AI Resource RPC failed" });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByRole("alert")).toHaveTextContent("AI Resource RPC failed");
    });
  });

  describe("Runtime bundle multi-select", () => {
    it("renders the runtime bundle picker", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("ide.createPod.selectRuntimeBundles")).toBeInTheDocument();
    });

    it("lists only runtime-kind bundles as checkbox rows (excludes credentials)", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      // Runtime bundle shows up.
      expect(screen.getByText("dev-preferences")).toBeInTheDocument();
      // Credential bundles only appear in the credential select <option> elements
      // — never as checkbox rows. There must be exactly zero checkboxes labeled
      // with a credential name.
      const checkboxes = screen.getAllByRole("checkbox");
      const checkboxLabels = checkboxes.map((cb) => cb.getAttribute("aria-label") ?? "");
      expect(checkboxLabels.every((l) => !l.includes("My Credentials"))).toBe(true);
    });

    it("toggles selection through the runtime row checkbox", () => {
      const { mockSetSelectedRuntimeBundleNames } = setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      const checkboxes = screen.getAllByRole("checkbox");
      // Only one runtime bundle in this fixture; click it.
      fireEvent.click(checkboxes[0]);
      expect(mockSetSelectedRuntimeBundleNames).toHaveBeenCalledWith(["dev-preferences"]);
    });

    it("shows merge-order hint when a runtime bundle is selected", () => {
      setupAgentSelectedState({ selectedRuntimeBundleNames: ["dev-preferences"] });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("ide.createPod.multiBundleHint")).toBeInTheDocument();
    });

    it("shows loading state while bundles load", () => {
      setupAgentSelectedState({ loadingBundles: true });
      const { container } = render(<CreatePodForm config={{ scenario: "workspace" }} />);
      // Both pickers render the same loading affordance.
      expect(screen.getAllByText("common.loading").length).toBeGreaterThan(0);
      expect(container.querySelectorAll(".animate-spin").length).toBeGreaterThan(0);
    });
  });

  describe("repository selection", () => {
    it("should render repository select on runtime step", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.selectRepository")).toBeInTheDocument();
    });

    it("should render repositories in select", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("org/repo1")).toBeInTheDocument();
      expect(screen.getByText("org/repo2")).toBeInTheDocument();
    });

    it("should call setSelectedRepository when changed", () => {
      const { mockSetSelectedRepository } = setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      fireEvent.change(screen.getByLabelText("ide.createPod.selectRepository"), { target: { value: "1" } });
      expect(mockSetSelectedRepository).toHaveBeenCalledWith(1);
    });

    it("should call setSelectedRepository with null when deselected", () => {
      const { mockSetSelectedRepository } = setupAgentSelectedState({ selectedRepository: 1 });
      render(<CreatePodForm config={workspaceConfig} />);
      fireEvent.change(screen.getByLabelText("ide.createPod.selectRepository"), { target: { value: "" } });
      expect(mockSetSelectedRepository).toHaveBeenCalledWith(null);
    });
  });

  describe("branch input", () => {
    it("should render branch input when repository is selected", () => {
      setupAgentSelectedState({ selectedRepository: 1 });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.branch")).toBeInTheDocument();
    });

    it("should not render branch input when no repository is selected", () => {
      setupAgentSelectedState({ selectedRepository: null });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.queryByLabelText("ide.createPod.branch")).not.toBeInTheDocument();
    });

    it("should call setSelectedBranch when changed", () => {
      const { mockSetSelectedBranch } = setupAgentSelectedState({ selectedRepository: 1 });
      render(<CreatePodForm config={workspaceConfig} />);
      fireEvent.change(screen.getByLabelText("ide.createPod.branch"), { target: { value: "feature/test" } });
      expect(mockSetSelectedBranch).toHaveBeenCalledWith("feature/test");
    });

    it("should show branch validation error", () => {
      setupAgentSelectedState({
        selectedRepository: 1,
        validationErrors: { branch: "Branch is required" },
      });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("Branch is required")).toBeInTheDocument();
    });
  });

  describe("prompt textarea", () => {
    it("should render prompt textarea when agent is selected", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.prompt")).toBeInTheDocument();
    });

    it("should use custom placeholder when provided", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={{ scenario: "workspace", promptPlaceholder: "Custom placeholder" }} />);
      expect(screen.getByPlaceholderText("Custom placeholder")).toBeInTheDocument();
    });

    it("should call setPrompt when changed", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      fireEvent.change(screen.getByLabelText("ide.createPod.prompt"), { target: { value: "New prompt" } });
      expect(mockSetPrompt).toHaveBeenCalledWith("New prompt");
    });
  });

  describe("alias input", () => {
    it("should render alias input when agent is selected", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.alias")).toBeInTheDocument();
    });

    it("should call setAlias when changed", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      fireEvent.change(screen.getByLabelText("ide.createPod.alias"), { target: { value: "my-pod" } });
      expect(mockSetAlias).toHaveBeenCalledWith("my-pod");
    });

    it("should show alias value from form state", () => {
      setupAgentSelectedState({ alias: "existing-alias" });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.alias")).toHaveValue("existing-alias");
    });

    it("should have maxLength of 100", () => {
      setupAgentSelectedState();
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByLabelText("ide.createPod.alias")).toHaveAttribute("maxLength", "100");
    });

    it("should not render alias input when no agent is selected", () => {
      vi.mocked(useCreatePodForm).mockReturnValue(defaultFormState);
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.queryByLabelText("ide.createPod.alias")).not.toBeInTheDocument();
    });
  });

  describe("config options", () => {
    it("should show loading state for config", () => {
      setupAgentSelectedState();
      vi.mocked(useConfigOptions).mockReturnValue({ ...defaultConfigOptions, loading: true });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("ide.createPod.loadingPlugins")).toBeInTheDocument();
    });

    it("should render config form when config fields are available", () => {
      setupAgentSelectedState();
      vi.mocked(useConfigOptions).mockReturnValue({
        ...defaultConfigOptions,
        fields: [{ name: "model", type: "select" }],
      });
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.getByText("ide.createPod.pluginConfig")).toBeInTheDocument();
      expect(screen.getByTestId("config-form")).toBeInTheDocument();
    });

    it("should not render config form when no config fields available", () => {
      setupAgentSelectedState();
      vi.mocked(useConfigOptions).mockReturnValue(defaultConfigOptions);
      render(<CreatePodForm config={workspaceConfig} />);
      expect(screen.queryByText("ide.createPod.pluginConfig")).not.toBeInTheDocument();
    });
  });
});
