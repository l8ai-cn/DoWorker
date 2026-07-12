import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AcquireButton } from "./acquire-button";

const target = {
  market: "do-worker-market",
  listing: "software-delivery-expert",
  version: "301",
};

describe("AcquireButton", () => {
  it("links applications to Core Web acquisition", () => {
    render(
      <AcquireButton
        coreWebUrl="https://app.l8ai.cn"
        resourceType="application"
        target={target}
      />,
    );

    expect(screen.getByRole("link", { name: "启用应用" })).toHaveAttribute(
      "href",
      "https://app.l8ai.cn/marketplace/acquire?market=do-worker-market&listing=software-delivery-expert&version=301",
    );
  });

  it("explains why acquisition is disabled when Core Web is not configured", () => {
    render(
      <AcquireButton
        coreWebUrl={undefined}
        resourceType="application"
        target={target}
      />,
    );

    expect(screen.getByRole("button", { name: "启用应用" })).toBeDisabled();
    expect(screen.getByText("获取入口尚未配置")).toBeInTheDocument();
  });

  it("does not expose unsupported Skill installation as available", () => {
    render(
      <AcquireButton
        coreWebUrl="https://dowork.l8ai.cn"
        resourceType="skill"
        target={{ market: "market", listing: "skill", version: "1" }}
      />,
    );

    expect(screen.getByRole("button", { name: "安装 Skill" })).toBeDisabled();
    expect(screen.getByText("对应运行时接入后开放")).toBeInTheDocument();
  });
});
