import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { Pod } from "@/stores/pod";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

import { SidebarPodActionsMenu } from "../SidebarPodActionsMenu";

const pod = {
  pod_key: "pod-123",
  status: "running",
  perpetual: false,
} as Pod;

describe("SidebarPodActionsMenu", () => {
  it("shows a visible actions button for an active worker", async () => {
    const user = userEvent.setup();
    render(
      <SidebarPodActionsMenu
        pod={pod}
        onRename={vi.fn()}
        onShare={vi.fn()}
        onOpenMobile={vi.fn()}
        onDelete={vi.fn()}
        onTerminate={vi.fn()}
        onTogglePerpetual={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Worker actions" }));

    expect(screen.getByText("contextMenu.rename")).toBeVisible();
    expect(screen.getByText("contextMenu.enablePerpetual")).toBeVisible();
    expect(screen.getByText("contextMenu.terminate")).toBeVisible();
  });
});
