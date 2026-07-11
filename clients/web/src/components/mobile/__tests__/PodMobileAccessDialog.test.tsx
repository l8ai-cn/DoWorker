import { describe, expect, it, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { PodMobileAccessDialog } from "../PodMobileAccessDialog";
import { getMobileAccessDescriptor } from "@/lib/api/facade/podConnect";

const h = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("qrcode.react", () => ({
  QRCodeSVG: ({ value }: { value: string }) => (
    <div data-testid="qr-code" data-value={value} />
  ),
}));

vi.mock("sonner", () => ({
  toast: {
    success: h.toastSuccess,
    error: h.toastError,
  },
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/lib/api/facade/podConnect", () => ({
  getMobileAccessDescriptor: vi.fn(),
}));

describe("PodMobileAccessDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
    });
    vi.mocked(getMobileAccessDescriptor).mockResolvedValue({
      canonical_url: "https://app.example/acme/mobile/workers/pod-1",
      pod_key: "pod-1",
      status: "running",
      interaction_mode: "pty",
      console_available: true,
      preview_available: false,
      relay_available: true,
    });
  });

  it("uses the backend canonical URL instead of the browser origin", async () => {
    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 0 }}
      />,
    );

    const expected = "https://app.example/acme/mobile/workers/pod-1";
    await waitFor(() =>
      expect(screen.getByTestId("qr-code")).toHaveAttribute(
        "data-value",
        expected,
      ),
    );
    expect(getMobileAccessDescriptor).toHaveBeenCalledWith("acme", "pod-1");
    expect(screen.getByTestId("qr-code")).toHaveAttribute("data-value", expected);
    expect(screen.getByRole("link", { name: "mobile.access.open" })).toHaveAttribute("href", expected);
    expect(screen.queryByRole("tab", { name: "mobile.access.preview" })).not.toBeInTheDocument();
  });

  it("offers a token-free preview link when the backend enables it", async () => {
    vi.mocked(getMobileAccessDescriptor).mockResolvedValue({
      canonical_url: "https://app.example/acme/mobile/workers/pod-1",
      pod_key: "pod-1",
      status: "running",
      interaction_mode: "pty",
      console_available: true,
      preview_available: true,
      relay_available: true,
      preview_path: "/",
    });
    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 3000 }}
      />,
    );

    await screen.findByRole("tab", { name: "mobile.access.preview" });
    fireEvent.click(screen.getByRole("tab", { name: "mobile.access.preview" }));

    const value = screen.getByTestId("qr-code").getAttribute("data-value") ?? "";
    expect(value).toBe("https://app.example/acme/mobile/workers/pod-1/preview");
    expect(value).not.toContain("token=");
  });

  it("copies the selected mobile link", async () => {
    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 3000 }}
      />,
    );

    await screen.findByTestId("qr-code");
    fireEvent.click(screen.getByRole("button", { name: "mobile.access.copy" }));

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      "https://app.example/acme/mobile/workers/pod-1",
    );
  });

  it("fails closed when the descriptor cannot be loaded", async () => {
    vi.mocked(getMobileAccessDescriptor).mockRejectedValue(
      new Error("public base URL not configured"),
    );

    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 3000 }}
      />,
    );

    expect(await screen.findByText("public base URL not configured")).toBeInTheDocument();
    expect(screen.queryByTestId("qr-code")).not.toBeInTheDocument();
  });
});
