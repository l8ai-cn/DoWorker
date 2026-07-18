import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ExpertDetailPane } from "../ExpertDetailPane";

const mocks = vi.hoisted(() => ({
  expert: null as Record<string, unknown> | null,
  fetchExpert: vi.fn(),
  runExpert: vi.fn(),
  deleteExpert: vi.fn(),
  clearError: vi.fn(),
  push: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mocks.push }),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/stores/expert", () => ({
  useCurrentExpert: () => mocks.expert,
  useExpertStore: (
    selector: (state: Record<string, unknown>) => unknown,
  ) =>
    selector({
      expertLoading: false,
      error: null,
      fetchExpert: mocks.fetchExpert,
      runExpert: mocks.runExpert,
      deleteExpert: mocks.deleteExpert,
      clearError: mocks.clearError,
    }),
}));

vi.mock("@/stores/pod", () => ({
  usePodStore: { getState: () => ({ upsertPod: vi.fn() }) },
}));

vi.mock("../ExpertConfigList", () => ({
  ExpertConfigList: () => <div data-testid="expert-config" />,
}));

vi.mock("../ExpertEditDrawer", () => ({
  ExpertEditDrawer: ({ open }: { open: boolean }) => (
    <div data-testid="legacy-editor" data-open={String(open)} />
  ),
}));

vi.mock("../ExpertRevisionDialog", () => ({
  ExpertRevisionDialog: ({
    open,
    onApplied,
  }: {
    open: boolean;
    onApplied: () => void;
  }) => (
    <button
      type="button"
      data-testid="revision-editor"
      data-open={String(open)}
      onClick={onApplied}
    >
      apply resource revision
    </button>
  ),
}));

vi.mock("@/components/ui/confirm-dialog", () => ({
  ConfirmDialog: ({ open }: { open: boolean }) => (
    <div data-testid="delete-confirm" data-open={String(open)} />
  ),
}));

function expert(overrides: Record<string, unknown> = {}) {
  return {
    id: 17,
    slug: "release-reviewer",
    name: "Release reviewer",
    description: "Reviews release candidates",
    agent_slug: "codex",
    interaction_mode: "pty",
    automation_level: "supervised",
    perpetual: false,
    used_env_bundles: [],
    skill_slugs: [],
    knowledge_mounts: [],
    run_count: 0,
    created_at: "2026-07-17T00:00:00Z",
    updated_at: "2026-07-17T00:00:00Z",
    ...overrides,
  };
}

describe("ExpertDetailPane", () => {
  beforeEach(() => {
    mocks.expert = expert();
    vi.clearAllMocks();
  });

  it.each([
    ["worker_spec_snapshot_id", 41],
    ["orchestration_resource_id", 52],
    ["orchestration_resource_revision", 3],
  ])("uses the revision editor when %s is present", (field, value) => {
    mocks.expert = expert({ [field]: value });

    render(<ExpertDetailPane slug="release-reviewer" orgSlug="acme" />);
    fireEvent.click(
      screen.getByRole("button", { name: "edit.editExpert" }),
    );

    expect(screen.getByTestId("revision-editor")).toHaveAttribute(
      "data-open",
      "true",
    );
    expect(screen.queryByTestId("legacy-editor")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "deleteExpert" }),
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByTestId("revision-editor"));
    expect(mocks.fetchExpert).toHaveBeenCalledWith("release-reviewer");
  });

  it("keeps legacy editing and deletion for an unmanaged expert", () => {
    render(<ExpertDetailPane slug="release-reviewer" orgSlug="acme" />);
    fireEvent.click(
      screen.getByRole("button", { name: "edit.editExpert" }),
    );

    expect(screen.getByTestId("legacy-editor")).toHaveAttribute(
      "data-open",
      "true",
    );
    expect(screen.queryByTestId("revision-editor")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "deleteExpert" }),
    ).toBeInTheDocument();
  });
});
