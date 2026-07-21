import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  clearAgentCloudSession,
  isSessionLoggedOut,
  markSessionLoggedOut,
  persistAgentCloudSession,
  readAgentCloudJWT,
} from "./auth-session";

const SESSION_KEY = "agent-cloud-auth/http_localhost_10000/session";

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
});

afterEach(() => {
  localStorage.clear();
  sessionStorage.clear();
});

describe("session logout mark", () => {
  it("blocks readAgentCloudJWT after markSessionLoggedOut even with stored token", () => {
    persistAgentCloudSession({ accessToken: "tok", expiresIn: 3600 });
    expect(readAgentCloudJWT()).toBe("tok");
    markSessionLoggedOut();
    expect(readAgentCloudJWT()).toBeNull();
    expect(isSessionLoggedOut()).toBe(true);
  });

  it("clears logged-out mark on persistAgentCloudSession after login", () => {
    markSessionLoggedOut();
    persistAgentCloudSession({ accessToken: "new-tok", expiresIn: 3600 });
    expect(isSessionLoggedOut()).toBe(false);
    expect(readAgentCloudJWT()).toBe("new-tok");
  });

  it("clearAgentCloudSession keeps logged-out mark set", () => {
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ access_token: "tok", expires_at: Math.floor(Date.now() / 1000) + 3600 }),
    );
    markSessionLoggedOut();
    clearAgentCloudSession();
    expect(localStorage.getItem(SESSION_KEY)).toBeNull();
    expect(isSessionLoggedOut()).toBe(true);
    expect(readAgentCloudJWT()).toBeNull();
  });
});

describe("expiry semantics (SSOT alignment with web/Rust)", () => {
  it("treats a blob missing expires_at as expired -> null", () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify({ access_token: "tok" }));
    expect(readAgentCloudJWT()).toBeNull();
  });

  it("treats an expired expires_at as null", () => {
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ access_token: "tok", expires_at: Math.floor(Date.now() / 1000) - 10 }),
    );
    expect(readAgentCloudJWT()).toBeNull();
  });

  it("returns the token when expires_at is in the future", () => {
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ access_token: "tok", expires_at: Math.floor(Date.now() / 1000) + 3600 }),
    );
    expect(readAgentCloudJWT()).toBe("tok");
  });
});
