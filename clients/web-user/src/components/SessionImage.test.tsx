import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const authenticatedFetch = vi.fn();

vi.mock("@/lib/identity", () => ({
  authenticatedFetch: (path: string) => authenticatedFetch(path),
}));

vi.mock("@/components/ui/spinner", () => ({
  Spinner: () => <span data-testid="spinner" />,
}));

import { SessionImage } from "./SessionImage";

let createObjectURL: ReturnType<typeof vi.fn>;
let revokeObjectURL: ReturnType<typeof vi.fn>;

beforeEach(() => {
  createObjectURL = vi.fn(() => "blob:fake-url");
  revokeObjectURL = vi.fn();
  vi.stubGlobal("URL", { createObjectURL, revokeObjectURL });
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});

describe("SessionImage", () => {
  it("shows the loading placeholder before the fetch resolves", () => {
    authenticatedFetch.mockReturnValue(new Promise(() => {}));
    render(<SessionImage path="/p" alt="pic" />);
    expect(screen.getByRole("status", { name: "Loading image" })).toBeInTheDocument();
    expect(screen.getByTestId("spinner")).toBeInTheDocument();
  });

  it("loads standalone session content through authenticated fetch", async () => {
    const blob = new Blob(["x"]);
    authenticatedFetch.mockResolvedValue({ ok: true, blob: () => Promise.resolve(blob) });
    render(<SessionImage path="/p" alt="pic" className="cls" />);
    const img = await screen.findByRole("img", { name: "pic" });
    expect(img).toHaveAttribute("src", "blob:fake-url");
    expect(img).toHaveClass("cls");
    expect(createObjectURL).toHaveBeenCalledWith(blob);
    expect(authenticatedFetch).toHaveBeenCalledWith("/p");
  });

  it("renders the error fallback when the response is not ok", async () => {
    authenticatedFetch.mockResolvedValue({ ok: false, status: 404 });
    render(<SessionImage path="/missing" alt="gone" />);
    await waitFor(() => {
      const fallback = screen.getByRole("img", { name: "gone" });
      expect(fallback).not.toHaveAttribute("src");
      expect(fallback).toHaveTextContent("gone");
    });
  });

  it("renders the error fallback when the fetch rejects", async () => {
    authenticatedFetch.mockRejectedValue(new Error("boom"));
    render(<SessionImage path="/p" alt="broken" />);
    await waitFor(() => {
      expect(screen.getByRole("img", { name: "broken" })).toHaveTextContent("broken");
    });
  });

  it("renders the error fallback immediately when no path is given", async () => {
    render(<SessionImage path={undefined} alt="nopath" />);
    await waitFor(() => {
      expect(screen.getByRole("img", { name: "nopath" })).toBeInTheDocument();
    });
    expect(authenticatedFetch).not.toHaveBeenCalled();
  });

  it("revokes the object URL on unmount to avoid leaking blobs", async () => {
    const blob = new Blob(["x"]);
    authenticatedFetch.mockResolvedValue({ ok: true, blob: () => Promise.resolve(blob) });
    const { unmount } = render(<SessionImage path="/p" alt="pic" />);
    await screen.findByRole("img", { name: "pic" });
    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:fake-url");
  });
});
