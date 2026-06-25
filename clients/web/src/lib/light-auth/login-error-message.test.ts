import { describe, it, expect, vi } from "vitest";
import { ApiError } from "@/lib/api/api-types";
import { resolveLoginErrorMessage } from "./login-error-message";

const t = vi.fn((key: string) => key);

describe("resolveLoginErrorMessage", () => {
  it("maps 401 to invalid credentials", () => {
    const msg = resolveLoginErrorMessage(new ApiError(401, "Unauthorized", { code: "UNAUTHENTICATED" }), t);
    expect(msg).toBe("auth.loginPage.invalidCredentials");
  });

  it("maps 502 to server unavailable", () => {
    const msg = resolveLoginErrorMessage(new ApiError(502, "Bad Gateway", { message: "Bad Gateway" }), t);
    expect(msg).toBe("auth.loginPage.serverUnavailable");
  });

  it("maps network errors to server unavailable", () => {
    const msg = resolveLoginErrorMessage(new TypeError("Failed to fetch"), t);
    expect(msg).toBe("auth.loginPage.serverUnavailable");
  });
});
