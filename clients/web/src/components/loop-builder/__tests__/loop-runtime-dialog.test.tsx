import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@/test/test-utils";
import { LoopRuntimeDialog } from "../loop-runtime-dialog";
import type { LoopRuntimeMessages } from "../loop-workbench-messages";

const snapshots = [
  { id: "31", alias: "结算环境", workerType: "codex", createdAt: "2026-07-15T00:00:00Z" },
  {
    id: "9007199254740993",
    alias: "审查环境",
    workerType: "claude-code",
    createdAt: "2026-07-15T00:00:00Z",
  },
];
const messages: LoopRuntimeMessages = {
  title: "选择运行环境",
  description: "运行环境只在本次启动时绑定，不属于循环编排。",
  field: "运行环境",
  placeholder: "选择运行环境",
  unnamed: "未命名环境",
  loading: "正在加载运行环境",
  retry: "重新加载",
  empty: "当前组织没有可用的运行环境",
  cancel: "取消",
  start: "启动循环",
  snapshotLabel: (name, workerType, id) => `${name} · ${workerType} · 模板 ${id}`,
};

describe("LoopRuntimeDialog", () => {
  it("renders above Blockly toolbox and widget layers", () => {
    render(
      <LoopRuntimeDialog
        loading={false}
        messages={messages}
        open
        running={false}
        snapshots={snapshots}
        onOpenChange={vi.fn()}
        onRetry={vi.fn()}
        onRun={vi.fn()}
      />,
    );

    expect(document.querySelector("[data-dialog-overlay]")).toHaveClass("z-[100001]");
  });

  it("submits the explicitly selected runtime snapshot", () => {
    const onRun = vi.fn();
    render(
      <LoopRuntimeDialog
        loading={false}
        messages={messages}
        open
        running={false}
        snapshots={snapshots}
        onOpenChange={vi.fn()}
        onRetry={vi.fn()}
        onRun={onRun}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "选择运行环境" }));
    fireEvent.click(screen.getByRole("option", { name: /审查环境/ }));
    fireEvent.click(screen.getByRole("button", { name: "启动循环" }));

    expect(onRun).toHaveBeenCalledWith("9007199254740993");
  });

  it("explains the empty state and disables execution", () => {
    render(
      <LoopRuntimeDialog
        loading={false}
        messages={messages}
        open
        running={false}
        snapshots={[]}
        onOpenChange={vi.fn()}
        onRetry={vi.fn()}
        onRun={vi.fn()}
      />,
    );

    expect(screen.getByText("当前组织没有可用的运行环境")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "启动循环" })).toBeDisabled();
  });

  it("shows runtime loading without reporting a false empty state", () => {
    render(
      <LoopRuntimeDialog
        loading
        messages={messages}
        open
        running={false}
        snapshots={[]}
        onOpenChange={vi.fn()}
        onRetry={vi.fn()}
        onRun={vi.fn()}
      />,
    );

    expect(screen.getByText("正在加载运行环境")).toBeInTheDocument();
    expect(screen.queryByText("当前组织没有可用的运行环境")).not.toBeInTheDocument();
  });

  it("shows a retryable load error without reporting an empty organization", () => {
    const onRetry = vi.fn();
    render(
      <LoopRuntimeDialog
        error="运行环境加载失败，请稍后重试"
        loading={false}
        messages={messages}
        open
        running={false}
        snapshots={[]}
        onOpenChange={vi.fn()}
        onRetry={onRetry}
        onRun={vi.fn()}
      />,
    );

    expect(screen.getByText("运行环境加载失败，请稍后重试")).toBeInTheDocument();
    expect(screen.queryByText("当前组织没有可用的运行环境")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "重新加载" }));
    expect(onRetry).toHaveBeenCalledOnce();
  });
});
