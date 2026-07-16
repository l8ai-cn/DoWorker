import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SourceSelectionActions } from "./SourceSelectionActions";

const mocks = vi.hoisted(() => ({
  addComposerAttachment: vi.fn(),
  getEmbedRoot: vi.fn<() => HTMLElement | null>(),
}));

vi.mock("@/lib/host", () => ({ getEmbedRoot: mocks.getEmbedRoot }));
vi.mock("@/store/chatStore", () => ({
  useChatStore: {
    getState: () => ({ addComposerAttachment: mocks.addComposerAttachment }),
  },
}));

const anchor = {
  x: 100,
  y: 100,
  start_index: 1,
  end_index: 5,
  anchor_content: "bc\nd",
};

beforeEach(() => {
  mocks.addComposerAttachment.mockReset();
  mocks.getEmbedRoot.mockReset();
});

describe("SourceSelectionActions", () => {
  it("renders into the embed root and confirms a comment selection", () => {
    const embedRoot = document.createElement("div");
    document.body.appendChild(embedRoot);
    mocks.getEmbedRoot.mockReturnValue(embedRoot);
    const onSetActiveSelection = vi.fn();
    const onClose = vi.fn();

    render(
      <SourceSelectionActions
        anchor={anchor}
        canAttachToAgent={false}
        path="src/app.ts"
        rawLines={["abc", "def"]}
        onSetActiveSelection={onSetActiveSelection}
        onClose={onClose}
      />,
    );

    const button = screen.getByRole("button", { name: "Add comment" });
    expect(embedRoot.contains(button)).toBe(true);
    fireEvent.click(button);
    expect(onSetActiveSelection).toHaveBeenCalledWith({
      start_index: 1,
      end_index: 5,
      anchor_content: "bc\nd",
    });
    expect(onClose).toHaveBeenCalled();
    embedRoot.remove();
  });

  it("attaches the selected line range to the agent", () => {
    const onClose = vi.fn();
    mocks.getEmbedRoot.mockReturnValue(null);
    render(
      <SourceSelectionActions
        anchor={anchor}
        canAttachToAgent
        path="src/app.ts"
        rawLines={["abc", "def"]}
        onSetActiveSelection={() => {}}
        onClose={onClose}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Attach to agent" }));
    expect(mocks.addComposerAttachment).toHaveBeenCalledWith({
      path: "src/app.ts",
      isDir: false,
      lineRange: { start: 1, end: 2 },
    });
    expect(onClose).toHaveBeenCalled();
  });
});
