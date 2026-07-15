import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LightAuthButtons } from "../LightAuthButtons";

const state = vi.hoisted(() => ({
  session: {
    hydrated: true,
    session: {
      isAuthenticated: true,
      currentOrgSlug: null as string | null,
    },
  },
  push: vi.fn(),
  discoverFirstOrgSlug: vi.fn(),
  updateOrgSlug: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: state.push }),
}));
vi.mock("@/hooks/useLightSession", () => ({
  useLightSession: () => state.session,
}));
vi.mock("@/lib/light-auth", () => ({
  discoverFirstOrgSlug: state.discoverFirstOrgSlug,
}));
vi.mock("@/lib/light-session", () => ({
  updateLightSessionOrgSlug: state.updateOrgSlug,
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => ({
    "landing.nav.console": "Console",
    "landing.nav.consoleUnavailable": "Console unavailable",
    "landing.nav.signIn": "Sign in",
    "landing.nav.getStarted": "Get started",
  })[key] ?? key,
}));

describe("LightAuthButtons", () => {
  beforeEach(() => {
    state.session.session = {
      isAuthenticated: true,
      currentOrgSlug: null,
    };
    state.push.mockReset();
    state.discoverFirstOrgSlug.mockReset();
    state.updateOrgSlug.mockReset();
  });

  it("discovers an existing organization before opening the console", async () => {
    state.discoverFirstOrgSlug.mockResolvedValue({ status: "found", slug: "acme" });
    render(<LightAuthButtons />);

    fireEvent.click(screen.getByRole("button", { name: "Console" }));

    await waitFor(() => {
      expect(state.updateOrgSlug).toHaveBeenCalledWith("acme");
      expect(state.push).toHaveBeenCalledWith("/acme/workspace");
    });
  });

  it("sends authenticated users without an organization to onboarding", async () => {
    state.discoverFirstOrgSlug.mockResolvedValue({ status: "empty" });
    render(<LightAuthButtons />);

    fireEvent.click(screen.getByRole("button", { name: "Console" }));

    await waitFor(() => {
      expect(state.push).toHaveBeenCalledWith("/onboarding/create-org");
    });
  });

  it("reports organization discovery failures without changing routes", async () => {
    state.discoverFirstOrgSlug.mockResolvedValue({ status: "unavailable" });
    render(<LightAuthButtons />);

    fireEvent.click(screen.getByRole("button", { name: "Console" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Console unavailable");
    expect(state.push).not.toHaveBeenCalled();
  });
});
