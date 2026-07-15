export const STATIC_HTML_SANDBOX = "";
export const STATIC_HTML_REFERRER_POLICY = "no-referrer";
export const STATIC_HTML_CSP = [
  "default-src 'none'",
  "img-src data: blob:",
  "media-src data: blob:",
  "style-src 'unsafe-inline'",
  "font-src data:",
  "connect-src 'none'",
  "frame-src 'none'",
  "worker-src 'none'",
  "object-src 'none'",
  "form-action 'none'",
  "base-uri 'none'",
].join("; ");

type StaticHtmlWindowOpenResult =
  | { opened: true }
  | { opened: false; reason: "popup-blocked" };

export function staticHtmlDocument(html: string): string {
  const parsed = new DOMParser().parseFromString(html, "text/html");
  parsed.querySelectorAll("base").forEach((element) => element.remove());
  parsed.querySelectorAll("meta").forEach((element) => {
    const name = element.name.toLowerCase();
    const httpEquiv = element.httpEquiv.toLowerCase();
    if (
      name === "referrer" ||
      httpEquiv === "referrer" ||
      httpEquiv === "content-security-policy"
    ) {
      element.remove();
    }
  });

  const csp = parsed.createElement("meta");
  csp.httpEquiv = "Content-Security-Policy";
  csp.content = STATIC_HTML_CSP;

  const referrer = parsed.createElement("meta");
  referrer.name = "referrer";
  referrer.content = STATIC_HTML_REFERRER_POLICY;

  const base = parsed.createElement("base");
  base.target = "_blank";
  parsed.head.prepend(csp, referrer, base);
  return `<!doctype html>${parsed.documentElement.outerHTML}`;
}

export function openStaticHtmlInNewWindow(
  html: string,
  title = "",
  opener: Pick<Window, "open"> = window,
): StaticHtmlWindowOpenResult {
  const shell = opener.open("about:blank", "_blank");
  if (!shell) {
    return { opened: false, reason: "popup-blocked" };
  }

  shell.opener = null;

  const frame = shell.document.createElement("iframe");
  frame.setAttribute("sandbox", STATIC_HTML_SANDBOX);
  frame.setAttribute("referrerpolicy", STATIC_HTML_REFERRER_POLICY);
  frame.srcdoc = staticHtmlDocument(html);
  frame.style.cssText = "border:0;height:100vh;width:100%";

  shell.document.title = title;
  shell.document.body.style.margin = "0";
  shell.document.body.replaceChildren(frame);

  return { opened: true };
}
