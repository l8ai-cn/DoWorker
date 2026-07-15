import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { Markdown } from "../markdown";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("Markdown media rendering", () => {
  it("keeps a remote html URL as an explicit link instead of embedding it", () => {
    const url = "https://cdn.example.test/report.html";
    const { container } = render(
      <Markdown content={`[report.html](${url})`} enableMedia />,
    );

    expect(screen.getByRole("link", { name: "report.html" })).toHaveAttribute("href", url);
    expect(container.querySelector("iframe")).toBeNull();
  });

  it("keeps a remote image as an explicit link without loading it", () => {
    const url = "https://tracker.example.test/pixel.png";
    const { container } = render(
      <Markdown content={`![tracker](${url})`} enableMedia />,
    );

    expect(screen.queryByRole("img", { name: "tracker" })).toBeNull();
    expect(screen.getByRole("link", { name: "tracker" })).toHaveAttribute("href", url);
    expect(container.querySelector("img")).toBeNull();
  });

  it("applies the remote image policy when media upgrades are disabled", () => {
    const url = "https://tracker.example.test/default.png";
    const { container } = render(
      <Markdown content={`![default tracker](${url})`} />,
    );

    expect(screen.queryByRole("img", { name: "default tracker" })).toBeNull();
    expect(screen.getByRole("link", { name: "default tracker" })).toHaveAttribute(
      "href",
      url,
    );
    expect(container.querySelector("img")).toBeNull();
  });

  it("keeps a remote video URL as a link instead of auto-embedding it", () => {
    const url = "https://cdn.example.test/demo.mp4";
    const { container } = render(
      <Markdown content={`[demo.mp4](${url})`} enableMedia />,
    );

    expect(screen.getByRole("link", { name: "demo.mp4" })).toHaveAttribute("href", url);
    expect(container.querySelector("video")).toBeNull();
    expect(container.querySelector("iframe")).toBeNull();
  });

  it("renders inline data images through the media lightbox", () => {
    render(
      <Markdown
        content="![generated](data:image/png;base64,AA==)"
        enableMedia
      />,
    );

    expect(screen.getByRole("img", { name: "generated" })).toHaveAttribute(
      "src",
      "data:image/png;base64,AA==",
    );
  });
});
