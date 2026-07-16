import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useLightSession } from "@/hooks/useLightSession";
import { discoverFirstOrgSlug } from "@/lib/light-auth";
import { updateLightSessionOrgSlug } from "@/lib/light-session";
import { installMarketplaceApplication } from "@/lib/marketplace-install";
import { listMarketplaceModelResources } from "@/lib/marketplace-model-resources";
import { listMarketplaceToolModelResources } from "@/lib/marketplace-tool-model-resources";
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
vi.mock("@/lib/light-auth", () => ({
  discoverFirstOrgSlug: vi.fn(),
}));
vi.mock("@/lib/marketplace-model-resources", () => ({
  listMarketplaceModelResources: vi.fn(),
}));
vi.mock("@/lib/marketplace-tool-model-resources", () => ({
  listMarketplaceToolModelResources: vi.fn(),
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
    vi.mocked(listMarketplaceToolModelResources).mockResolvedValue([]);
    vi.mocked(discoverFirstOrgSlug).mockResolvedValue({ status: "empty" });
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
        {},
      );
      expect(screen.getByRole("button", { name: "立即启用" })).toBeEnabled();
    });
  });

  it("requires every canonical tool model before install", async () => {
    vi.mocked(listMarketplaceToolModelResources).mockResolvedValue([
      {
        role: "seedance-video",
        resources: [{ id: 77, label: "Doubao · Seedance 2.0" }],
      },
    ]);
    render(
      <MarketplaceInstallButton
        applicationSlug="seedance-director"
        agentSlug="seedance-expert"
      />,
    );

    const install = await screen.findByRole("button", { name: "立即启用" });
    fireEvent.click(screen.getByRole("button", { name: "选择运行模型" }));
    fireEvent.click(screen.getByRole("option", { name: "OpenAI · GPT 5.5" }));
    expect(install).toBeDisabled();
    fireEvent.change(screen.getByLabelText("选择视频生成模型"), {
      target: { value: "77" },
    });
    expect(install).toBeEnabled();
    fireEvent.click(install);

    await waitFor(() => {
      expect(installMarketplaceApplication).toHaveBeenCalledWith(
        "acme",
        "seedance-director",
        42,
        { "seedance-video": 77 },
      );
    });
  });

  it("keeps the resolved organization when model loading fails", async () => {
    vi.mocked(useLightSession).mockReturnValue({
      hydrated: true,
      session: {
        currentOrgSlug: null,
        isAuthenticated: true,
        expiresAt: Date.now() + 60_000,
      },
    });
    vi.mocked(discoverFirstOrgSlug).mockResolvedValue({
      status: "found",
      slug: "acme",
    });
    vi.mocked(listMarketplaceModelResources).mockRejectedValue(
      new Error("model API unavailable"),
    );

    render(
      <MarketplaceInstallButton
        applicationSlug="video-production-expert"
        agentSlug="codex-cli"
      />,
    );

    expect(
      await screen.findByRole("button", { name: "重新加载模型" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "创建组织后启用" }),
    ).not.toBeInTheDocument();
  });

  it("shows retry when organization discovery is unavailable", async () => {
    vi.mocked(useLightSession).mockReturnValue({
      hydrated: true,
      session: {
        currentOrgSlug: null,
        isAuthenticated: true,
        expiresAt: Date.now() + 60_000,
      },
    });
    vi.mocked(discoverFirstOrgSlug).mockResolvedValue({
      status: "unavailable",
    });

    render(
      <MarketplaceInstallButton
        applicationSlug="video-production-expert"
        agentSlug="codex-cli"
      />,
    );

    expect(
      await screen.findByRole("button", { name: "重新加载模型" }),
    ).toBeInTheDocument();
  });

  it("does not persist an organization from a cancelled discovery", async () => {
    let resolveDiscovery:
      | ((result: { status: "found"; slug: string }) => void)
      | undefined;
    vi.mocked(discoverFirstOrgSlug).mockReturnValue(
      new Promise((resolve) => {
        resolveDiscovery = resolve;
      }),
    );
    vi.mocked(useLightSession).mockReturnValue({
      hydrated: true,
      session: {
        currentOrgSlug: null,
        isAuthenticated: true,
        expiresAt: Date.now() + 60_000,
      },
    });
    const view = render(
      <MarketplaceInstallButton
        applicationSlug="video-production-expert"
        agentSlug="codex-cli"
      />,
    );
    vi.mocked(useLightSession).mockReturnValue({
      hydrated: true,
      session: {
        currentOrgSlug: "beta",
        isAuthenticated: true,
        expiresAt: Date.now() + 60_000,
      },
    });
    view.rerender(
      <MarketplaceInstallButton
        applicationSlug="video-production-expert"
        agentSlug="codex-cli"
      />,
    );
    resolveDiscovery?.({ status: "found", slug: "acme" });

    await screen.findByRole("button", { name: "立即启用" });
    expect(updateLightSessionOrgSlug).not.toHaveBeenCalledWith("acme");
  });
});
