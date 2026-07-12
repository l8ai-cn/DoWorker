import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@/test/test-utils";

import { ApplicationsPage } from "./ApplicationsPage";

const fetchExperts = vi.fn();
let applications: unknown[] = [];
let applicationError: Error | null = null;
let experts = [{ id: 12, slug: "delivery-agent" }];

vi.mock("@/stores/auth", () => ({
  useCurrentOrg: () => ({ id: 9, slug: "dev-org", name: "研发组织" }),
}));

vi.mock("@/stores/expert", () => ({
  useExperts: () => experts,
  useExpertStore: (selector: (state: {
    error: string | null;
    fetchExperts: () => Promise<void>;
  }) => unknown) => selector({ error: null, fetchExperts }),
}));

vi.mock("@/lib/marketplace/application-api", () => ({
  fetchOrganizationApplications: () => (
    applicationError ? Promise.reject(applicationError) : Promise.resolve(applications)
  ),
  expertIDFromRuntimeRef: (runtimeRef: string) => {
    const match = /^expert:(\d+)$/.exec(runtimeRef);
    return match ? Number(match[1]) : null;
  },
}));

describe("ApplicationsPage", () => {
  beforeEach(() => {
    applications = [];
    applicationError = null;
    experts = [{ id: 12, slug: "delivery-agent" }];
    fetchExperts.mockResolvedValue(undefined);
  });

  it("uses the market application result and starts the mapped expert", async () => {
    applications = [{
      installation_id: "installation-1",
      market_slug: "do-worker-market",
      listing_slug: "software-delivery-expert",
      display_name: "软件交付专家",
      tagline: "把需求变成可验证的代码交付",
      resource_type: "application",
      outcomes: ["执行关键路径验证"],
      runtime_ref: "expert:12",
      status: "active",
      installed_at: "2026-07-12T08:00:00Z",
    }];

    render(<ApplicationsPage orgSlug="dev-org" />);

    expect(await screen.findByText("软件交付专家")).toBeInTheDocument();
    expect(screen.getByText("把需求变成可验证的代码交付")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "开始第一个任务" }))
      .toHaveAttribute("href", "/dev-org/experts/delivery-agent");
  });

  it("does not fabricate a runtime action when the app has no runtime reference", async () => {
    applications = [{
      installation_id: "installation-2",
      market_slug: "do-worker-market",
      listing_slug: "connector",
      display_name: "仓库连接器",
      tagline: "把仓库授权接入工作流",
      resource_type: "mcp_connector",
      outcomes: [],
      runtime_ref: "",
      status: "verifying",
      installed_at: "2026-07-12T08:00:00Z",
    }];

    render(<ApplicationsPage orgSlug="dev-org" />);

    expect(await screen.findByText("仓库连接器")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "开始第一个任务" })).not.toBeInTheDocument();
    expect(screen.getByText("查看启用详情")).toBeInTheDocument();
  });

  it("explains when the organization has not enabled an app", async () => {
    render(<ApplicationsPage orgSlug="dev-org" />);

    expect(await screen.findByText("还没有已启用的应用")).toBeInTheDocument();
  });

  it("shows a recovery message when the application API cannot be read", async () => {
    applicationError = new Error("没有查看此组织应用的权限");

    render(<ApplicationsPage orgSlug="dev-org" />);

    await waitFor(() => {
      expect(screen.getByText("应用中心暂时无法加载")).toBeInTheDocument();
    });
    expect(screen.getByText("没有查看此组织应用的权限")).toBeInTheDocument();
  });
});
