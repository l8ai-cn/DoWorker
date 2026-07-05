import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

const mockPathname = vi.fn(() => "/");
vi.mock("next/navigation", () => ({
  usePathname: () => mockPathname(),
}));

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    onClick,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    onClick?: () => void;
    className?: string;
  }) => (
    <a href={href} onClick={onClick} {...props}>
      {children}
    </a>
  ),
}));

const mockLogout = vi.fn();
const mockUser = {
  id: 1,
  email: "admin@test.com",
  username: "admin",
  name: "Admin User",
  avatar_url: null as string | null,
  is_system_admin: true,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({
    user: mockUser,
    logout: mockLogout,
  }),
}));

vi.mock("@/components/ui/sheet", () => ({
  Sheet: ({
    children,
    open,
  }: {
    children: React.ReactNode;
    open: boolean;
  }) => (open ? <div data-testid="sheet">{children}</div> : null),
  SheetContent: ({
    children,
  }: {
    children: React.ReactNode;
    side?: string;
    className?: string;
  }) => <div data-testid="sheet-content">{children}</div>,
  SheetTitle: ({
    children,
    className,
  }: {
    children: React.ReactNode;
    className?: string;
  }) => (
    <span data-testid="sheet-title" className={className}>
      {children}
    </span>
  ),
}));

import { SidebarContent, Sidebar, MobileSidebar } from "../sidebar";

describe("SidebarContent", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPathname.mockReturnValue("/");
    mockUser.name = "Admin User";
    mockUser.avatar_url = null;
  });

  it("should render the Do Worker admin brand", () => {
    render(<SidebarContent />);
    expect(screen.getByText("Do Worker")).toBeInTheDocument();
    expect(screen.getByText("管理控制台")).toBeInTheDocument();
  });

  it("should render all navigation items", () => {
    render(<SidebarContent />);
    const expectedItems = [
      "仪表盘",
      "用户",
      "组织",
      "单点登录",
      "Runner",
      "中继",
      "技能源",
      "优惠码",
      "支持工单",
      "审计日志",
    ];
    expectedItems.forEach((item) => {
      expect(screen.getByText(item)).toBeInTheDocument();
    });
  });

  it("should highlight active nav item based on pathname", () => {
    mockPathname.mockReturnValue("/users");
    render(<SidebarContent />);
    const usersLink = screen.getByText("用户").closest("a");
    expect(usersLink?.className).toContain("bg-primary/10");
  });

  it("should highlight nested routes (e.g. /organizations/5)", () => {
    mockPathname.mockReturnValue("/organizations/5");
    render(<SidebarContent />);
    const orgLink = screen.getByText("组织").closest("a");
    expect(orgLink?.className).toContain("bg-primary/10");
  });

  it("should not highlight non-matching items", () => {
    mockPathname.mockReturnValue("/users");
    render(<SidebarContent />);
    const runnersLink = screen.getByText("Runner").closest("a");
    expect(runnersLink?.className).not.toContain("bg-primary/10");
  });

  it("should display user info with initial when no avatar", () => {
    render(<SidebarContent />);
    expect(screen.getByText("A")).toBeInTheDocument();
    expect(screen.getByText("Admin User")).toBeInTheDocument();
    expect(screen.getByText("admin@test.com")).toBeInTheDocument();
  });

  it("should display avatar image when avatar_url is set", () => {
    mockUser.avatar_url = "https://example.com/avatar.png";
    render(<SidebarContent />);
    const img = screen.getByRole("img");
    expect(img).toHaveAttribute("src", "https://example.com/avatar.png");
  });

  it("should display username when name is null", () => {
    mockUser.name = null as unknown as string;
    render(<SidebarContent />);
    // Should fall back to username display
    expect(screen.getByText("admin")).toBeInTheDocument();
  });

  it("should render Sign Out button", () => {
    render(<SidebarContent />);
    expect(screen.getByText("退出登录")).toBeInTheDocument();
  });

  it("should call logout when Sign Out is clicked", () => {
    render(<SidebarContent />);
    fireEvent.click(screen.getByText("退出登录"));
    expect(mockLogout).toHaveBeenCalledTimes(1);
  });

  it("should call onNavigate when a nav link is clicked", () => {
    const handleNavigate = vi.fn();
    render(<SidebarContent onNavigate={handleNavigate} />);
    fireEvent.click(screen.getByText("用户"));
    expect(handleNavigate).toHaveBeenCalledTimes(1);
  });

  it("should not throw when onNavigate is not provided", () => {
    render(<SidebarContent />);
    expect(() => fireEvent.click(screen.getByText("用户"))).not.toThrow();
  });
});

describe("Sidebar", () => {
  beforeEach(() => {
    mockPathname.mockReturnValue("/");
  });

  it("should render as an aside element", () => {
    render(<Sidebar />);
    const aside = document.querySelector("aside");
    expect(aside).toBeInTheDocument();
  });

  it("should have hidden md:flex classes for responsive behavior", () => {
    render(<Sidebar />);
    const aside = document.querySelector("aside");
    expect(aside?.className).toContain("hidden");
    expect(aside?.className).toContain("md:flex");
  });

  it("should have w-64 width", () => {
    render(<Sidebar />);
    const aside = document.querySelector("aside");
    expect(aside?.className).toContain("w-64");
  });
});

describe("MobileSidebar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPathname.mockReturnValue("/");
  });

  it("should not render Sheet content when closed", () => {
    render(<MobileSidebar open={false} onOpenChange={() => {}} />);
    expect(screen.queryByTestId("sheet")).not.toBeInTheDocument();
  });

  it("should render Sheet content when open", () => {
    render(<MobileSidebar open={true} onOpenChange={() => {}} />);
    expect(screen.getByTestId("sheet")).toBeInTheDocument();
  });

  it("should render navigation content inside Sheet when open", () => {
    render(<MobileSidebar open={true} onOpenChange={() => {}} />);
    expect(screen.getByText("仪表盘")).toBeInTheDocument();
    expect(screen.getByText("用户")).toBeInTheDocument();
  });

  it("should have sr-only Navigation title for accessibility", () => {
    render(<MobileSidebar open={true} onOpenChange={() => {}} />);
    const title = screen.getByTestId("sheet-title");
    expect(title).toHaveTextContent("导航");
    expect(title?.className).toContain("sr-only");
  });

  it("should call onOpenChange(false) when a nav link is clicked", () => {
    const handleOpenChange = vi.fn();
    render(<MobileSidebar open={true} onOpenChange={handleOpenChange} />);
    fireEvent.click(screen.getByText("用户"));
    expect(handleOpenChange).toHaveBeenCalledWith(false);
  });
});
