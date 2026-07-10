import { describe, expect, it, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { PodMobileAccessDialog } from "../PodMobileAccessDialog";

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

describe("PodMobileAccessDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.history.replaceState(null, "", "http://localhost:3000/acme/workspace");
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
    });
  });

  it("shows only the mobile console link when preview is not enabled", () => {
    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 0 }}
      />,
    );

    const expected = "http://localhost:3000/acme/mobile/pods/pod-1";
    expect(screen.getByTestId("qr-code")).toHaveAttribute("data-value", expected);
    expect(screen.getByRole("link", { name: "mobile.access.open" })).toHaveAttribute("href", expected);
    expect(screen.queryByRole("tab", { name: "mobile.access.preview" })).not.toBeInTheDocument();
  });

  it("offers a token-free preview link when preview metadata exists", () => {
    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 3000 }}
      />,
    );

    fireEvent.click(screen.getByRole("tab", { name: "mobile.access.preview" }));

    const value = screen.getByTestId("qr-code").getAttribute("data-value") ?? "";
    expect(value).toBe("http://localhost:3000/acme/mobile/pods/pod-1/preview");
    expect(value).not.toContain("token=");
  });

  it("copies the selected mobile link", () => {
    render(
      <PodMobileAccessDialog
        open
        onOpenChange={vi.fn()}
        orgSlug="acme"
        pod={{ pod_key: "pod-1", preview_port: 3000 }}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "mobile.access.copy" }));

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      "http://localhost:3000/acme/mobile/pods/pod-1",
    );
  });
});
