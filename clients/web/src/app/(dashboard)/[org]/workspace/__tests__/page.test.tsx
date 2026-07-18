import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";
import { render, waitFor } from "@/test/test-utils";
import WorkspacePage from "../page";

const mockOpenDeepLinkedPane = vi.fn();
const mockReplace = vi.fn();
const mockFetchPod = vi.fn();
let mockSearch = "pod=2-standalone-02467fd1";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn(), replace: mockReplace }),
  useSearchParams: () => new URLSearchParams(mockSearch),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
  NextIntlClientProvider: ({ children }: { children: ReactNode }) => children,
}));

vi.mock("@/stores/auth", () => ({
  useCurrentOrg: () => ({ slug: "dev-org" }),
}));

vi.mock("@/stores/workspace", () => ({
  useWorkspaceStore: (selector: (state: {
    panes: never[];
    addPane: ReturnType<typeof vi.fn>;
    openDeepLinkedPane: typeof mockOpenDeepLinkedPane;
    _hasHydrated: boolean;
  }) => unknown) => selector({
    panes: [],
    addPane: vi.fn(),
    openDeepLinkedPane: mockOpenDeepLinkedPane,
    _hasHydrated: true,
  }),
}));

vi.mock("@/stores/pod", () => ({
  usePodStore: Object.assign(
    (selector: (state: { fetchPod: typeof mockFetchPod }) => unknown) =>
      selector({ fetchPod: mockFetchPod }),
    { getState: () => ({ upsertPod: vi.fn() }) },
  ),
}));

vi.mock("@/components/workspace", () => ({
  WorkspaceManager: () => <div data-testid="workspace-manager" />,
}));

vi.mock("@/components/workspace/WorkspaceEmptyState", () => ({
  WorkspaceEmptyState: () => <div data-testid="workspace-empty-state" />,
}));

vi.mock("@/components/ide/CreatePodModal", () => ({
  CreatePodModal: () => null,
}));

describe("WorkspacePage", () => {
  beforeEach(() => {
    mockSearch = "pod=2-standalone-02467fd1";
    mockOpenDeepLinkedPane.mockReset();
    mockReplace.mockReset();
    mockFetchPod.mockReset();
    mockFetchPod.mockResolvedValue(undefined);
  });

  it("keeps the pod query after opening a Worker from a direct link", async () => {
    render(<WorkspacePage />);

    await waitFor(() => {
      expect(mockOpenDeepLinkedPane).toHaveBeenCalledWith("2-standalone-02467fd1");
      expect(mockFetchPod).toHaveBeenCalledWith("2-standalone-02467fd1");
    });

    expect(mockReplace).not.toHaveBeenCalled();
  });
});
