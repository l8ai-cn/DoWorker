import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  getExpertMarketUpgrade,
  upgradeExpertFromMarket,
} from "@/lib/api/expertMarketApi";
import { ExpertMarketUpgradePanel } from "./ExpertMarketUpgradePanel";

vi.mock("@/lib/api/expertMarketApi", () => ({
  getExpertMarketUpgrade: vi.fn(),
  upgradeExpertFromMarket: vi.fn(),
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => ({
    current: "Up to date",
    upgrade: "Upgrade expert",
    upgrading: "Upgrading",
    retry: "Retry",
  })[key] ?? key,
}));

describe("ExpertMarketUpgradePanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows an explicit current state when no upgrade exists", async () => {
    vi.mocked(getExpertMarketUpgrade).mockResolvedValue({ upgrade_available: false });

    render(<ExpertMarketUpgradePanel expertSlug="video-editor" onUpgraded={vi.fn()} />);

    expect(await screen.findByText("Up to date")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Upgrade expert" })).not.toBeInTheDocument();
  });

  it("upgrades once and disables the action while busy", async () => {
    vi.mocked(getExpertMarketUpgrade)
      .mockResolvedValueOnce({ upgrade_available: true })
      .mockResolvedValueOnce({ upgrade_available: false });
    let resolveUpgrade: (() => void) | undefined;
    vi.mocked(upgradeExpertFromMarket).mockReturnValue(new Promise((resolve) => {
      resolveUpgrade = () => resolve({ upgraded: true });
    }));
    const onUpgraded = vi.fn();

    render(<ExpertMarketUpgradePanel expertSlug="video-editor" onUpgraded={onUpgraded} />);

    const upgrade = await screen.findByRole("button", { name: "Upgrade expert" });
    fireEvent.click(upgrade);
    expect(upgrade).toBeDisabled();
    expect(screen.getByRole("button", { name: "Upgrading" })).toBeDisabled();
    resolveUpgrade?.();

    await waitFor(() => {
      expect(upgradeExpertFromMarket).toHaveBeenCalledTimes(1);
      expect(onUpgraded).toHaveBeenCalledTimes(1);
      expect(screen.getByText("Up to date")).toBeInTheDocument();
    });
  });

  it("keeps the upgrade available when the async expert refresh fails", async () => {
    vi.mocked(upgradeExpertFromMarket).mockResolvedValue({ upgraded: true });
    const onUpgraded = vi.fn().mockRejectedValue(new Error("Expert refresh failed"));

    render(
      <ExpertMarketUpgradePanel
        expertSlug="video-editor"
        initialAvailability
        onUpgraded={onUpgraded}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Upgrade expert" }));

    expect(await screen.findByText("Expert refresh failed")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Upgrade expert" })).toBeEnabled();
    expect(getExpertMarketUpgrade).not.toHaveBeenCalled();
  });

  it("shows a failed availability check and retries it", async () => {
    vi.mocked(getExpertMarketUpgrade)
      .mockRejectedValueOnce(new Error("Upgrade check failed"))
      .mockResolvedValueOnce({ upgrade_available: false });

    render(<ExpertMarketUpgradePanel expertSlug="video-editor" onUpgraded={vi.fn()} />);

    expect(await screen.findByText("Upgrade check failed")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("Up to date")).toBeInTheDocument();
  });
});
