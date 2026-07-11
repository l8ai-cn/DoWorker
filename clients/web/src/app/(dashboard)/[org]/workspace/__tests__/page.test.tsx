import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";
import { render, waitFor } from "@/test/test-utils";
import WorkspacePage from "../page";

const mockAddPane = vi.fn();
const mockReplace = vi.fn();
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
    addPane: typeof mockAddPane;
    _hasHydrated: boolean;
  }) => unknown) => selector({
    panes: [],
    addPane: mockAddPane,
    _hasHydrated: true,
  }),
}));

vi.mock("@/stores/pod", () => ({
  usePodStore: { getState: () => ({ upsertPod: vi.fn() }) },
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
    mockAddPane.mockReset();
    mockReplace.mockReset();
  });

  it("keeps the pod query after opening a Worker from a direct link", async () => {
    render(<WorkspacePage />);

    await waitFor(() => {
      expect(mockAddPane).toHaveBeenCalledWith("2-standalone-02467fd1");
    });

    expect(mockReplace).not.toHaveBeenCalled();
  });
});
