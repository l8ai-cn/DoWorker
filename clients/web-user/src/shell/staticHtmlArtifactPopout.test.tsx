import { act, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "@agent-cloud/agent-ui";
import { openHtmlArtifactInNewTab } from "./staticHtmlArtifactPopout";
import { Toaster } from "@/components/ui/toast";

describe("openHtmlArtifactInNewTab", () => {
  it("uses an opaque-origin static HTML iframe", () => {
    const shellDocument = document.implementation.createHTMLDocument("");
    const open = vi.fn(() => ({ document: shellDocument }) as unknown as Window);

    expect(openHtmlArtifactInNewTab("<h1>hi</h1>", "art.html", { open })).toBe(true);
    expect(open).toHaveBeenCalledWith("about:blank", "_blank");
    const frame = shellDocument.querySelector("iframe");
    expect(frame?.getAttribute("sandbox")).toBe(STATIC_HTML_SANDBOX);
    expect(frame?.getAttribute("referrerpolicy")).toBe(STATIC_HTML_REFERRER_POLICY);
    expect(frame?.getAttribute("srcdoc")).toBe(staticHtmlDocument("<h1>hi</h1>"));
  });

  it("shows a visible localized error when the popup is blocked", () => {
    document.documentElement.lang = "zh-CN";
    render(<Toaster />);
    const open = vi.fn(() => null);
    act(() => {
      expect(openHtmlArtifactInNewTab("<h1>hi</h1>", "art.html", { open })).toBe(false);
    });
    expect(screen.getByTestId("toast")).toHaveTextContent("浏览器阻止了新窗口");
  });
});
