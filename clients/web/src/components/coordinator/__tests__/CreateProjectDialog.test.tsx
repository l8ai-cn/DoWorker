import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { CreateProjectDialog } from "../CreateProjectDialog";

const { mockCreateProject, mockRepositoryList } = vi.hoisted(() => ({
  mockCreateProject: vi.fn(),
  mockRepositoryList: vi.fn(),
}));

vi.mock("@/stores/coordinator", () => ({
  useCoordinatorStore: (selector: (state: { createProject: typeof mockCreateProject }) => unknown) =>
    selector({ createProject: mockCreateProject }),
}));

vi.mock("@/lib/api/facade/repository", () => ({
  repositoryApi: {
    list: mockRepositoryList,
  },
}));

vi.mock("@/components/pod/hooks", () => ({
  usePodCreationData: () => ({
    runners: [
      {
        id: 1,
        node_id: "runner-claude",
        current_pods: 0,
        max_concurrent_pods: 5,
        status: "online",
        available_agents: ["claude-code"],
        is_enabled: true,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ],
    repositories: [],
    loading: false,
    selectedRunner: null,
    setSelectedRunnerId: vi.fn(),
    availableAgents: [
      { name: "Claude Code", slug: "claude-code", is_builtin: true, is_active: true },
    ],
    agents: [
      { name: "Claude Code", slug: "claude-code", is_builtin: true, is_active: true },
    ],
    error: null,
  }),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("CreateProjectDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRepositoryList.mockResolvedValue({
      items: [
        {
          id: 42,
          name: "Repo One",
          slug: "repo-one",
          organization_id: 1,
          provider_type: "github",
          provider_base_url: "https://github.com",
          http_clone_url: "https://github.com/acme/repo-one.git",
          external_id: "acme/repo-one",
          default_branch: "main",
          visibility: "organization",
          is_active: true,
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
        },
      ],
    });
    mockCreateProject.mockResolvedValue(undefined);
  });

  it("submits an explicit compatible agent slug", async () => {
    render(<CreateProjectDialog open onOpenChange={() => {}} />);

    fireEvent.change(screen.getByPlaceholderText("create.namePlaceholder"), {
      target: { value: "Issue automation" },
    });
    fireEvent.click(screen.getByText("create.repositoryPlaceholder"));
    await screen.findByText("Repo One");
    fireEvent.click(screen.getByText("Repo One"));
    fireEvent.click(screen.getByText("create.agentPlaceholder"));
    fireEvent.click(screen.getByText("Claude Code"));

    fireEvent.click(screen.getByRole("button", { name: "create.submit" }));

    await waitFor(() => expect(mockCreateProject).toHaveBeenCalledTimes(1));
    expect(mockCreateProject).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Issue automation",
        repository_id: 42,
        agent_slug: "claude-code",
      })
    );
  });
});
