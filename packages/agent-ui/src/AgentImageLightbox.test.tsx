import "@testing-library/jest-dom/vitest";

import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { ImageLightboxProvider, ZoomableImage } from "./AgentImageLightbox";

afterEach(cleanup);

function renderImage() {
  return render(
    <ImageLightboxProvider>
      <ZoomableImage alt="diagram" className="size-10" src="/pic.png" />
    </ImageLightboxProvider>,
  );
}

describe("AgentImageLightbox", () => {
  it("keeps image semantics on the thumbnail", () => {
    renderImage();
    expect(
      screen.getByRole("button", { name: "Zoom image: diagram" }),
    ).toBeInTheDocument();
    const img = screen.getByRole("img", { name: "diagram" });
    expect(img).toHaveAttribute("src", "/pic.png");
    expect(img).toHaveClass("size-10");
  });

  it("opens and closes the full image viewer", async () => {
    renderImage();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Zoom image: diagram" }));
    expect(screen.getByRole("dialog", { name: "diagram" })).toHaveAttribute(
      "aria-modal",
      "true",
    );
    const close = screen.getByRole("button", { name: "Close" });
    await waitFor(() => expect(close).toHaveFocus());
    fireEvent.click(close);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("closes the viewer with Escape", () => {
    renderImage();
    fireEvent.click(screen.getByRole("button", { name: "Zoom image: diagram" }));
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("zooms in and resets the preview", () => {
    renderImage();
    fireEvent.click(screen.getByRole("button", { name: "Zoom image: diagram" }));
    const preview = screen.getAllByRole("img", { name: "diagram" }).at(-1)!;
    expect(screen.getByRole("button", { name: "Reset zoom" })).toHaveTextContent("100%");
    fireEvent.click(screen.getByRole("button", { name: "Zoom in" }));
    expect(screen.getByRole("button", { name: "Reset zoom" })).toHaveTextContent("150%");
    expect(preview).toHaveStyle({ transform: "translate(0px, 0px) scale(1.5)" });
    fireEvent.click(screen.getByRole("button", { name: "Reset zoom" }));
    expect(screen.getByRole("button", { name: "Reset zoom" })).toHaveTextContent("100%");
  });
});
