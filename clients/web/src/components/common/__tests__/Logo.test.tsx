import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Logo } from "../Logo";

describe("Logo", () => {
  it("renders four capability modules as one Expert unit", () => {
    const { container } = render(<Logo className="h-8 w-8" />);
    const svg = container.querySelector("svg");

    expect(svg).toHaveAttribute("viewBox", "0 0 512 512");
    expect(svg).toHaveAttribute("aria-hidden", "true");
    expect(svg).toHaveClass("h-8", "w-8");
    expect(svg?.querySelector("[data-logo-background]")).toBeInTheDocument();
    expect(svg?.querySelectorAll("[data-logo-module]")).toHaveLength(4);
    expect(svg?.querySelectorAll("[data-logo-active-module]")).toHaveLength(1);
    expect(svg?.querySelector("[data-logo-passage]")).not.toBeInTheDocument();
    expect(svg?.querySelector("[data-logo-keystone]")).not.toBeInTheDocument();
    expect(svg?.querySelector("[data-logo-monogram]")).not.toBeInTheDocument();
    expect(svg?.querySelector("[data-logo-core]")).not.toBeInTheDocument();
    expect(svg?.querySelector("linearGradient")).not.toBeInTheDocument();
    expect(svg?.querySelector("filter")).not.toBeInTheDocument();
    expect(svg?.querySelector("text")).not.toBeInTheDocument();
  });
});
