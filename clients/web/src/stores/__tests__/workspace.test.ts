import { describe, it, expect, beforeEach } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useWorkspaceStore } from "../workspace";

describe("Workspace Store", () => {
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

  describe("initial state", () => {
    it("should have default values", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      expect(result.current.panes).toEqual([]);
      expect(result.current.activePane).toBeNull();
      expect(result.current.splitTree).toBeNull();
      expect(result.current.mobileActiveIndex).toBe(0);
      expect(result.current.terminalFontSize).toBe(14);
    });
  });

  describe("panes management", () => {
    it("should add a new pane", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      let paneId: string;
      act(() => {
        paneId = result.current.addPane("pod-123");
      });

      expect(result.current.panes).toHaveLength(1);
      expect(result.current.panes[0].podKey).toBe("pod-123");
      expect(result.current.activePane).toBe(paneId!);
    });

    it("should add a new pane with podKey", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-123");
      });

      expect(result.current.panes[0].podKey).toBe("pod-123");
    });

    it("should create a split tree leaf for first pane", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-123");
      });

      expect(result.current.splitTree).not.toBeNull();
      expect(result.current.splitTree!.type).toBe("leaf");
    });

    it("should create a split node when adding second pane", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-1");
        result.current.addPane("pod-2");
      });

      expect(result.current.splitTree).not.toBeNull();
      expect(result.current.splitTree!.type).toBe("split");
      if (result.current.splitTree!.type === "split") {
        expect(result.current.splitTree!.direction).toBe("horizontal");
        expect(result.current.splitTree!.children).toHaveLength(2);
      }
    });

    it("should add third pane to same group with even sizes", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-1");
        result.current.addPane("pod-2");
        result.current.addPane("pod-3");
      });

      expect(result.current.panes).toHaveLength(3);
      const tree = result.current.splitTree;
      expect(tree).not.toBeNull();
      expect(tree!.type).toBe("split");
      if (tree!.type === "split") {
        expect(tree!.children).toHaveLength(3);
        const evenSize = 100 / 3;
        tree!.sizes.forEach((s) => expect(s).toBeCloseTo(evenSize, 1));
      }
    });

    it("should return existing pane id if pod already open", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      let firstId: string;
      let secondId: string;
      act(() => {
        firstId = result.current.addPane("pod-123");
        secondId = result.current.addPane("pod-123");
      });

      expect(firstId!).toBe(secondId!);
      expect(result.current.panes).toHaveLength(1);
    });

    it("should replace persisted panes when opening a deep-linked pod", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-1");
        result.current.addPane("pod-2");
        result.current.openDeepLinkedPane("pod-3");
      });

      expect(result.current.panes).toEqual([
        expect.objectContaining({ podKey: "pod-3" }),
      ]);
      expect(result.current.activePane).toBe(result.current.panes[0].id);
      expect(result.current.mobileActiveIndex).toBe(0);
      expect(result.current.splitTree).toMatchObject({
        type: "leaf",
        paneId: result.current.panes[0].id,
      });
    });

    it("should retain the target pane id while collapsing a deep-linked layout", () => {
      const { result } = renderHook(() => useWorkspaceStore());
      let targetPaneId = "";

      act(() => {
        result.current.addPane("pod-1");
        targetPaneId = result.current.addPane("pod-2");
        result.current.openDeepLinkedPane("pod-2");
      });

      expect(result.current.panes).toEqual([
        { id: targetPaneId, podKey: "pod-2" },
      ]);
      expect(result.current.activePane).toBe(targetPaneId);
    });

    it("should remove a pane", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      let paneId: string;
      act(() => {
        paneId = result.current.addPane("pod-123");
      });

      expect(result.current.panes).toHaveLength(1);

      act(() => {
        result.current.removePane(paneId!);
      });

      expect(result.current.panes).toHaveLength(0);
      expect(result.current.activePane).toBeNull();
      expect(result.current.splitTree).toBeNull();
    });

    it("should set next pane as active when active pane is removed", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      let firstId: string;
      act(() => {
        firstId = result.current.addPane("pod-1");
        result.current.addPane("pod-2");
      });

      expect(result.current.panes).toHaveLength(2);
      expect(result.current.activePane).toBe(result.current.panes[1].id);

      act(() => {
        result.current.removePane(result.current.panes[1].id);
      });

      expect(result.current.panes).toHaveLength(1);
      expect(result.current.activePane).toBe(firstId!);
    });

    it("should clear all panes", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-1");
        result.current.addPane("pod-2");
        result.current.addPane("pod-3");
      });

      expect(result.current.panes).toHaveLength(3);

      act(() => {
        result.current.clearAllPanes();
      });

      expect(result.current.panes).toHaveLength(0);
      expect(result.current.activePane).toBeNull();
      expect(result.current.mobileActiveIndex).toBe(0);
      expect(result.current.splitTree).toBeNull();
    });
  });

  describe("active pane", () => {
    it("should set active pane", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      let firstId: string;
      act(() => {
        firstId = result.current.addPane("pod-1");
        result.current.addPane("pod-2");
      });

      act(() => {
        result.current.setActivePane(firstId!);
      });

      expect(result.current.activePane).toBe(firstId!);
    });

    it("should set active pane to null", () => {
      const { result } = renderHook(() => useWorkspaceStore());

      act(() => {
        result.current.addPane("pod-1");
        result.current.setActivePane(null);
      });

      expect(result.current.activePane).toBeNull();
    });
  });

});

// NOTE: Relay Connection Pool tests live in relayConnection.test.ts.
