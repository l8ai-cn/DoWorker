import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  load: vi.fn(),
}));

vi.mock("@/lib/api/podWorkspaceArtifactApi", () => ({
  listPodWorkspaceArtifacts: mocks.load,
}));

vi.mock("./webAgentWorkbenchWorkspaceArtifacts", () => ({
  prepareWorkspaceArtifacts: (artifacts: unknown) => artifacts,
}));

import { usePodWorkspaceArtifacts } from "./usePodWorkspaceArtifacts";

describe("usePodWorkspaceArtifacts", () => {
  beforeEach(() => {
    mocks.load.mockReset();
  });

  it("hides a previous Pod's artifacts while its replacement is loading", async () => {
    mocks.load
      .mockResolvedValueOnce([{ artifactId: "workspace:pod-a.png" }])
      .mockResolvedValueOnce([{ artifactId: "workspace:pod-b.png" }]);

    const { result, rerender } = renderHook(
      ({ podKey, enabled }) => usePodWorkspaceArtifacts(podKey, enabled),
      { initialProps: { enabled: true, podKey: "pod-a" } },
    );

    await waitFor(() => {
      expect(result.current.artifacts).toHaveLength(1);
    });
    rerender({ enabled: true, podKey: "pod-b" });

    expect(result.current.artifacts).toEqual([]);
    await waitFor(() => {
      expect(result.current.artifacts).toEqual([
        { artifactId: "workspace:pod-b.png" },
      ]);
    });
  });

  it("does not load artifacts while the Pod is unreadable", () => {
    renderHook(() => usePodWorkspaceArtifacts("pod-a", false));

    expect(mocks.load).not.toHaveBeenCalled();
  });
});
