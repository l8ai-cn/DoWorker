import { describe, expect, it, vi } from "vitest";
import LegacyMobilePodPage from "./page";
import { permanentRedirect } from "next/navigation";

vi.mock("next/navigation", () => ({
  permanentRedirect: vi.fn(),
}));

describe("LegacyMobilePodPage", () => {
  it("permanently redirects to the canonical Worker route", async () => {
    await LegacyMobilePodPage({
      params: Promise.resolve({ org: "acme", podKey: "pod-1" }),
    });

    expect(permanentRedirect).toHaveBeenCalledWith(
      "/acme/mobile/workers/pod-1",
    );
  });
});
