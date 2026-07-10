import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/react";
import MobilePodPreviewPage from "./page";
import { getPodPreviewSession } from "@/lib/api/podPreview";

const replaceMock = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme", podKey: "pod-1" }),
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/lib/api/podPreview", () => ({
  getPodPreviewSession: vi.fn().mockResolvedValue({
    preview_base_url: "https://relay.example/preview/pod-1/",
    session_url: "https://relay.example/preview/pod-1/__session?token=secret",
    expires_at: "2026-07-10T00:00:00Z",
  }),
}));

describe("MobilePodPreviewPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { replace: replaceMock },
    });
  });

  it("redirects only to the backend-issued preview session URL", async () => {
    render(<MobilePodPreviewPage />);

    await waitFor(() => {
      expect(getPodPreviewSession).toHaveBeenCalledWith("acme", "pod-1");
      expect(replaceMock).toHaveBeenCalledWith(
        "https://relay.example/preview/pod-1/__session?token=secret",
      );
    });
  });
});
