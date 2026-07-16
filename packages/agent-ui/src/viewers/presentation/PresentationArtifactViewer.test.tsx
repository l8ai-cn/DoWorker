import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import {
  PRESENTATION_GRANTS,
  PresentationArtifactViewer,
  type PresentationArtifactViewerProps,
} from "./PresentationArtifactViewer";

const slides = [
  {
    imageSrc: "/slides/third.png",
    notes: "第三页备注",
    position: 30,
    slideId: "slide-c",
    title: "路线图",
  },
  {
    imageSrc: "/slides/first.png",
    notes: "先说明业务目标。",
    position: 10,
    slideId: "slide-a",
    title: "项目概览",
  },
  {
    imageSrc: "/slides/second.png",
    position: 20,
    slideId: "slide-b",
    title: "核心能力",
  },
] as const;

const versions = [
  { id: "v1", label: "初稿", revision: 11n },
  { id: "v2", label: "评审稿", revision: 12n },
] as const;

function createProps(
  overrides: Partial<PresentationArtifactViewerProps> = {},
): PresentationArtifactViewerProps {
  return {
    actionSchemaVersion: "1",
    artifactId: "artifact-deck",
    baseRevision: 12n,
    grants: Object.values(PRESENTATION_GRANTS),
    onAction: vi.fn(),
    onSelectVersion: vi.fn(),
    representationId: "rendered-slides",
    selectedVersionId: "v2",
    slides,
    versions,
    ...overrides,
  };
}

describe("PresentationArtifactViewer", () => {
  it("按 position 和 slideId 稳定排列缩略图，并同步当前页、页码和讲者备注", () => {
    render(
      <PresentationArtifactViewer
        {...createProps({ initialSlideId: "slide-b" })}
      />,
    );

    const thumbnails = screen.getAllByRole("button", { name: /转到第/ });
    expect(
      thumbnails.map((button) => button.getAttribute("aria-label")),
    ).toEqual([
      "转到第 1 页：项目概览",
      "转到第 2 页：核心能力",
      "转到第 3 页：路线图",
    ]);
    expect(screen.getByText("第 2 / 3 页")).toBeVisible();
    expect(
      screen.getByRole("img", { name: "第 2 页：核心能力" }),
    ).toHaveAttribute("src", "/slides/second.png");
    expect(screen.getByText("暂无讲者备注")).toBeVisible();

    fireEvent.click(
      screen.getByRole("button", { name: "转到第 1 页：项目概览" }),
    );
    expect(screen.getByText("第 1 / 3 页")).toBeVisible();
    expect(screen.getByText("先说明业务目标。")).toBeVisible();
  });

  it("支持缩放和恢复适应窗口", () => {
    render(<PresentationArtifactViewer {...createProps()} />);

    const image = screen.getByTestId("presentation-slide-image");
    const zoom = screen.getByRole("slider", { name: "缩放比例" });
    expect(image).toHaveStyle({ transform: "scale(1)" });
    expect(screen.getByText("100%")).toBeVisible();

    fireEvent.change(zoom, { target: { value: "150" } });
    expect(image).toHaveStyle({ transform: "scale(1.5)" });
    expect(screen.getByText("150%")).toBeVisible();
    expect(screen.getByRole("button", { name: "适应窗口" })).toHaveAttribute(
      "aria-pressed",
      "false",
    );

    fireEvent.click(screen.getByRole("button", { name: "适应窗口" }));
    expect(image).toHaveStyle({ transform: "scale(1)" });
    expect(screen.getByRole("button", { name: "适应窗口" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
  });

  it("预览舞台受父容器宽度约束，窄屏不按最小高度撑宽", () => {
    render(<PresentationArtifactViewer {...createProps()} />);

    expect(
      screen.getByTestId("presentation-slide-image").parentElement,
    ).toHaveClass("min-w-0", "w-full");
  });

  it("通过宿主回调发起全屏命令", () => {
    const onRequestFullscreen = vi.fn();
    render(
      <PresentationArtifactViewer {...createProps({ onRequestFullscreen })} />,
    );

    fireEvent.click(screen.getByRole("button", { name: "全屏查看" }));
    expect(onRequestFullscreen).toHaveBeenCalledOnce();
  });

  it("宿主未接管时调用原生全屏 API", () => {
    const requestFullscreen = vi.fn().mockResolvedValue(undefined);
    const original = HTMLElement.prototype.requestFullscreen;
    HTMLElement.prototype.requestFullscreen = requestFullscreen;

    render(
      <PresentationArtifactViewer
        {...createProps({ onRequestFullscreen: undefined })}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "全屏查看" }));
    expect(requestFullscreen).toHaveBeenCalledOnce();

    HTMLElement.prototype.requestFullscreen = original;
  });

  it("版本选择由宿主控制，不在本地乐观切换", () => {
    const onSelectVersion = vi.fn();
    render(
      <PresentationArtifactViewer {...createProps({ onSelectVersion })} />,
    );

    const select = screen.getByRole("combobox", { name: "选择演示文稿版本" });
    expect(select).toHaveValue("v2");
    fireEvent.change(select, { target: { value: "v1" } });
    expect(onSelectVersion).toHaveBeenCalledWith("v1");
    expect(select).toHaveValue("v2");
  });

  it("按精确 grant 禁用重生成、替换、重排和导出", () => {
    render(
      <PresentationArtifactViewer
        {...createProps({
          grants: [PRESENTATION_GRANTS.regenerateSlide],
        })}
      />,
    );

    expect(
      screen.getByRole("button", { name: "重新生成当前页" }),
    ).toBeEnabled();
    expect(screen.getByRole("button", { name: "替换当前页" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "上移当前页" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "下移当前页" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "导出演示文稿" })).toBeDisabled();
  });

  it("发出的动作携带 Artifact 身份、基线修订、页面和幂等 ID", () => {
    const onAction = vi.fn();
    const randomUUID = vi
      .spyOn(globalThis.crypto, "randomUUID")
      .mockReturnValueOnce("00000000-0000-4000-8000-000000000001")
      .mockReturnValueOnce("00000000-0000-4000-8000-000000000002")
      .mockReturnValueOnce("00000000-0000-4000-8000-000000000003")
      .mockReturnValueOnce("00000000-0000-4000-8000-000000000004");

    render(
      <PresentationArtifactViewer
        {...createProps({
          initialSlideId: "slide-b",
          onAction,
        })}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "重新生成当前页" }));
    fireEvent.click(screen.getByRole("button", { name: "替换当前页" }));
    fireEvent.click(screen.getByRole("button", { name: "下移当前页" }));
    fireEvent.click(screen.getByRole("button", { name: "导出演示文稿" }));

    expect(onAction).toHaveBeenNthCalledWith(1, {
      actionSchemaVersion: "1",
      actionType: PRESENTATION_GRANTS.regenerateSlide,
      artifactId: "artifact-deck",
      baseRevision: 12n,
      commandId: "00000000-0000-4000-8000-000000000001",
      payload: { slideId: "slide-b" },
      representationId: "rendered-slides",
    });
    expect(onAction).toHaveBeenNthCalledWith(2, {
      actionSchemaVersion: "1",
      actionType: PRESENTATION_GRANTS.replaceSlide,
      artifactId: "artifact-deck",
      baseRevision: 12n,
      commandId: "00000000-0000-4000-8000-000000000002",
      payload: { slideId: "slide-b" },
      representationId: "rendered-slides",
    });
    expect(onAction).toHaveBeenNthCalledWith(3, {
      actionSchemaVersion: "1",
      actionType: PRESENTATION_GRANTS.reorderSlide,
      artifactId: "artifact-deck",
      baseRevision: 12n,
      commandId: "00000000-0000-4000-8000-000000000003",
      payload: { slideId: "slide-b", targetIndex: 2 },
      representationId: "rendered-slides",
    });
    expect(onAction).toHaveBeenNthCalledWith(4, {
      actionSchemaVersion: "1",
      actionType: PRESENTATION_GRANTS.exportPresentation,
      artifactId: "artifact-deck",
      baseRevision: 12n,
      commandId: "00000000-0000-4000-8000-000000000004",
      payload: { format: "pptx", slideId: "slide-b" },
      representationId: "rendered-slides",
    });
    expect(screen.getByText("修订 12")).toBeVisible();

    randomUUID.mockRestore();
  });

  it("没有可渲染页面时保持空态并禁用页面动作", () => {
    render(<PresentationArtifactViewer {...createProps({ slides: [] })} />);

    expect(screen.getByText("暂无可预览页面")).toBeVisible();
    expect(
      screen.getByRole("button", { name: "重新生成当前页" }),
    ).toBeDisabled();
    expect(screen.getByRole("button", { name: "全屏查看" })).toBeDisabled();
  });
});
