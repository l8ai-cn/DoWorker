import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const publicDirectory = resolve(process.cwd(), "public");
const serviceWorker = readFileSync(resolve(publicDirectory, "sw.js"), "utf8");
const offlineShell = readFileSync(resolve(publicDirectory, "offline.html"), "utf8");

describe("offline shell", () => {
  it("serves a static document that does not depend on Next.js chunks", () => {
    expect(serviceWorker).toContain("'/offline.html'");
    expect(serviceWorker).toContain("caches.match('/offline.html')");
    expect(serviceWorker).not.toContain("caches.match('/offline')");
    expect(offlineShell).not.toContain("/_next/");
    expect(offlineShell).not.toContain("<script");
  });

  it("keeps the recovery action usable without application JavaScript", () => {
    expect(offlineShell).toContain('lang="zh-CN"');
    expect(offlineShell).toContain("网络连接中断");
    expect(offlineShell).toContain('<a class="retry" href="">重新连接</a>');
    expect(offlineShell).not.toContain("onclick=");
  });
});
