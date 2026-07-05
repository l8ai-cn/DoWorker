import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

const mockReplace = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const mockToastSuccess = vi.fn();
const mockToastError = vi.fn();
vi.mock("sonner", () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}));

const mockSetAuth = vi.fn();
const mockAuthState = { token: null as string | null, setAuth: mockSetAuth };
vi.mock("@/stores/auth", () => ({
  useAuthStore: () => mockAuthState,
}));

const mockLogin = vi.fn();
vi.mock("@/lib/api/admin", () => ({
  login: (req: unknown) => mockLogin(req),
}));

import LoginPage from "../page";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockAuthState.token = null;
  });

  it("should render login form", () => {
    render(<LoginPage />);
    expect(screen.getByText("管理控制台")).toBeInTheDocument();
    expect(screen.getByLabelText("用户名")).toBeInTheDocument();
    expect(screen.getByLabelText("密码")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "登录" })).toBeInTheDocument();
  });

  it("should render description text", () => {
    render(<LoginPage />);
    expect(
      screen.getByText(/使用管理员账号登录/)
    ).toBeInTheDocument();
  });

  it("should render admin-only notice", () => {
    render(<LoginPage />);
    expect(
      screen.getByText(/仅系统管理员可以访问此控制台/)
    ).toBeInTheDocument();
  });

  it("should redirect to / if already authenticated", async () => {
    mockAuthState.token = "existing-token";
    render(<LoginPage />);
    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith("/");
    });
  });

  it("should show error toast when submitting empty form", async () => {
    render(<LoginPage />);
    fireEvent.click(screen.getByRole("button", { name: "登录" }));
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith("请输入用户名和密码");
    });
  });

  it("should call login API on valid submit", async () => {
    mockLogin.mockResolvedValue({
      token: "new-token",
      refresh_token: "new-refresh",
      user: {
        id: 1,
        email: "admin@test.com",
        username: "admin",
        name: "Admin",
        avatar_url: null,
        is_system_admin: true,
      },
    });

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText("用户名"), {
      target: { value: "admin@test.com" },
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: "admin@test.com",
        password: "password123",
      });
    });
  });

  it("should call setAuth and redirect on successful login", async () => {
    const mockUser = {
      id: 1,
      email: "admin@test.com",
      username: "admin",
      name: "Admin",
      avatar_url: null,
      is_system_admin: true,
    };
    mockLogin.mockResolvedValue({
      token: "new-token",
      refresh_token: "new-refresh",
      user: mockUser,
    });

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText("用户名"), {
      target: { value: "admin@test.com" },
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "pass" },
    });
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => {
      expect(mockSetAuth).toHaveBeenCalledWith(
        "new-token",
        "new-refresh",
        mockUser
      );
      expect(mockToastSuccess).toHaveBeenCalledWith("欢迎回来，Admin");
      expect(mockReplace).toHaveBeenCalledWith("/");
    });
  });

  it("should show error toast on login failure", async () => {
    mockLogin.mockRejectedValue({ error: "Invalid credentials" });

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText("用户名"), {
      target: { value: "admin@test.com" },
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "wrong" },
    });
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith("Invalid credentials");
    });
  });

  it("should show generic error when error object has no message", async () => {
    mockLogin.mockRejectedValue({});

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText("用户名"), {
      target: { value: "a@b.com" },
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "x" },
    });
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(
        "登录失败，请检查账号和密码。"
      );
    });
  });

  it("should show loading state during submission", async () => {
    mockLogin.mockReturnValue(new Promise(() => {}));

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText("用户名"), {
      target: { value: "a@b.com" },
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "x" },
    });
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => {
      expect(screen.getByText("登录中...")).toBeInTheDocument();
    });
  });
});
