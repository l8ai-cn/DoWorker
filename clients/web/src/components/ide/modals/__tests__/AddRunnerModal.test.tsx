import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { AddRunnerModal } from "../AddRunnerModal";

const listExecutionClusters = vi.fn();
const createRegistrationCommand = vi.fn();

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/stores/auth", () => ({
  useCurrentOrg: () => ({ slug: "test-org" }),
}));

vi.mock("@/lib/api/connect/executionClusterConnect", () => ({
  listExecutionClusters: (...args: unknown[]) => listExecutionClusters(...args),
}));

vi.mock("@/lib/api/facade/executionClusterApi", () => ({
  listExecutionClusters: (...args: unknown[]) => listExecutionClusters(...args),
  createRegistrationCommand: (...args: unknown[]) =>
    createRegistrationCommand(...args),
}));

describe("AddRunnerModal", () => {
  const onClose = vi.fn();
  const onCreated = vi.fn();
  const command =
    "agent-cloud-runner register --server https://gateway.example.test --token one-time-secret";

  beforeEach(() => {
    vi.clearAllMocks();
    listExecutionClusters.mockResolvedValue([
      {
        id: 12,
        slug: "local",
        name: "本地集群",
        kind: "local",
        status: "ready",
        runnerCount: 0,
        onlineRunnerCount: 0,
        availableRunnerCount: 0,
        tunnelStatus: "connected",
      },
    ]);
    createRegistrationCommand.mockResolvedValue({
      command,
      expiresAt: "2026-07-12T12:15:00Z",
    });
  });

  it("does not render when closed", () => {
    render(
      <AddRunnerModal open={false} onClose={onClose} onCreated={onCreated} />,
    );

    expect(
      screen.queryByText("runners.addRunnerModal.title"),
    ).not.toBeInTheDocument();
  });

  it("requires an explicit cluster selection before generating a command", async () => {
    render(<AddRunnerModal open onClose={onClose} onCreated={onCreated} />);

    const generate = screen.getByRole("button", {
      name: "runners.addRunnerModal.generate",
    });
    await waitFor(() =>
      expect(listExecutionClusters).toHaveBeenCalledWith("test-org"),
    );
    expect(generate).toBeDisabled();
    expect(createRegistrationCommand).not.toHaveBeenCalled();
  });

  it("uses the selected cluster to request the server-signed command", async () => {
    render(<AddRunnerModal open onClose={onClose} onCreated={onCreated} />);

    await selectLocalCluster();
    fireEvent.click(
      screen.getByRole("button", {
        name: "runners.addRunnerModal.generate",
      }),
    );

    await waitFor(() => {
      expect(createRegistrationCommand).toHaveBeenCalledWith("test-org", 12);
    });
    expect(screen.getByText(command)).toBeInTheDocument();
  });

  it("copies the exact server-signed command", async () => {
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    });
    render(<AddRunnerModal open onClose={onClose} onCreated={onCreated} />);

    await selectLocalCluster();
    fireEvent.click(
      screen.getByRole("button", {
        name: "runners.addRunnerModal.generate",
      }),
    );
    await screen.findByText(command);

    const copyButtons = screen.getAllByText(
      "runners.addRunnerModal.copyCommand",
    );
    const commandCopyButton = copyButtons.find(
      (button) =>
        button.closest("div.bg-muted")?.querySelector("code")?.textContent ===
        command,
    );
    expect(commandCopyButton).toBeDefined();
    fireEvent.click(commandCopyButton!);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(command);
  });

  async function selectLocalCluster() {
    await screen.findByText("runners.addRunnerModal.clusterPlaceholder");
    fireEvent.click(
      screen.getByRole("button", {
        name: "runners.addRunnerModal.clusterPlaceholder",
      }),
    );
    fireEvent.click(await screen.findByRole("option", { name: "本地集群" }));
    await waitFor(() => {
      expect(
        screen.getByRole("button", {
          name: "runners.addRunnerModal.generate",
        }),
      ).toBeEnabled();
    });
  }
});
