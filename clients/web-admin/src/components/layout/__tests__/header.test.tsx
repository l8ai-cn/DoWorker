import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

const mockPathname = vi.fn(() => "/");
vi.mock("next/navigation", () => ({
  usePathname: () => mockPathname(),
}));

import { Header } from "../header";

describe("Header", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPathname.mockReturnValue("/");
  });

  describe("page titles", () => {
    it("should show dashboard title for '/'", () => {
      mockPathname.mockReturnValue("/");
      render(<Header />);
      expect(screen.getByText("仪表盘")).toBeInTheDocument();
    });

    it("should show users title for '/users'", () => {
      mockPathname.mockReturnValue("/users");
      render(<Header />);
      expect(screen.getByText("用户")).toBeInTheDocument();
    });

    it("should show organizations title for '/organizations'", () => {
      mockPathname.mockReturnValue("/organizations");
      render(<Header />);
      expect(screen.getByText("组织")).toBeInTheDocument();
    });

    it("should show runners title for '/runners'", () => {
      mockPathname.mockReturnValue("/runners");
      render(<Header />);
      expect(screen.getByText("Runner")).toBeInTheDocument();
    });

    it("should show relays title for '/relays'", () => {
      mockPathname.mockReturnValue("/relays");
      render(<Header />);
      expect(screen.getByText("中继")).toBeInTheDocument();
    });

    it("should show skill registries title for '/skill-registries'", () => {
      mockPathname.mockReturnValue("/skill-registries");
      render(<Header />);
      expect(screen.getByText("技能源")).toBeInTheDocument();
    });

    it("should show promo codes title for '/promo-codes'", () => {
      mockPathname.mockReturnValue("/promo-codes");
      render(<Header />);
      expect(screen.getByText("优惠码")).toBeInTheDocument();
    });

    it("should show support tickets title for '/support-tickets'", () => {
      mockPathname.mockReturnValue("/support-tickets");
      render(<Header />);
      expect(screen.getByText("支持工单")).toBeInTheDocument();
    });

    it("should show audit logs title for '/audit-logs'", () => {
      mockPathname.mockReturnValue("/audit-logs");
      render(<Header />);
      expect(screen.getByText("审计日志")).toBeInTheDocument();
    });
  });

  describe("dynamic route titles", () => {
    it("should show user details for '/users/123'", () => {
      mockPathname.mockReturnValue("/users/123");
      render(<Header />);
      expect(screen.getByText("用户详情")).toBeInTheDocument();
    });

    it("should show organization details for '/organizations/5'", () => {
      mockPathname.mockReturnValue("/organizations/5");
      render(<Header />);
      expect(screen.getByText("组织详情")).toBeInTheDocument();
    });

    it("should show runner details for '/runners/10'", () => {
      mockPathname.mockReturnValue("/runners/10");
      render(<Header />);
      expect(screen.getByText("Runner 详情")).toBeInTheDocument();
    });

    it("should show relay details for '/relays/abc'", () => {
      mockPathname.mockReturnValue("/relays/abc");
      render(<Header />);
      expect(screen.getByText("中继详情")).toBeInTheDocument();
    });

    it("should show create promo code for '/promo-codes/new'", () => {
      mockPathname.mockReturnValue("/promo-codes/new");
      render(<Header />);
      expect(screen.getByText("创建优惠码")).toBeInTheDocument();
    });

    it("should show promo code details for '/promo-codes/5'", () => {
      mockPathname.mockReturnValue("/promo-codes/5");
      render(<Header />);
      expect(screen.getByText("优惠码详情")).toBeInTheDocument();
    });

    it("should show ticket details for '/support-tickets/7'", () => {
      mockPathname.mockReturnValue("/support-tickets/7");
      render(<Header />);
      expect(screen.getByText("工单详情")).toBeInTheDocument();
    });

    it("should show skill registry details for '/skill-registries/3'", () => {
      mockPathname.mockReturnValue("/skill-registries/3");
      render(<Header />);
      expect(screen.getByText("技能源详情")).toBeInTheDocument();
    });

    it("should fall back to admin console for unknown paths", () => {
      mockPathname.mockReturnValue("/unknown/deep/path");
      render(<Header />);
      expect(screen.getByText("管理控制台")).toBeInTheDocument();
    });
  });

  describe("hamburger menu", () => {
    it("should not render menu button when onMenuClick is not provided", () => {
      render(<Header />);
      expect(screen.queryByText("打开菜单")).not.toBeInTheDocument();
    });

    it("should render menu button when onMenuClick is provided", () => {
      render(<Header onMenuClick={() => {}} />);
      expect(screen.getByText("打开菜单")).toBeInTheDocument();
    });

    it("should call onMenuClick when menu button is clicked", () => {
      const handleMenuClick = vi.fn();
      render(<Header onMenuClick={handleMenuClick} />);
      fireEvent.click(screen.getByText("打开菜单").closest("button")!);
      expect(handleMenuClick).toHaveBeenCalledTimes(1);
    });
  });

  describe("notification bell", () => {
    it("should always render notification button", () => {
      render(<Header />);
      const buttons = screen.getAllByRole("button");
      expect(buttons.length).toBeGreaterThanOrEqual(1);
    });
  });
});
