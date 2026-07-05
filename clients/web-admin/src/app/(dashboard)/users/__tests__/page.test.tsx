import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

const mockToastSuccess = vi.fn();
const mockToastError = vi.fn();
vi.mock("sonner", () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}));

const mockListUsers = vi.fn();
const mockDisableUser = vi.fn();
const mockEnableUser = vi.fn();
const mockGrantAdmin = vi.fn();
const mockRevokeAdmin = vi.fn();
const mockVerifyUserEmail = vi.fn();
const mockUnverifyUserEmail = vi.fn();

vi.mock("@/lib/api/admin", () => ({
  listUsers: (...args: unknown[]) => mockListUsers(...args),
  disableUser: (...args: unknown[]) => mockDisableUser(...args),
  enableUser: (...args: unknown[]) => mockEnableUser(...args),
  grantAdmin: (...args: unknown[]) => mockGrantAdmin(...args),
  revokeAdmin: (...args: unknown[]) => mockRevokeAdmin(...args),
  verifyUserEmail: (...args: unknown[]) => mockVerifyUserEmail(...args),
  unverifyUserEmail: (...args: unknown[]) => mockUnverifyUserEmail(...args),
}));

import UsersPage from "../page";

const mockUsersResponse = {
  data: [
    {
      id: 1,
      email: "alice@test.com",
      username: "alice",
      name: "Alice Admin",
      avatar_url: null,
      is_active: true,
      is_system_admin: true,
      is_email_verified: true,
      last_login_at: "2024-06-15T10:00:00Z",
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-06-15T10:00:00Z",
    },
    {
      id: 2,
      email: "bob@test.com",
      username: "bob",
      name: null,
      avatar_url: "https://example.com/bob.png",
      is_active: false,
      is_system_admin: false,
      is_email_verified: false,
      last_login_at: null,
      created_at: "2024-02-01T00:00:00Z",
      updated_at: "2024-02-01T00:00:00Z",
    },
  ],
  total: 2,
  page: 1,
  page_size: 20,
  total_pages: 1,
};

describe("UsersPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListUsers.mockResolvedValue(mockUsersResponse);
    mockDisableUser.mockResolvedValue({ id: 1, is_active: false });
    mockEnableUser.mockResolvedValue({ id: 2, is_active: true });
    mockGrantAdmin.mockResolvedValue({ id: 2, is_system_admin: true });
    mockRevokeAdmin.mockResolvedValue({ id: 1, is_system_admin: false });
  });

  it("should render search input", async () => {
    render(<UsersPage />);
    expect(
      screen.getByPlaceholderText("搜索用户...")
    ).toBeInTheDocument();
  });

  it("should display user list after loading", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");
    expect(screen.getByText("Alice Admin")).toBeInTheDocument();
    expect(screen.getByText("alice@test.com")).toBeInTheDocument();
  });

  it("should display user initial for users without avatar", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");
    expect(screen.getByText("A")).toBeInTheDocument();
  });

  it("should display avatar image when available", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");
    const img = screen.getByRole("img");
    expect(img).toHaveAttribute("src", "https://example.com/bob.png");
  });

  it("should show username when name is null", async () => {
    render(<UsersPage />);
    await screen.findByText("bob");
    expect(screen.getByText("bob")).toBeInTheDocument();
  });

  it("should show Admin badge for admin users", async () => {
    render(<UsersPage />);
    await screen.findByText("管理员");
    expect(screen.getByText("管理员")).toBeInTheDocument();
  });

  it("should show Disabled badge for inactive users", async () => {
    render(<UsersPage />);
    await screen.findByText("已停用");
    expect(screen.getByText("已停用")).toBeInTheDocument();
  });

  it("should show Unverified badge for unverified users", async () => {
    render(<UsersPage />);
    await screen.findByText("未验证");
    expect(screen.getByText("未验证")).toBeInTheDocument();
  });

  it("should show total user count", async () => {
    render(<UsersPage />);
    await screen.findByText("用户 (2)");
    expect(screen.getByText("用户 (2)")).toBeInTheDocument();
  });

  it("should show empty state when no users found", async () => {
    mockListUsers.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      page_size: 20,
      total_pages: 0,
    });
    render(<UsersPage />);
    await screen.findByText("暂无用户");
    expect(screen.getByText("暂无用户")).toBeInTheDocument();
  });

  it("should show loading skeleton initially", () => {
    mockListUsers.mockReturnValue(new Promise(() => {}));
    render(<UsersPage />);
    const skeletons = document.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("should call listUsers with search param when searching", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    const searchInput = screen.getByPlaceholderText("搜索用户...");
    fireEvent.change(searchInput, { target: { value: "alice" } });

    await waitFor(() => {
      expect(mockListUsers).toHaveBeenCalledWith(
        expect.objectContaining({ search: "alice", page: 1 })
      );
    });
  });

  it("should disable user and show success toast", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    const disableBtn = screen.getByTitle("停用用户");
    fireEvent.click(disableBtn);

    await waitFor(() => {
      expect(mockDisableUser).toHaveBeenCalledWith(1);
      expect(mockToastSuccess).toHaveBeenCalledWith(
        "用户已停用"
      );
    });
  });

  it("should enable user and show success toast", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    const enableBtn = screen.getByTitle("启用用户");
    fireEvent.click(enableBtn);

    await waitFor(() => {
      expect(mockEnableUser).toHaveBeenCalledWith(2);
      expect(mockToastSuccess).toHaveBeenCalledWith(
        "用户已启用"
      );
    });
  });

  it("should grant admin and show success toast", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    const grantBtn = screen.getByTitle("授予管理员");
    fireEvent.click(grantBtn);

    await waitFor(() => {
      expect(mockGrantAdmin).toHaveBeenCalledWith(2);
      expect(mockToastSuccess).toHaveBeenCalledWith(
        "管理员权限已授予"
      );
    });
  });

  it("should revoke admin and show success toast", async () => {
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    const revokeBtn = screen.getByTitle("撤销管理员");
    fireEvent.click(revokeBtn);

    await waitFor(() => {
      expect(mockRevokeAdmin).toHaveBeenCalledWith(1);
      expect(mockToastSuccess).toHaveBeenCalledWith(
        "管理员权限已撤销"
      );
    });
  });

  it("should show error toast when action fails", async () => {
    mockDisableUser.mockRejectedValue({ error: "Permission denied" });
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    fireEvent.click(screen.getByTitle("停用用户"));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith("Permission denied");
    });
  });

  it("should show generic error when error has no message", async () => {
    mockDisableUser.mockRejectedValue({});
    render(<UsersPage />);
    await screen.findByText("Alice Admin");

    fireEvent.click(screen.getByTitle("停用用户"));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith("停用用户失败");
    });
  });
});
