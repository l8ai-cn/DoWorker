import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AuthBootstrapGate } from "./auth-bootstrap-gate";

vi.mock("@/lib/auth-store", () => ({
  restoreAuthIdentity: vi.fn(),
}));

import { restoreAuthIdentity } from "@/lib/auth-store";

describe("AuthBootstrapGate", () => {
  it("waits for the shared auth manager before rendering application routes", async () => {
    let resolveRestore: (() => void) | undefined;
    vi.mocked(restoreAuthIdentity).mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveRestore = () =>
            resolve({ authenticated: true, email: "dev@example.com", orgSlug: "dev" });
        }),
    );

    render(
      <AuthBootstrapGate>
        <p>application content</p>
      </AuthBootstrapGate>,
    );

    expect(screen.queryByText("application content")).toBeNull();
    resolveRestore?.();
    await waitFor(() => expect(screen.getByText("application content")).toBeTruthy());
  });
});
