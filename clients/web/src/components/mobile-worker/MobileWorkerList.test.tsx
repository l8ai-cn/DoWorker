import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MobileWorkerList } from "./MobileWorkerList";

const mocks = vi.hoisted(() => ({
  fetchPods: vi.fn(),
  pods: [] as Record<string, unknown>[],
  error: null as string | null,
}));

vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme" }),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/stores/pod", () => {
  const usePodStore = Object.assign(
    (selector: (state: Record<string, unknown>) => unknown) =>
      selector({ fetchPods: mocks.fetchPods }),
    { getState: () => ({ error: mocks.error }) },
  );
  return {
    usePods: () => mocks.pods,
    usePodStore,
  };
});

describe("MobileWorkerList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.pods = [];
    mocks.error = null;
    mocks.fetchPods.mockResolvedValue(undefined);
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
  });

  it("loads active Workers and links to the canonical mobile route", async () => {
    mocks.pods = [{
      pod_key: "pod-1",
      status: "running",
      interaction_mode: "pty",
      title: "Build Worker",
    }];

    render(<MobileWorkerList />);

    expect(await screen.findByText("Build Worker")).toBeInTheDocument();
    expect(mocks.fetchPods).toHaveBeenCalledWith({
      status: "running,initializing,paused,disconnected,orphaned",
    });
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/acme/mobile/workers/pod-1",
    );
    expect(screen.getByText("status.running")).toBeInTheDocument();
    expect(screen.getByText("terminal")).toBeInTheDocument();
  });

  it("renders the empty state after a successful load", async () => {
    render(<MobileWorkerList />);

    expect(await screen.findByText("emptyTitle")).toBeInTheDocument();
    expect(screen.getByText("emptyBody")).toBeInTheDocument();
  });

  it("shows an error and retries the same request", async () => {
    mocks.error = "relay unavailable";

    render(<MobileWorkerList />);

    expect(await screen.findByText("relay unavailable")).toBeInTheDocument();
    mocks.error = null;
    fireEvent.click(screen.getByRole("button", { name: "retry" }));

    await waitFor(() => expect(mocks.fetchPods).toHaveBeenCalledTimes(2));
  });
});
