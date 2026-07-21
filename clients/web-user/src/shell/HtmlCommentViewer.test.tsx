import { cleanup, render } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "@agent-cloud/agent-ui";
import { HtmlCommentViewer } from "./HtmlCommentViewer";

afterEach(cleanup);

function renderViewer(content: string, truncated = false) {
  return render(<HtmlCommentViewer content={content} truncated={truncated} />);
}

describe("HtmlCommentViewer", () => {
  it("renders the preview with the shared static HTML policy", () => {
    const content = "<html><body><p>doc</p></body></html>";
    const { container } = renderViewer(content);
    const iframe = container.querySelector('iframe[title="HTML preview"]') as HTMLIFrameElement;
    expect(iframe).not.toBeNull();
    const sandbox = iframe.getAttribute("sandbox") ?? "";
    expect(sandbox).toBe(STATIC_HTML_SANDBOX);
    expect(iframe.getAttribute("referrerpolicy")).toBe(STATIC_HTML_REFERRER_POLICY);
    expect(iframe.getAttribute("srcdoc")).toBe(staticHtmlDocument(content));
  });

  it("does not inject the executable comment bridge into the static preview", () => {
    const { container } = renderViewer("<html><head></head><body><p>doc</p></body></html>");
    const iframe = container.querySelector('iframe[title="HTML preview"]') as HTMLIFrameElement;
    const srcDoc = iframe.getAttribute("srcdoc") ?? "";
    expect(srcDoc).not.toContain("<script>");
    expect(srcDoc).not.toContain("omni-html-comment");
  });

  it("shows the truncated banner only when truncated", () => {
    const { queryByText, rerender } = renderViewer("<body>x</body>", false);
    expect(queryByText(/truncated/i)).toBeNull();
    rerender(<HtmlCommentViewer content="<body>x</body>" truncated={true} />);
    expect(queryByText(/truncated/i)).not.toBeNull();
  });

  it("does not show the floating Add-comment button before any selection", () => {
    renderViewer("<body><p>doc</p></body>");
    expect(document.querySelector("[data-add-comment-btn]")).toBeNull();
  });
});
