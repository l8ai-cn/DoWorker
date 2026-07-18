import { vi } from "vitest";

import {
  STATIC_HTML_CSP,
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  openStaticHtmlInNewWindow,
  staticHtmlDocument,
} from "./staticHtmlProfile";

describe("static HTML profile", () => {
  it("runs scripts inside an opaque sandbox with no referrer", () => {
    expect(STATIC_HTML_SANDBOX).toBe("allow-scripts");
    expect(STATIC_HTML_SANDBOX).not.toContain("allow-same-origin");
    expect(STATIC_HTML_REFERRER_POLICY).toBe("no-referrer");
  });

  it("injects a deny-by-default CSP before artifact resources", () => {
    const html = staticHtmlDocument(
      '<html><head><base href="https://tracker.test/"><meta name="referrer" content="unsafe-url"><meta http-equiv="REFERRER" content="unsafe-url"><meta http-equiv="content-security-policy" content="default-src *"><meta http-equiv="ReFrEsH" content="0;url=https://tracker.test/escape"><link rel="stylesheet" href="https://tracker.test/x.css"></head><body><img src="https://tracker.test/p.png"></body></html>',
    );
    const parsed = new DOMParser().parseFromString(html, "text/html");
    const csp = [...parsed.head.querySelectorAll("meta")].filter(
      (element) => element.httpEquiv.toLowerCase() === "content-security-policy",
    );
    const base = parsed.head.querySelectorAll("base");
    const referrer = [...parsed.head.querySelectorAll("meta")].filter(
      (element) =>
        element.name.toLowerCase() === "referrer" ||
        element.httpEquiv.toLowerCase() === "referrer",
    );
    const refresh = [...parsed.head.querySelectorAll("meta")].filter(
      (element) => element.httpEquiv.toLowerCase() === "refresh",
    );

    expect(csp).toHaveLength(1);
    expect(csp[0]?.getAttribute("content")).toBe(STATIC_HTML_CSP);
    expect(parsed.head.firstElementChild).toBe(csp[0]);
    expect(base).toHaveLength(1);
    expect(base[0]?.getAttribute("target")).toBe("_blank");
    expect(base[0]?.hasAttribute("href")).toBe(false);
    expect(referrer).toHaveLength(1);
    expect(referrer[0]?.getAttribute("content")).toBe("no-referrer");
    expect(refresh).toHaveLength(0);
    expect(STATIC_HTML_CSP).toContain("default-src 'none'");
    expect(STATIC_HTML_CSP).toContain("script-src 'unsafe-inline'");
    expect(STATIC_HTML_CSP).toContain("connect-src 'none'");
    expect(STATIC_HTML_CSP).toContain("frame-src 'none'");
    expect(STATIC_HTML_CSP).toContain("worker-src 'none'");
    expect(STATIC_HTML_CSP).toContain("form-action 'none'");
  });

  it("returns an explicit result when the shell cannot be opened", () => {
    const open = vi.fn(() => null);

    expect(
      openStaticHtmlInNewWindow("<h1>blocked</h1>", "Blocked", { open }),
    ).toEqual({
      opened: false,
      reason: "popup-blocked",
    });
    expect(open).toHaveBeenCalledTimes(1);
  });

  it("writes untrusted HTML only into an isolated iframe", () => {
    const shellDocument = document.implementation.createHTMLDocument("");
    const shell = {
      document: shellDocument,
      opener: window,
    };
    const open = vi.fn(() => shell as unknown as Window);
    const html = "<h1>Preview</h1><script>window.compromised = true</script>";

    const result = openStaticHtmlInNewWindow(html, "Artifact preview", {
      open,
    });

    expect(result).toEqual({ opened: true });
    expect(open).toHaveBeenCalledWith("about:blank", "_blank");
    expect(shell.opener).toBeNull();
    expect(shellDocument.title).toBe("Artifact preview");

    const frame = shellDocument.querySelector("iframe");
    expect(frame).not.toBeNull();
    expect(frame?.getAttribute("sandbox")).toBe("allow-scripts");
    expect(frame?.getAttribute("referrerpolicy")).toBe("no-referrer");
    expect(frame?.hasAttribute("src")).toBe(false);
    expect(frame?.srcdoc).toBe(staticHtmlDocument(html));
    const frameDocument = new DOMParser().parseFromString(frame?.srcdoc ?? "", "text/html");
    expect(frameDocument.head.querySelector("base")?.getAttribute("target")).toBe("_blank");
    expect(frameDocument.head.querySelector("base")?.hasAttribute("href")).toBe(false);
    expect(shellDocument.body.childElementCount).toBe(1);
    expect(shellDocument.body.querySelector("script")).toBeNull();
  });
});
