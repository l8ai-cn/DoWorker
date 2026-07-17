import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { PodPreviewInfo } from "@/hooks/usePodPreview";

import {
  buildPreviewWindowUrl,
  getPreviewFrameKey,
  PreviewPanel,
} from "./PreviewPanel";

const usePodPreviewMock = vi.fn();

vi.mock("@/hooks/usePodPreview", () => ({
  PodPreviewError: class extends Error {
    status = 0;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
  usePodPreview: () => usePodPreviewMock(),
  buildPreviewSrc: (info: PodPreviewInfo) => info.session_url,
}));

function previewState(overrides: Partial<PodPreviewInfo>) {
  return {
    data: {
      preview_base_url: "https://pod1.preview.test/preview/pod1/",
      session_url: "https://pod1.preview.test/preview/pod1/__session?token=old",
      expires_at: "2026-07-12T00:00:00Z",
      ...overrides,
    },
    isPending: false,
    isFetching: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof usePodPreviewMock>;
}

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("getPreviewFrameKey", () => {
  it("uses session URL and expiry as refresh drivers and ignores legacy token field", () => {
    const stable = {
      preview_base_url: "https://d/preview/pod1/",
      session_url: "https://d/preview/pod1/__session?token=abc",
      expires_at: "2026-07-12T00:00:00Z",
    };
    const withLegacyToken = {
      ...stable,
      token: "legacy-a",
    } as PodPreviewInfo & { token: string };
    const withDifferentLegacyToken = {
      ...stable,
      token: "legacy-b",
    } as PodPreviewInfo & { token: string };

    expect(getPreviewFrameKey(stable, 0)).toBe(getPreviewFrameKey(withLegacyToken, 0));
    expect(getPreviewFrameKey(stable, 0)).toBe(getPreviewFrameKey(withDifferentLegacyToken, 0));
  });
});

describe("PreviewPanel iframe key behavior", () => {
  it("rebuilds the iframe when the session URL changes", () => {
    usePodPreviewMock.mockReturnValueOnce(previewState({}));
    const { rerender } = render(<PreviewPanel podKey="pod1" />);
    const first = screen.getByTitle("Pod pod1 preview");
    expect(first).toHaveAttribute(
      "src",
      "https://pod1.preview.test/preview/pod1/__session?token=old",
    );
    expect(first).toHaveAttribute(
      "sandbox",
      "allow-scripts allow-same-origin allow-forms allow-downloads",
    );
    expect(first).toHaveAttribute("referrerpolicy", "no-referrer");
    expect(first.getAttribute("allow")).not.toContain("camera");

    usePodPreviewMock.mockReturnValueOnce(
      previewState({
        session_url: "https://pod1.preview.test/preview/pod1/__session?token=new",
      }),
    );
    rerender(<PreviewPanel podKey="pod1" />);
    const second = screen.getByTitle("Pod pod1 preview");
    expect(second).not.toBe(first);
  });

  it("rebuilds the iframe when only the expiry changes", () => {
    usePodPreviewMock.mockReturnValueOnce(previewState({}));
    const { rerender } = render(<PreviewPanel podKey="pod1" />);
    const first = screen.getByTitle("Pod pod1 preview");

    usePodPreviewMock.mockReturnValueOnce(previewState({ expires_at: "2026-07-12T01:00:00Z" }));
    rerender(<PreviewPanel podKey="pod1" />);

    expect(screen.getByTitle("Pod pod1 preview")).not.toBe(first);
  });

  it("rebuilds the iframe on manual refresh when the session URL is unchanged", () => {
    const state = previewState({});
    usePodPreviewMock.mockReturnValue(state);
    render(<PreviewPanel podKey="pod1" />);
    const first = screen.getByTitle("Pod pod1 preview");

    fireEvent.click(screen.getByRole("button", { name: "刷新预览" }));

    expect(screen.getByTitle("Pod pod1 preview")).not.toBe(first);
    expect(state.refetch).toHaveBeenCalledOnce();
  });
});

describe("preview window", () => {
  it("opens the trusted local shell instead of the preview session directly", () => {
    const state = previewState({});
    usePodPreviewMock.mockReturnValue(state);
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    render(<PreviewPanel podKey="pod1" />);

    fireEvent.click(screen.getByRole("button", { name: "在新窗口打开预览" }));

    expect(open).toHaveBeenCalledWith(
      buildPreviewWindowUrl(state.data),
      "_blank",
      "noopener,noreferrer",
    );
    expect(open.mock.calls[0]?.[0]).not.toBe(state.data.session_url);
    open.mockRestore();
  });
});
