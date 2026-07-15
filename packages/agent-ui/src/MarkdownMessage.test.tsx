import { render, screen } from "@testing-library/react";

import { MarkdownMessage } from "./MarkdownMessage";

describe("MarkdownMessage", () => {
  it("renders remote images as explicit links without creating an image", () => {
    render(
      <MarkdownMessage text="![tracker](https://tracker.test/pixel.png)" />,
    );

    expect(screen.queryByRole("img", { name: "tracker" })).not.toBeInTheDocument();
    const link = screen.getByRole("link", { name: "tracker" });
    expect(link).toHaveAttribute("href", "https://tracker.test/pixel.png");
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
    expect(link).toHaveAttribute("referrerpolicy", "no-referrer");
  });

  it.each([
    ["blob:https://app.test/image-id", "blob"],
    ["data:image/png;base64,AA==", "data"],
  ])("renders %s image sources inline", (src, alt) => {
    render(<MarkdownMessage text={`![${alt}](${src})`} />);

    expect(screen.getByRole("img", { name: alt })).toHaveAttribute("src", src);
  });

  it("keeps ordinary markdown links unchanged", () => {
    render(<MarkdownMessage text="[documentation](https://docs.example.test)" />);

    expect(screen.getByRole("link", { name: "documentation" })).toHaveAttribute(
      "href",
      "https://docs.example.test",
    );
  });
});
