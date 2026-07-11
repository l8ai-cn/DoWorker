import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { Pod } from "@/stores/pod";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

import { SidebarPodContextMenu } from "../SidebarPodContextMenu";

const pod = {
  pod_key: "pod-123",
  status: "terminated",
  perpetual: false,
} as Pod;

describe("SidebarPodContextMenu", () => {
  it("offers wake from the worker right-click menu", () => {
    const onWake = vi.fn();
    render(
      <SidebarPodContextMenu
        pod={pod}
        onRename={vi.fn()}
        onShare={vi.fn()}
        onOpenMobile={vi.fn()}
        onDelete={vi.fn()}
        onTerminate={vi.fn()}
        onTogglePerpetual={vi.fn()}
        onWake={onWake}
      >
        <button type="button">Worker</button>
      </SidebarPodContextMenu>,
    );

    fireEvent.contextMenu(screen.getByRole("button", { name: "Worker" }));
    fireEvent.click(screen.getByText("contextMenu.wake"));

    expect(onWake).toHaveBeenCalledTimes(1);
  });
});
