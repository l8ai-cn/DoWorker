import { render, screen } from "@testing-library/react";
import type { AnchorHTMLAttributes } from "react";
import { describe, expect, it, vi } from "vitest";
import { MobileWorkerConsoleHeader } from "./mobile-worker-console-header";

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    children,
    params,
    to,
    ...props
  }: AnchorHTMLAttributes<HTMLAnchorElement> & {
    params?: { podKey?: string };
    to: string;
  }) => <a {...props} href={to.replace("$podKey", params?.podKey ?? "")}>{children}</a>,
}));

describe("MobileWorkerConsoleHeader", () => {
  it("shows a token-free Preview route only when the Worker exposes Preview", () => {
    const { rerender } = render(
      <MobileWorkerConsoleHeader mode="acp" podKey="pod-1" previewAvailable />,
    );

    expect(screen.getByRole("link", { name: "打开预览" }).getAttribute("href")).toBe(
      "/workers/pod-1/preview",
    );

    rerender(<MobileWorkerConsoleHeader mode="pty" podKey="pod-1" previewAvailable={false} />);

    expect(screen.queryByRole("link", { name: "打开预览" })).toBeNull();
  });
});
