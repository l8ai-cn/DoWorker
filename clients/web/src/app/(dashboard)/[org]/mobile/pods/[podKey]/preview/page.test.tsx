import { describe, expect, it, vi } from "vitest";
import LegacyMobilePodPreviewPage from "./page";
import { permanentRedirect } from "next/navigation";

vi.mock("next/navigation", () => ({
  permanentRedirect: vi.fn(),
}));

describe("LegacyMobilePodPreviewPage", () => {
  it("permanently redirects to the canonical preview route", async () => {
    await LegacyMobilePodPreviewPage({
      params: Promise.resolve({ org: "acme", podKey: "pod-1" }),
    });

    expect(permanentRedirect).toHaveBeenCalledWith(
      "/acme/mobile/workers/pod-1/preview",
    );
  });
});
