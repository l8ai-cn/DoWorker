import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SiteHeader } from "./site-header";

describe("SiteHeader", () => {
  it("keeps unavailable MVP destinations non-interactive", () => {
    vi.stubEnv("CORE_WEB_URL", "https://app.l8ai.cn");
    render(<SiteHeader />);

    expect(screen.queryByRole("link", { name: /我的应用/ })).toBeNull();
    expect(screen.queryByRole("link", { name: /额度/ })).toBeNull();
    expect(screen.getAllByText("即将开放")).toHaveLength(3);
  });
});
