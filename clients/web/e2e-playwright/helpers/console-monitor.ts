// Default-deny console monitor. Every `console.error` and every
// `page.on("pageerror", ...)` (= uncaught exception in renderer) becomes
// a teardown failure unless a spec has explicitly opted out via
// `monitor.allow(/regex/)`. Replaces the old `assertNoWasmErrors`
// allow-list mechanism (7 hard-coded patterns, only 4 specs used it).
//
// Why default-deny: the WASM bridge regression that motivated this PR
// (R6 store calling a deleted `set_pods` method → TypeError → silent
// catch in zustand → empty sidebar) produced a `console.error` every
// time, but no e2e caught it because the workspace spec never read the
// console. With default-deny + automatic fixture attach (see
// fixtures/index.ts), forgetting to wire the monitor is impossible.
//
// Specs should narrowly allow legitimate noise (third-party SDKs that
// spam errors on dev backends, etc.) and **never** silence "missing
// field" / "is not valid JSON" / "TypeError" — those are exactly the
// signals this exists to catch.

import { expect, type ConsoleMessage, type Page } from "@playwright/test";

export interface ConsoleMonitor {
  /** Allow any message matching `pattern`. Multiple patterns OR together. */
  allow(pattern: RegExp): void;
  /** Snapshot of currently-recorded offending entries (for ad-hoc debug). */
  errors(): ReadonlyArray<string>;
  /** Throw if any non-allowed error has been collected. Called in teardown. */
  assertClean(): void;
  /** Detach Playwright listeners — fixtures call this after `assertClean`. */
  dispose(): void;
}

/**
 * wasm-bindgen's RefCell-style borrow check throws this exact string when
 * `&mut self` and `&self` calls overlap on the same WASM object. The ACP
 * session manager regression (render-time wasm read racing relay-driven
 * mutators) surfaced as exactly this message. Kept as a standalone assert
 * because some specs want it to fail loudly the moment it appears, not at
 * teardown, so the trace snapshot lines up with the offending action.
 */
export function assertNoWasmRecursiveBorrow(errors: ReadonlyArray<string>): void {
  const offenders = errors.filter((e) =>
    e.includes("recursive use of an object detected"),
  );
  expect(
    offenders,
    `wasm borrow conflict regressed:\n${offenders.join("\n")}`,
  ).toHaveLength(0);
}

export function createConsoleMonitor(page: Page): ConsoleMonitor {
  const errors: string[] = [];
  const allowPatterns: RegExp[] = [
    // Next.js dev-mode noise — React Server Component performance
    // measurement passes a negative timestamp during fast-refresh /
    // redirect bursts. Not a runtime issue, only fires in dev mode.
    // (`Performance.measure` will throw if endTime < startTime.)
    /Failed to execute 'measure' on 'Performance'.*cannot have a negative time stamp/i,
    // Next.js dev devtools intercept stub — extra log layer wraps any
    // error so the message appears twice with [console] prefix.
    /next-devtools.*intercept-console-error/i,
    // 4xx HTTP responses surface as `Failed to load resource: ... 4XX`
    // console.errors via the browser. These are expected application
    // errors (401 unauth, 403 forbidden, 404 not found, 409 conflict,
    // 422 validation) that specs intentionally trigger to verify error
    // paths. 5xx is NOT covered — server-side bugs must still fail e2e.
    /Failed to load resource:.*status of 4[0-9]{2}/i,
    // Transport-layer hiccups on the dev server's OWN assets. The Next.js
    // turbopack devserver briefly drops connections under hot-reload churn,
    // an OOM-restart, or a heavy-route on-demand compile — surfacing as
    // net::ERR_* on its _next chunks / fonts / the page document. These are
    // infra flakes, categorically NOT app errors: 5xx, TypeError, "missing
    // field", "is not valid JSON" are all still caught (none are net::ERR_
    // transport codes). A dead server yields a blank page that still fails
    // the spec's real content assertions, so allowing these only removes a
    // redundant, noisy failure channel — it cannot mask a code regression.
    /Failed to load resource:.*net::ERR_(CONNECTION_REFUSED|EMPTY_RESPONSE|INCOMPLETE_CHUNKED_ENCODING|CONNECTION_RESET|CONNECTION_CLOSED|NETWORK_CHANGED|ABORTED)/i,
    // Next.js regex-in-pattern parsing on register page (browser's `v`
    // flag stricter than what the form uses). Cosmetic — submission
    // still validates server-side. TODO: pattern attribute should use
    // `[a-zA-Z0-9_\\-]+` (escaped hyphen) to satisfy `v` flag.
    /Pattern attribute value.*not a valid regular expression/i,
    // Next.js 15 SSR/CSR mismatch on auth marketing shells under turbopack
    // cold compile — tree regenerates on the client; not an app regression.
    /Hydration failed because the server rendered HTML didn't match the client/i,
  ];

  const onConsole = (msg: ConsoleMessage) => {
    if (msg.type() !== "error") return;
    errors.push(formatConsoleMessage(msg));
  };
  const onPageError = (err: Error) => {
    errors.push(`[pageerror] ${err.message}\n${err.stack ?? ""}`);
  };

  page.on("console", onConsole);
  page.on("pageerror", onPageError);

  return {
    allow(pattern) {
      allowPatterns.push(pattern);
    },
    errors() {
      return errors.slice();
    },
    assertClean() {
      const offenders = errors.filter((e) => !allowPatterns.some((p) => p.test(e)));
      if (offenders.length === 0) return;
      const summary = offenders.map((e, i) => `  [${i + 1}] ${e}`).join("\n");
      throw new Error(
        `console-monitor: ${offenders.length} unallowed console.error/pageerror entries:\n${summary}\n\n` +
          `If any of these are expected for this spec, narrow them with monitor.allow(/regex/). ` +
          `Do NOT broaden a pattern to mask "missing field" / "is not valid JSON" / "TypeError" — ` +
          `those exist to catch wasm bridge / proto drift regressions.`,
      );
    },
    dispose() {
      page.off("console", onConsole);
      page.off("pageerror", onPageError);
    },
  };
}

function formatConsoleMessage(msg: ConsoleMessage): string {
  const loc = msg.location();
  const where = loc.url ? ` (${loc.url}:${loc.lineNumber}:${loc.columnNumber})` : "";
  return `[console.error]${where} ${msg.text()}`;
}
