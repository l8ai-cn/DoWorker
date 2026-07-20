import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MobileWorkerList } from "./MobileWorkerList";

const mocks = vi.hoisted(() => ({
  fetchMobileWorkerPods: vi.fn(),
  pods: [] as Record<string, unknown>[],
}));

vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme" }),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("./mobileWorkerPodQuery", () => ({
  fetchMobileWorkerPods: mocks.fetchMobileWorkerPods,
  useMobileWorkerPods: () => mocks.pods,
}));

describe("MobileWorkerList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.pods = [];
    mocks.fetchMobileWorkerPods.mockResolvedValue(undefined);
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
  });

  it("loads readable Workers and links to the canonical mobile route", async () => {
    mocks.pods = [{
      pod_key: "pod-1",
      status: "running",
      interaction_mode: "pty",
      title: "Build Worker",
    }];

    render(<MobileWorkerList />);

    expect(await screen.findByText("Build Worker")).toBeInTheDocument();
    expect(mocks.fetchMobileWorkerPods).toHaveBeenCalledWith("acme");
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/acme/mobile/workers/pod-1",
    );
    expect(screen.getByText("running")).toBeInTheDocument();
    expect(screen.getByText("terminal")).toBeInTheDocument();
  });

  it("keeps a completed ACP Worker reachable from the mobile list", async () => {
    mocks.pods = [{
      pod_key: "pod-complete",
      status: "completed",
      interaction_mode: "acp",
      title: "Pattern Seamless Run R3",
    }];

    render(<MobileWorkerList />);

    expect(await screen.findByText("Pattern Seamless Run R3")).toBeInTheDocument();
    expect(screen.getByText("completed")).toBeInTheDocument();
    expect(screen.getByText("acp")).toBeInTheDocument();
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/acme/mobile/workers/pod-complete",
    );
  });

  it("renders the empty state after a successful load", async () => {
    render(<MobileWorkerList />);

    expect(await screen.findByText("emptyTitle")).toBeInTheDocument();
    expect(screen.getByText("emptyBody")).toBeInTheDocument();
  });

  it("shows an error and retries the same request", async () => {
    mocks.fetchMobileWorkerPods.mockRejectedValueOnce(new Error("relay unavailable"));

    render(<MobileWorkerList />);

    expect(await screen.findByText("relay unavailable")).toBeInTheDocument();
    mocks.fetchMobileWorkerPods.mockResolvedValueOnce(undefined);
    fireEvent.click(screen.getByRole("button", { name: "retry" }));

    await waitFor(() => expect(mocks.fetchMobileWorkerPods).toHaveBeenCalledTimes(2));
  });
});
