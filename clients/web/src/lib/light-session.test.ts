import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { urlSlug, sessionStorageKey, readLightSession } from "./light-session";

describe("urlSlug", () => {
  it("normalizes scheme + host to lowercase + alphanumeric_underscore", () => {
    expect(urlSlug("https://AGENTSMESH.AI")).toBe("https_agentsmesh_ai");
  });

  it("includes port when present", () => {
    expect(urlSlug("http://localhost:10007")).toBe("http_localhost_10007");
  });

  it("strips trailing path / query / hash", () => {
    expect(urlSlug("https://agentsmesh.ai/path?q=1#h")).toBe("https_agentsmesh_ai");
  });

  it("caps at 64 chars", () => {
    const long = "https://" + "a".repeat(200) + ".com";
    expect(urlSlug(long).length).toBeLessThanOrEqual(64);
  });

  it("falls back gracefully on garbage input", () => {
    const slug = urlSlug("not a url");
    expect(slug).toMatch(/^[a-z0-9_]+$/);
  });
});

describe("sessionStorageKey", () => {
  it("formats as agentsmesh-auth/<slug>/session", () => {
    expect(sessionStorageKey("https://agentsmesh.ai"))
      .toBe("agentsmesh-auth/https_agentsmesh_ai/session");
  });
});

describe("readLightSession", () => {
  const ORIGIN = "http://localhost:10007";
  const KEY = sessionStorageKey(ORIGIN);

  beforeEach(() => {
    window.localStorage.clear();
  });

  afterEach(() => {
    window.localStorage.clear();
  });

  it("returns null when no session is stored", () => {
    expect(readLightSession(ORIGIN)).toBeNull();
  });

  it("parses a fresh session as authenticated", () => {
    const future = Math.floor(Date.now() / 1000) + 3600;
    window.localStorage.setItem(
      KEY,
      JSON.stringify({
        access_token: "tok",
        refresh_token: "r",
        expires_at: future,
        base_url: ORIGIN,
        user_id: 42,
        current_org_slug: "dev-org",
        schema_version: 1,
      }),
    );
    const s = readLightSession(ORIGIN);
    expect(s).toEqual({
      userId: 42,
      currentOrgSlug: "dev-org",
      expiresAt: future,
      isAuthenticated: true,
    });
  });

  it("treats expired session as unauthenticated but still returns user_id", () => {
    const past = Math.floor(Date.now() / 1000) - 60;
    window.localStorage.setItem(
      KEY,
      JSON.stringify({
        access_token: "tok",
        expires_at: past,
        user_id: 42,
        current_org_slug: "dev-org",
      }),
    );
    const s = readLightSession(ORIGIN);
    expect(s?.isAuthenticated).toBe(false);
    expect(s?.userId).toBe(42);
  });

  it("treats missing access_token as unauthenticated", () => {
    window.localStorage.setItem(
      KEY,
      JSON.stringify({ expires_at: 9_999_999_999, user_id: 42 }),
    );
    const s = readLightSession(ORIGIN);
    expect(s?.isAuthenticated).toBe(false);
  });

  it("returns null on corrupt JSON", () => {
    window.localStorage.setItem(KEY, "{not json");
    expect(readLightSession(ORIGIN)).toBeNull();
  });

  it("isolates sessions across base_url namespaces", () => {
    const dev = "http://localhost:10007";
    const prod = "https://agentsmesh.ai";
    window.localStorage.setItem(
      sessionStorageKey(dev),
      JSON.stringify({ access_token: "dev", expires_at: 9_999_999_999, user_id: 1, current_org_slug: "d" }),
    );
    window.localStorage.setItem(
      sessionStorageKey(prod),
      JSON.stringify({ access_token: "prod", expires_at: 9_999_999_999, user_id: 2, current_org_slug: "p" }),
    );
    expect(readLightSession(dev)?.userId).toBe(1);
    expect(readLightSession(prod)?.userId).toBe(2);
  });
});
