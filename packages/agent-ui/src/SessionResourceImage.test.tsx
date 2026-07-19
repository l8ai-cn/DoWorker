import "@testing-library/jest-dom/vitest";

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ImageLightboxProvider,
} from "./AgentImageLightbox";
import { SessionResourceImage } from "./SessionResourceImage";

let createObjectURL: ReturnType<typeof vi.fn>;
let revokeObjectURL: ReturnType<typeof vi.fn>;

beforeEach(() => {
  createObjectURL = vi.fn(() => "blob:session-image");
  revokeObjectURL = vi.fn();
  vi.stubGlobal("URL", { createObjectURL, revokeObjectURL });
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("SessionResourceImage", () => {
  it("shows the loading placeholder before the loader resolves", () => {
    renderImage(vi.fn((_path: string) => new Promise<Blob>(() => {})));
    expect(
      screen.getByRole("status", { name: "Loading image" }),
    ).toBeInTheDocument();
  });

  it("loads session content through the host loader", async () => {
    const blob = new Blob(["x"], { type: "image/png" });
    const loadBlob = vi.fn(async () => blob);
    renderImage(loadBlob);
    const img = await screen.findByRole("img", { name: "pic" });
    expect(img).toHaveAttribute("src", "blob:session-image");
    expect(img).toHaveClass("cls");
    expect(loadBlob).toHaveBeenCalledWith("/p");
    expect(createObjectURL).toHaveBeenCalledWith(blob);
  });

  it("renders an error fallback for missing paths and loader failures", async () => {
    const loadBlob = vi.fn(async () => {
      throw new Error("missing");
    });
    renderImage(loadBlob, "/missing");
    await waitFor(() =>
      expect(screen.getByRole("img", { name: "pic" })).toHaveTextContent("pic"),
    );

    cleanup();
    loadBlob.mockClear();
    renderImage(loadBlob, undefined);
    await waitFor(() =>
      expect(screen.getByRole("img", { name: "pic" })).toBeInTheDocument(),
    );
    expect(loadBlob).toHaveBeenCalledTimes(1);
  });

  it("revokes the object URL on unmount", async () => {
    const blob = new Blob(["x"], { type: "image/png" });
    const { unmount } = renderImage(vi.fn(async () => blob));
    await screen.findByRole("img", { name: "pic" });
    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:session-image");
  });
});

function renderImage(
  loadBlob: (path: string) => Promise<Blob>,
  path: string | undefined = "/p",
) {
  return render(
    <ImageLightboxProvider>
      <SessionResourceImage
        alt="pic"
        className="cls"
        loadBlob={loadBlob}
        path={path}
      />
    </ImageLightboxProvider>,
  );
}
