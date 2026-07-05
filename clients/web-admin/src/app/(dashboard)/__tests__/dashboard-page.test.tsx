import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

const mockGetDashboardStats = vi.fn();
vi.mock("@/lib/api/admin", () => ({
  getDashboardStats: () => mockGetDashboardStats(),
}));

import DashboardPage from "../page";

const mockStats = {
  total_users: 1200,
  active_users: 950,
  total_organizations: 85,
  total_runners: 42,
  online_runners: 38,
  total_pods: 250,
  active_pods: 120,
  total_subscriptions: 60,
  active_subscriptions: 45,
  new_users_today: 8,
  new_users_this_week: 35,
  new_users_this_month: 150,
};

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetDashboardStats.mockResolvedValue(mockStats);
  });

  it("should show loading skeleton initially", () => {
    mockGetDashboardStats.mockReturnValue(new Promise(() => {}));
    render(<DashboardPage />);
    const skeletons = document.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("should display stats after loading", async () => {
    render(<DashboardPage />);
    await screen.findByText("1,200");
    expect(screen.getByText("1,200")).toBeInTheDocument();
    expect(screen.getByText("85")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByText("120")).toBeInTheDocument();
  });

  it("should display stat card titles", async () => {
    render(<DashboardPage />);
    await screen.findByText("用户总数");
    expect(screen.getByText("用户总数")).toBeInTheDocument();
    expect(screen.getByText("组织")).toBeInTheDocument();
    expect(screen.getByText("Runner")).toBeInTheDocument();
    expect(screen.getByText("活跃 Pod")).toBeInTheDocument();
  });

  it("should display sub-values", async () => {
    render(<DashboardPage />);
    await screen.findByText("950 个活跃");
    expect(screen.getByText("950 个活跃")).toBeInTheDocument();
    expect(screen.getByText("38 个在线")).toBeInTheDocument();
    expect(screen.getByText("共 250 个")).toBeInTheDocument();
  });

  it("should display new users breakdown", async () => {
    render(<DashboardPage />);
    await screen.findByText("新增用户");
    expect(screen.getByText("8")).toBeInTheDocument();
    expect(screen.getByText("35")).toBeInTheDocument();
  });

  it("should display subscriptions section", async () => {
    render(<DashboardPage />);
    await screen.findByText("订阅");
    expect(screen.getByText("45")).toBeInTheDocument();
    expect(screen.getByText("60")).toBeInTheDocument();
  });

  it("should display system health section", async () => {
    render(<DashboardPage />);
    await screen.findByText("系统健康");
    expect(screen.getByText("所有系统运行正常")).toBeInTheDocument();
    expect(
      screen.getByText("42 个 Runner 中有 38 个在线")
    ).toBeInTheDocument();
  });

  it("should show error state when API fails", async () => {
    mockGetDashboardStats.mockRejectedValue(new Error("Network error"));
    render(<DashboardPage />);
    await screen.findByText("仪表盘数据加载失败");
    expect(
      screen.getByText("仪表盘数据加载失败")
    ).toBeInTheDocument();
  });
});
