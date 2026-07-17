import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { MarkdownPreview } from "./MarkdownPreview";

function renderPreview(content: string) {
  return render(<MarkdownPreview content={content} />);
}

afterEach(cleanup);

describe("MarkdownPreview image policy", () => {
  it("renders an http image as an explicit link without creating an img", () => {
    renderPreview("![tracker](https://tracker.test/p.png)");

    expect(screen.queryByRole("img", { name: "tracker" })).toBeNull();
    const link = screen.getByRole("link", { name: "tracker" });
    expect(link.getAttribute("href")).toBe("https://tracker.test/p.png");
    expect(link.getAttribute("target")).toBe("_blank");
    expect(link.getAttribute("rel")).toContain("noopener");
    expect(link.getAttribute("rel")).toContain("noreferrer");
  });

  it("applies the same remote-image policy to raw HTML img tags", () => {
    renderPreview('<img src="http://tracker.test/raw.png" alt="raw tracker">');

    expect(screen.queryByRole("img", { name: "raw tracker" })).toBeNull();
    expect(screen.getByRole("link", { name: "raw tracker" })).toBeDefined();
  });

  it("removes remote picture sources and img srcsets", () => {
    const { container } = renderPreview(
      '<picture><source srcset="https://tracker.test/pixel.png 1x"><img src="data:image/png;base64,AA==" srcset="https://tracker.test/retina.png 2x" alt="safe fallback"></picture>',
    );

    expect(container.querySelector("source")).toBeNull();
    expect(container.innerHTML).not.toContain("tracker.test");
    expect(screen.getByRole("img", { name: "safe fallback" })).toHaveAttribute(
      "src",
      "data:image/png;base64,AA==",
    );
  });

  it.each([
    ["blob:https://app.test/id", "blob image"],
    ["data:image/png;base64,AA==", "data image"],
  ])("renders the allowed %s source as an img", (src, alt) => {
    renderPreview(`![${alt}](${src})`);

    expect(screen.getByRole("img", { name: alt }).getAttribute("src")).toBe(src);
  });
});
