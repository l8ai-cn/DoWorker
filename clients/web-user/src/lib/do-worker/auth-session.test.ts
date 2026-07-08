import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  clearDoWorkerSession,
  isSessionLoggedOut,
  markSessionLoggedOut,
  persistDoWorkerSession,
  readDoWorkerJWT,
} from "./auth-session";

const SESSION_KEY = "do-worker-auth/http_localhost_10000/session";

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
});

afterEach(() => {
  localStorage.clear();
  sessionStorage.clear();
});

describe("session logout mark", () => {
  it("blocks readDoWorkerJWT after markSessionLoggedOut even with stored token", () => {
    persistDoWorkerSession({ accessToken: "tok", expiresIn: 3600 });
    expect(readDoWorkerJWT()).toBe("tok");
    markSessionLoggedOut();
    expect(readDoWorkerJWT()).toBeNull();
    expect(isSessionLoggedOut()).toBe(true);
  });

  it("clears logged-out mark on persistDoWorkerSession after login", () => {
    markSessionLoggedOut();
    persistDoWorkerSession({ accessToken: "new-tok", expiresIn: 3600 });
    expect(isSessionLoggedOut()).toBe(false);
    expect(readDoWorkerJWT()).toBe("new-tok");
  });

  it("clearDoWorkerSession keeps logged-out mark set", () => {
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ access_token: "tok", expires_at: Math.floor(Date.now() / 1000) + 3600 }),
    );
    markSessionLoggedOut();
    clearDoWorkerSession();
    expect(localStorage.getItem(SESSION_KEY)).toBeNull();
    expect(isSessionLoggedOut()).toBe(true);
    expect(readDoWorkerJWT()).toBeNull();
  });
});

describe("expiry semantics (SSOT alignment with web/Rust)", () => {
  it("treats a blob missing expires_at as expired -> null", () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify({ access_token: "tok" }));
    expect(readDoWorkerJWT()).toBeNull();
  });

  it("treats an expired expires_at as null", () => {
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ access_token: "tok", expires_at: Math.floor(Date.now() / 1000) - 10 }),
    );
    expect(readDoWorkerJWT()).toBeNull();
  });

  it("returns the token when expires_at is in the future", () => {
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ access_token: "tok", expires_at: Math.floor(Date.now() / 1000) + 3600 }),
    );
    expect(readDoWorkerJWT()).toBe("tok");
  });
});
