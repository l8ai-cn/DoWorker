import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SiteHeader } from "./site-header";

describe("SiteHeader", () => {
  it("only exposes real public discovery navigation", () => {
    render(<SiteHeader />);

    expect(screen.getByRole("link", { name: "市场首页" })).toHaveAttribute(
      "href",
      "/",
    );
    expect(screen.getByRole("link", { name: "全部应用" })).toHaveAttribute(
      "href",
      "/catalog",
    );
    expect(screen.queryByText("即将开放")).toBeNull();
    expect(screen.queryByRole("link", { name: /我的应用|额度|账户/ })).toBeNull();
  });
});
