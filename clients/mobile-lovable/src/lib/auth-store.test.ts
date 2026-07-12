import { beforeEach, describe, expect, it, vi } from "vitest";
import { login, readAuthToken, readOrgSlug, restoreAuthIdentity } from "./auth-store";
import {
  getMobileAuthManager,
  mobileAuthSessionStorageKey,
  mobileAuthUrlSlug,
} from "./mobile-auth-manager";

vi.mock("./mobile-auth-manager", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./mobile-auth-manager")>()),
  getMobileAuthManager: vi.fn(),
}));

const manager = {
  login: vi.fn(),
  bootstrap: vi.fn(),
  fetch_organizations: vi.fn(),
  get_current_org_json: vi.fn(),
  clear_session: vi.fn(),
};

describe("mobile auth store", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(getMobileAuthManager).mockResolvedValue(manager as never);
    manager.login.mockReset();
    manager.bootstrap.mockReset();
    manager.fetch_organizations.mockReset();
    manager.get_current_org_json.mockReset();
    manager.clear_session.mockReset();
  });

  it("reads the shared AuthManager session format and rejects expired tokens", () => {
    localStorage.setItem(
      mobileAuthSessionStorageKey(),
      JSON.stringify({
        access_token: "valid-token",
        expires_at: Math.floor(Date.now() / 1000) + 60,
        current_org_slug: "dev-org",
      }),
    );

    expect(readAuthToken()).toBe("valid-token");
    expect(readOrgSlug()).toBe("dev-org");

    localStorage.setItem(
      mobileAuthSessionStorageKey(),
      JSON.stringify({ access_token: "expired-token", expires_at: 1 }),
    );
    expect(readAuthToken()).toBeNull();
  });

  it("uses the Rust URL slug rules for API URLs with paths and trailing slashes", () => {
    expect(mobileAuthUrlSlug("https://API.AgentsMesh.AI/api/")).toBe("https_api_agentsmesh_ai");
    expect(mobileAuthUrlSlug("http://localhost:10015/")).toBe("http_localhost_10015");
  });

  it("uses the Rust AuthManager login and organization selection flow", async () => {
    manager.login.mockResolvedValue(
      JSON.stringify({
        token: "token-1",
        refresh_token: "refresh-1",
        expires_in: 3600,
        user: { id: 7, email: "dev@agentsmesh.local" },
      }),
    );
    manager.fetch_organizations.mockResolvedValue("[]");
    manager.get_current_org_json.mockReturnValue(JSON.stringify({ slug: "dev-org" }));

    await expect(login("dev@agentsmesh.local", "password")).resolves.toEqual({
      token: "token-1",
      expiresIn: 3600,
      orgSlug: "dev-org",
      userId: "7",
      email: "dev@agentsmesh.local",
    });
    expect(manager.login).toHaveBeenCalledWith("dev@agentsmesh.local", "password");
    expect(manager.fetch_organizations).toHaveBeenCalledOnce();
  });

  it("clears the persisted session when organization initialization fails", async () => {
    manager.login.mockResolvedValue(
      JSON.stringify({
        token: "token-1",
        expires_in: 3600,
        user: { id: 7, email: "dev@agentsmesh.local" },
      }),
    );
    manager.fetch_organizations.mockRejectedValue(new Error("organization service unavailable"));

    await expect(login("dev@agentsmesh.local", "password")).rejects.toThrow(
      "organization service unavailable",
    );
    expect(manager.clear_session).toHaveBeenCalledOnce();
  });

  it("restores the user and organization from the Rust AuthManager bootstrap result", async () => {
    manager.bootstrap.mockResolvedValue(
      JSON.stringify({
        kind: "authenticated",
        user: { email: "dev@agentsmesh.local" },
        current_org: { slug: "dev-org" },
      }),
    );

    await expect(restoreAuthIdentity()).resolves.toEqual({
      authenticated: true,
      email: "dev@agentsmesh.local",
      orgSlug: "dev-org",
    });
  });
});
