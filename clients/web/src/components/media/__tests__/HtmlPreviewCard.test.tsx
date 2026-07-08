import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { HtmlPreviewCard } from "../HtmlPreviewCard";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

const HTML = "<html><body><h1>Hello</h1></body></html>";

describe("HtmlPreviewCard", () => {
  it("renders the preview iframe by default for complete inline html", () => {
    const { container } = render(<HtmlPreviewCard html={HTML} />);
    const iframe = container.querySelector("iframe");
    expect(iframe).toBeTruthy();
    expect(iframe?.getAttribute("srcdoc")).toBe(HTML);
    expect(iframe?.getAttribute("sandbox")).toBe("allow-scripts");
  });

  it("stays on the code tab while streaming", () => {
    const { container } = render(<HtmlPreviewCard html={HTML} streaming />);
    expect(container.querySelector("iframe")).toBeNull();
    expect(container.querySelector("pre code")?.textContent).toBe(HTML);
  });

  it("switches to preview automatically when streaming completes", () => {
    const { container, rerender } = render(<HtmlPreviewCard html={HTML} streaming />);
    expect(container.querySelector("iframe")).toBeNull();
    rerender(<HtmlPreviewCard html={HTML} streaming={false} />);
    expect(container.querySelector("iframe")).toBeTruthy();
  });

  it("respects a manual tab choice across streaming transitions", () => {
    const { container, rerender } = render(<HtmlPreviewCard html={HTML} streaming />);
    fireEvent.click(screen.getByRole("button", { name: /code/i }));
    rerender(<HtmlPreviewCard html={HTML} streaming={false} />);
    expect(container.querySelector("iframe")).toBeNull();
  });

  it("toggles between code and preview tabs", () => {
    const { container } = render(<HtmlPreviewCard html={HTML} />);
    fireEvent.click(screen.getByRole("button", { name: /code/i }));
    expect(container.querySelector("iframe")).toBeNull();
    expect(container.querySelector("pre code")?.textContent).toBe(HTML);
    fireEvent.click(screen.getByRole("button", { name: /preview/i }));
    expect(container.querySelector("iframe")).toBeTruthy();
  });

  it("renders remote html via src and hides the code tab", () => {
    const { container } = render(<HtmlPreviewCard src="https://cdn.example.com/report.html" />);
    const iframe = container.querySelector("iframe");
    expect(iframe?.getAttribute("src")).toBe("https://cdn.example.com/report.html");
    expect(screen.queryByRole("button", { name: /^code$/i })).toBeNull();
  });

  it("renders nothing for unsafe src without inline html", () => {
     
    const { container } = render(<HtmlPreviewCard src="javascript:alert(1)" />);
    expect(container.querySelector("[data-testid='html-preview-card']")).toBeNull();
  });
});
