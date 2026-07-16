import { beforeEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { renderToString } from "react-dom/server";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "@do-worker/agent-ui";
import { HtmlPreviewCard } from "../HtmlPreviewCard";

const openStaticHtmlInNewWindow = vi.hoisted(() => vi.fn());

vi.mock("@do-worker/agent-ui", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@do-worker/agent-ui")>()),
  openStaticHtmlInNewWindow,
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

const HTML = "<html><body><h1>Hello</h1></body></html>";

async function flushStaticDocument() {
  await act(async () => {
    await Promise.resolve();
  });
}

describe("HtmlPreviewCard", () => {
  beforeEach(() => {
    openStaticHtmlInNewWindow.mockClear();
    openStaticHtmlInNewWindow.mockReturnValue({ opened: true });
    vi.spyOn(window, "open").mockReturnValue(null);
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:unsafe-top-level-document"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
    });
  });

  it("builds the static document after the initial client render", async () => {
    const { container } = render(<HtmlPreviewCard html={HTML} />);
    const iframe = container.querySelector("iframe");
    expect(iframe).toBeTruthy();
    expect(iframe).not.toHaveAttribute("srcdoc");
    expect(iframe).toHaveAttribute("sandbox", STATIC_HTML_SANDBOX);
    expect(iframe).toHaveAttribute("referrerpolicy", STATIC_HTML_REFERRER_POLICY);
    await waitFor(() => {
      expect(iframe).toHaveAttribute("srcdoc", staticHtmlDocument(HTML));
    });
  });

  it("renders on the server without DOMParser or srcDoc", () => {
    vi.stubGlobal("DOMParser", undefined);

    try {
      const output = renderToString(<HtmlPreviewCard html={HTML} />);
      expect(output).not.toContain("srcDoc=");
      expect(output).not.toContain("srcdoc=");
    } finally {
      vi.unstubAllGlobals();
    }
  });

  it("rebuilds the static document without reusing stale html", async () => {
    const updatedHtml = "<html><body><h1>Updated</h1></body></html>";
    const { container, rerender } = render(<HtmlPreviewCard html={HTML} />);
    const iframe = container.querySelector("iframe");
    await waitFor(() => {
      expect(iframe).toHaveAttribute("srcdoc", staticHtmlDocument(HTML));
    });

    rerender(<HtmlPreviewCard html={updatedHtml} />);

    expect(iframe).not.toHaveAttribute("srcdoc");
    await waitFor(() => {
      expect(iframe).toHaveAttribute("srcdoc", staticHtmlDocument(updatedHtml));
    });
  });

  it("stays on the code tab while streaming", async () => {
    const { container } = render(<HtmlPreviewCard html={HTML} streaming />);
    await flushStaticDocument();
    expect(container.querySelector("iframe")).toBeNull();
    expect(container.querySelector("pre code")?.textContent).toBe(HTML);
  });

  it("switches to preview automatically when streaming completes", async () => {
    const { container, rerender } = render(<HtmlPreviewCard html={HTML} streaming />);
    await flushStaticDocument();
    expect(container.querySelector("iframe")).toBeNull();
    rerender(<HtmlPreviewCard html={HTML} streaming={false} />);
    expect(container.querySelector("iframe")).toBeTruthy();
  });

  it("respects a manual tab choice across streaming transitions", async () => {
    const { container, rerender } = render(<HtmlPreviewCard html={HTML} streaming />);
    await flushStaticDocument();
    fireEvent.click(screen.getByRole("button", { name: /code/i }));
    rerender(<HtmlPreviewCard html={HTML} streaming={false} />);
    expect(container.querySelector("iframe")).toBeNull();
  });

  it("toggles between code and preview tabs", async () => {
    const { container } = render(<HtmlPreviewCard html={HTML} />);
    await flushStaticDocument();
    fireEvent.click(screen.getByRole("button", { name: /code/i }));
    expect(container.querySelector("iframe")).toBeNull();
    expect(container.querySelector("pre code")?.textContent).toBe(HTML);
    fireEvent.click(screen.getByRole("button", { name: /preview/i }));
    expect(container.querySelector("iframe")).toBeTruthy();
  });

  it("opens inline html through the shared static document shell", async () => {
    render(<HtmlPreviewCard html={HTML} />);
    await flushStaticDocument();

    fireEvent.click(screen.getByRole("button", { name: "openInNewTab" }));

    expect(openStaticHtmlInNewWindow).toHaveBeenCalledWith(HTML, "htmlDocument");
    expect(window.open).not.toHaveBeenCalledWith(
      expect.stringMatching(/^blob:/),
      "_blank",
      expect.anything(),
    );
  });

  it("shows an inline error when the browser blocks the preview window", async () => {
    openStaticHtmlInNewWindow.mockReturnValue({
      opened: false,
      reason: "popup-blocked",
    });
    render(<HtmlPreviewCard html={HTML} />);
    await flushStaticDocument();

    fireEvent.click(screen.getByRole("button", { name: "openInNewTab" }));

    expect(screen.getByRole("alert")).toHaveTextContent("popupBlocked");
  });
});
