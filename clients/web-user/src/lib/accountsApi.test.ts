import { beforeEach, describe, expect, it, vi } from "vitest";
import { getMe, listMyOrgSlug, login } from "./accountsApi";

const fetchMock = vi.fn();

vi.stubGlobal("fetch", fetchMock);

describe("accountsApi shared auth protocol", () => {
  beforeEach(() => {
    fetchMock.mockReset();
    localStorage.clear();
  });

  it("sends password login through AuthService and maps its response", async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          token: "token-1",
          refreshToken: "refresh-1",
          expiresIn: "3600",
          user: { id: "7", isSystemAdmin: false },
        }),
        { status: 200 },
      ),
    );

    await expect(login({ username: "dev", password: "password" })).resolves.toEqual({
      ok: true,
      token: "token-1",
      refresh_token: "refresh-1",
      expires_in: 3600,
      user: { id: "7", is_admin: false },
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/proto.auth.v1.AuthService/Login",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ username: "dev", password: "password" }),
      }),
    );
  });

  it("reads identity through UserService with the persisted bearer token", async () => {
    localStorage.setItem(
      "agent-cloud-auth/http_localhost_10000/session",
      JSON.stringify({
        access_token: "token-1",
        refresh_token: "refresh-1",
        expires_at: Math.floor(Date.now() / 1000) + 3600,
      }),
    );
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "7",
          isSystemAdmin: true,
          createdAt: "2026-07-12T00:00:00Z",
          lastLoginAt: "2026-07-12T01:00:00Z",
        }),
        { status: 200 },
      ),
    );

    await expect(getMe()).resolves.toEqual({
      id: "7",
      is_admin: true,
      created_at: 1783814400,
      last_login_at: 1783818000,
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/proto.user.v1.UserService/GetMe",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ Authorization: "Bearer token-1" }),
      }),
    );
  });

  it("loads the initial organization through OrgService", async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ items: [{ slug: "dev-org" }] }), { status: 200 }),
    );

    await expect(listMyOrgSlug("token-1")).resolves.toBe("dev-org");
    expect(fetchMock).toHaveBeenCalledWith(
      "/proto.org.v1.OrgService/ListMyOrgs",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ Authorization: "Bearer token-1" }),
      }),
    );
  });

  it("rejects malformed organization responses instead of persisting an empty selection", async () => {
    fetchMock.mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }));

    await expect(listMyOrgSlug("token-1")).rejects.toThrow("invalid response");
  });
});
