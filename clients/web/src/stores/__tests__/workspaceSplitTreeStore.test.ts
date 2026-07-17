import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { useWorkspaceStore } from "../workspace";

describe("Workspace split tree store", () => {
  beforeEach(() => {
    localStorage.clear();
    useWorkspaceStore.setState({
      panes: [],
      activePane: null,
      splitTree: null,
      mobileActiveIndex: 0,
      terminalFontSize: 14,
      _hasHydrated: false,
    });
  });

  it("splits a pane horizontally", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let paneId = "";

    act(() => {
      paneId = result.current.addPane("pod-1");
      result.current.splitPane(paneId, "horizontal", "pod-2");
    });

    expect(result.current.splitTree).toMatchObject({
      type: "split",
      direction: "horizontal",
    });
    expect(result.current.panes.map((pane) => pane.podKey)).toEqual([
      "pod-1",
      "pod-2",
    ]);
    expect(result.current.activePane).toBe(result.current.panes[1].id);
  });

  it("splits a pane vertically", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let paneId = "";

    act(() => {
      paneId = result.current.addPane("pod-1");
      result.current.splitPane(paneId, "vertical", "pod-3");
    });

    expect(result.current.splitTree).toMatchObject({
      type: "split",
      direction: "vertical",
    });
    expect(result.current.panes[1].podKey).toBe("pod-3");
  });

  it("bubbles a same-direction split into its parent", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let firstPaneId = "";

    act(() => {
      firstPaneId = result.current.addPane("pod-1");
      result.current.addPane("pod-2");
      result.current.splitPane(firstPaneId, "horizontal", "pod-3");
    });

    const tree = result.current.splitTree;
    expect(tree?.type).toBe("split");
    if (tree?.type === "split") {
      expect(tree.children).toHaveLength(3);
      expect(tree.direction).toBe("horizontal");
      expect(tree.sizes[0]).toBeCloseTo(25, 1);
      expect(tree.sizes[1]).toBeCloseTo(25, 1);
      expect(tree.sizes[2]).toBeCloseTo(50, 1);
    }
  });

  it("nests a cross-direction split", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let firstPaneId = "";

    act(() => {
      firstPaneId = result.current.addPane("pod-1");
      result.current.addPane("pod-2");
      result.current.splitPane(firstPaneId, "vertical", "pod-3");
    });

    const tree = result.current.splitTree;
    expect(tree?.type).toBe("split");
    if (tree?.type === "split") {
      expect(tree.children).toHaveLength(2);
      expect(tree.direction).toBe("horizontal");
      expect(tree.children[0]).toMatchObject({
        type: "split",
        direction: "vertical",
        sizes: [50, 50],
      });
    }
  });

  it("normalizes sizes after removing a child", () => {
    const { result } = renderHook(() => useWorkspaceStore());

    act(() => {
      result.current.addPane("pod-1");
      result.current.addPane("pod-2");
      result.current.addPane("pod-3");
    });
    const tree = result.current.splitTree;
    if (tree?.type === "split") {
      act(() => result.current.updateSplitSizes(tree.id, [50, 30, 20]));
    }
    act(() => result.current.removePane(result.current.panes[1].id));

    const updated = result.current.splitTree;
    expect(updated?.type).toBe("split");
    if (updated?.type === "split") {
      expect(updated.children).toHaveLength(2);
      expect(updated.sizes[0]).toBeCloseTo(71.4, 0);
      expect(updated.sizes[1]).toBeCloseTo(28.6, 0);
    }
  });

  it("updates split sizes", () => {
    const { result } = renderHook(() => useWorkspaceStore());

    act(() => {
      result.current.addPane("pod-1");
      result.current.addPane("pod-2");
    });
    const tree = result.current.splitTree;
    expect(tree?.type).toBe("split");
    act(() => result.current.updateSplitSizes(tree!.id, [30, 70]));

    if (result.current.splitTree?.type === "split") {
      expect(result.current.splitTree.sizes).toEqual([30, 70]);
    }
  });

  it("creates a split from a root leaf", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let paneId = "";

    act(() => {
      paneId = result.current.addPane("pod-1");
    });
    expect(result.current.splitTree?.type).toBe("leaf");
    act(() => result.current.splitPane(paneId, "vertical", "pod-2"));

    expect(result.current.splitTree).toMatchObject({
      type: "split",
      direction: "vertical",
      sizes: [50, 50],
    });
  });

  it("grows a same-direction group across repeated splits", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let paneId = "";

    act(() => {
      paneId = result.current.addPane("pod-1");
    });
    for (let index = 2; index <= 5; index += 1) {
      act(() => {
        result.current.splitPane(paneId, "horizontal", `pod-${index}`);
      });
    }

    expect(result.current.panes).toHaveLength(5);
    const tree = result.current.splitTree;
    expect(tree?.type).toBe("split");
    if (tree?.type === "split") {
      expect(tree.children).toHaveLength(5);
      tree.sizes.forEach((size) => expect(size).toBeGreaterThan(0));
    }
  });

  it("collapses a two-child split after removal", () => {
    const { result } = renderHook(() => useWorkspaceStore());
    let firstPaneId = "";

    act(() => {
      firstPaneId = result.current.addPane("pod-1");
      result.current.addPane("pod-2");
    });
    act(() => {
      result.current.removePane(result.current.panes[1].id);
    });

    expect(result.current.splitTree).toMatchObject({
      type: "leaf",
      paneId: firstPaneId,
    });
  });

  it("wraps the root when no pane is active", () => {
    const { result } = renderHook(() => useWorkspaceStore());

    act(() => {
      result.current.addPane("pod-1");
      result.current.setActivePane(null);
      result.current.addPane("pod-2");
    });

    expect(result.current.panes).toHaveLength(2);
    expect(result.current.splitTree).toMatchObject({
      type: "split",
      direction: "horizontal",
    });
  });

  it("removes a pane by pod key", () => {
    const { result } = renderHook(() => useWorkspaceStore());

    act(() => {
      result.current.addPane("pod-1");
      result.current.addPane("pod-2");
      result.current.removePaneByPodKey("pod-1");
    });

    expect(result.current.panes).toEqual([
      expect.objectContaining({ podKey: "pod-2" }),
    ]);
  });
});
