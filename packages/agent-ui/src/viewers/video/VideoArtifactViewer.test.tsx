import { act, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, vi } from "vitest";

import { VideoArtifactViewer } from "./VideoArtifactViewer";

const readyProps = {
  src: "https://media.example.com/final.mp4",
  filename: "产品演示.mp4",
  mimeType: "video/mp4",
  posterSrc: "https://media.example.com/poster.jpg",
  durationSeconds: 65,
  status: "ready" as const,
};

const originalRequestFullscreen = HTMLElement.prototype.requestFullscreen;
const originalExitFullscreen = document.exitFullscreen;

describe("VideoArtifactViewer", () => {
  beforeEach(() => {
    Object.defineProperty(HTMLElement.prototype, "requestFullscreen", {
      configurable: true,
      value: vi.fn().mockResolvedValue(undefined),
    });
    Object.defineProperty(document, "exitFullscreen", {
      configurable: true,
      value: vi.fn().mockResolvedValue(undefined),
    });
  });

  afterEach(() => {
    Object.defineProperty(HTMLElement.prototype, "requestFullscreen", {
      configurable: true,
      value: originalRequestFullscreen,
    });
    Object.defineProperty(document, "exitFullscreen", {
      configurable: true,
      value: originalExitFullscreen,
    });
  });

  it("完成后使用原生视频控件渲染可播放资源和 MIME 类型", () => {
    render(<VideoArtifactViewer {...readyProps} />);

    const video = screen.getByLabelText("视频预览：产品演示.mp4");
    expect(video).toHaveAttribute("controls");
    expect(video).toHaveAttribute("playsinline");
    expect(video).toHaveAttribute("preload", "metadata");
    expect(video).toHaveAttribute("poster", readyProps.posterSrc);
    expect(video).toHaveAttribute("src", readyProps.src);
    expect(screen.getByText("时长 1:05")).toBeVisible();
    expect(screen.queryByRole("progressbar")).not.toBeInTheDocument();
  });

  it("用户模式不显示 MIME 和内部版本角色", () => {
    render(
      <VideoArtifactViewer
        {...readyProps}
        selectedVersionId="playable"
        technicalMetadata={false}
        versions={[
          { id: "playable", label: "playable" },
          { id: "original", label: "original" },
        ]}
      />,
    );

    expect(screen.queryByText("video/mp4")).not.toBeInTheDocument();
    expect(screen.queryByText("playable")).not.toBeInTheDocument();
    expect(screen.getByRole("option", { name: "版本 1" })).toBeVisible();
    expect(screen.getByRole("option", { name: "版本 2" })).toBeVisible();
  });

  it("支持从结果卡片进入和退出全屏预览", async () => {
    const user = userEvent.setup();
    let fullscreenElement: Element | null = null;
    Object.defineProperty(document, "fullscreenElement", {
      configurable: true,
      get: () => fullscreenElement,
    });
    const requestFullscreen = vi
      .spyOn(HTMLElement.prototype, "requestFullscreen")
      .mockImplementation(async function (this: HTMLElement) {
        fullscreenElement = this;
        document.dispatchEvent(new Event("fullscreenchange"));
      });
    const exitFullscreen = vi
      .spyOn(document, "exitFullscreen")
      .mockImplementation(async () => {
        fullscreenElement = null;
        document.dispatchEvent(new Event("fullscreenchange"));
      });
    render(<VideoArtifactViewer {...readyProps} />);

    const enter = screen.getByRole("button", { name: "View video fullscreen" });
    await user.click(enter);
    const exit = screen.getByRole("button", { name: "Exit fullscreen" });
    expect(requestFullscreen).toHaveBeenCalledOnce();
    await user.click(exit);
    expect(exitFullscreen).toHaveBeenCalledOnce();
    expect(
      screen.getByRole("button", { name: "View video fullscreen" }),
    ).toBeVisible();
    requestFullscreen.mockRestore();
    exitFullscreen.mockRestore();
  });

  it("浏览器退出全屏时同步恢复按钮状态", async () => {
    const user = userEvent.setup();
    let fullscreenElement: Element | null = null;
    Object.defineProperty(document, "fullscreenElement", {
      configurable: true,
      get: () => fullscreenElement,
    });
    const requestFullscreen = vi
      .spyOn(HTMLElement.prototype, "requestFullscreen")
      .mockImplementation(async function (this: HTMLElement) {
        fullscreenElement = this;
        document.dispatchEvent(new Event("fullscreenchange"));
      });
    render(<VideoArtifactViewer {...readyProps} />);

    await user.click(
      screen.getByRole("button", { name: "View video fullscreen" }),
    );
    fullscreenElement = null;
    act(() => document.dispatchEvent(new Event("fullscreenchange")));

    expect(
      screen.getByRole("button", { name: "View video fullscreen" }),
    ).toBeVisible();
    requestFullscreen.mockRestore();
  });

  it.each([
    ["queued", "视频已排队，等待生成"],
    ["rendering", "正在渲染视频"],
    ["transcoding", "正在转码视频"],
  ] as const)("状态为 %s 时只显示状态，不渲染可播放视频", (status, label) => {
    render(
      <VideoArtifactViewer
        {...readyProps}
        progress={undefined}
        status={status}
      />,
    );

    expect(screen.getByRole("status")).toHaveTextContent(label);
    expect(screen.getByRole("progressbar")).toHaveAttribute(
      "aria-valuetext",
      "进度未知",
    );
    expect(screen.queryByLabelText("视频预览：产品演示.mp4")).not.toBeInTheDocument();
  });

  it("生成中显示确定进度并把越界值限制在 0 到 100", () => {
    const { rerender } = render(
      <VideoArtifactViewer
        {...readyProps}
        progress={37}
        status="rendering"
      />,
    );

    expect(screen.getByRole("progressbar")).toHaveAttribute(
      "aria-valuenow",
      "37",
    );
    expect(screen.getByText("37%")).toBeVisible();

    rerender(
      <VideoArtifactViewer
        {...readyProps}
        progress={140}
        status="transcoding"
      />,
    );
    expect(screen.getByRole("progressbar")).toHaveAttribute(
      "aria-valuenow",
      "100",
    );
  });

  it("失败时使用告警语义，不渲染视频并禁用下载", () => {
    render(
      <VideoArtifactViewer
        {...readyProps}
        onDownload={vi.fn()}
        status="failed"
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("视频生成失败");
    expect(screen.queryByLabelText("视频预览：产品演示.mp4")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "下载视频：产品演示.mp4" }),
    ).toBeDisabled();
  });

  it("使用受控原生选择器切换版本", () => {
    const onSelectVersion = vi.fn();
    render(
      <VideoArtifactViewer
        {...readyProps}
        onSelectVersion={onSelectVersion}
        selectedVersionId="v2"
        versions={[
          { id: "v1", label: "版本 1" },
          { id: "v2", label: "版本 2" },
        ]}
      />,
    );

    const select = screen.getByRole("combobox", { name: "选择视频版本" });
    expect(select).toHaveValue("v2");
    fireEvent.change(select, { target: { value: "v1" } });
    expect(onSelectVersion).toHaveBeenCalledWith("v1");
  });

  it("版本资源变化时重建播放器以加载新 source", () => {
    const versions = [
      {
        id: "v1",
        filename: "v1.mp4",
        mimeType: "video/mp4",
        src: "https://media.example.com/v1.mp4",
      },
      {
        id: "v2",
        filename: "v2.mp4",
        mimeType: "video/mp4",
        src: "https://media.example.com/v2.mp4",
      },
    ];
    const { rerender } = render(
      <VideoArtifactViewer
        {...readyProps}
        selectedVersionId="v1"
        versions={versions}
      />,
    );
    const first = screen.getByLabelText("视频预览：v1.mp4");

    rerender(
      <VideoArtifactViewer
        {...readyProps}
        selectedVersionId="v2"
        versions={versions}
      />,
    );

    const second = screen.getByLabelText("视频预览：v2.mp4");
    expect(second).not.toBe(first);
    expect(second).toHaveAttribute("src", "https://media.example.com/v2.mp4");
  });

  it("下载命令具有中文名称、正确禁用态并支持键盘触发", async () => {
    const user = userEvent.setup();
    const onDownload = vi.fn();
    const { rerender } = render(
      <VideoArtifactViewer
        {...readyProps}
        onDownload={onDownload}
      />,
    );

    const download = screen.getByRole("button", {
      name: "下载视频：产品演示.mp4",
    });
    expect(download).toBeEnabled();
    await user.tab();
    expect(
      screen.getByRole("button", { name: "View video fullscreen" }),
    ).toHaveFocus();
    await user.tab();
    expect(download).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(onDownload).toHaveBeenCalledOnce();

    rerender(<VideoArtifactViewer {...readyProps} />);
    expect(
      screen.getByRole("button", { name: "下载视频：产品演示.mp4" }),
    ).toBeDisabled();
  });
});
