import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useLightSession } from "@/hooks/useLightSession";
import { installMarketplaceApplication } from "@/lib/marketplace-install";
import { listMarketplaceModelResources } from "@/lib/marketplace-model-resources";
import { MarketplaceInstallButton } from "./MarketplaceInstallButton";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));
vi.mock("@/hooks/useLightSession", () => ({
  useLightSession: vi.fn(),
}));
vi.mock("@/lib/marketplace-install", () => ({
  installMarketplaceApplication: vi.fn(),
}));
vi.mock("@/lib/marketplace-model-resources", () => ({
  listMarketplaceModelResources: vi.fn(),
}));
vi.mock("@/lib/light-session", () => ({
  updateLightSessionOrgSlug: vi.fn(),
}));

describe("MarketplaceInstallButton", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useLightSession).mockReturnValue({
      hydrated: true,
      session: {
        currentOrgSlug: "acme",
        isAuthenticated: true,
        expiresAt: Date.now() + 60_000,
      },
    });
    vi.mocked(listMarketplaceModelResources).mockResolvedValue([
      { id: 42, label: "OpenAI · GPT 5.5" },
    ]);
    vi.mocked(installMarketplaceApplication).mockResolvedValue({
      expert: { slug: "video-production-expert" },
      already_installed: false,
    });
  });

  it("requires an explicit compatible model selection before install", async () => {
    render(
      <MarketplaceInstallButton
        applicationSlug="video-production-expert"
        agentSlug="codex-cli"
      />,
    );

    const install = await screen.findByRole("button", { name: "立即启用" });
    expect(install).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "选择运行模型" }));
    fireEvent.click(screen.getByRole("option", { name: "OpenAI · GPT 5.5" }));
    expect(install).toBeEnabled();
    fireEvent.click(install);

    await waitFor(() => {
      expect(installMarketplaceApplication).toHaveBeenCalledWith(
        "acme",
        "video-production-expert",
        42,
      );
      expect(screen.getByRole("button", { name: "立即启用" })).toBeEnabled();
    });
  });
});
