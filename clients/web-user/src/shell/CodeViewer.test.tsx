import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { useFileContent } from "@/hooks/useFileContent";
import { CodeViewer } from "./CodeViewer";

vi.mock("@/hooks/usePermissions", () => ({ useCanEdit: vi.fn() }));
vi.mock("@/components/ai-elements/code-block", () => ({
  highlightCode: vi.fn(() => null),
}));
vi.mock("./MarkdownRichTextViewer", () => ({ MarkdownRichTextViewer: () => null }));
vi.mock("./MonacoCodeEditor", () => ({
  MonacoCodeEditor: () => <div data-testid="monaco-editor-stub" />,
}));

import * as permissions from "@/hooks/usePermissions";

function makeFileQuery(content: string): ReturnType<typeof useFileContent> {
  return {
    data: { content, encoding: "utf-8", truncated: false },
    isLoading: false,
    isError: false,
    isSuccess: true,
    error: null,
  } as unknown as ReturnType<typeof useFileContent>;
}

function makeVideoQuery(): ReturnType<typeof useFileContent> {
  return {
    data: {
      content: "AAAA",
      content_type: "video/mp4",
      encoding: "base64",
      truncated: false,
    },
    isLoading: false,
    isError: false,
    isSuccess: true,
    error: null,
  } as unknown as ReturnType<typeof useFileContent>;
}

const searchInputRef = { current: null };

function renderSource(content: string, panelOpen = true, path = "notes.md") {
  return render(
    <CodeViewer
      conversationId="conv_1"
      path={path}
      fileQuery={makeFileQuery(content)}
      comments={[]}
      activeSelection={null}
      onSetActiveSelection={() => {}}
      panelOpen={panelOpen}
      searchOpen={false}
      setSearchOpen={() => {}}
      searchInputRef={searchInputRef}
      viewMode="source"
    />,
  );
}

function SearchModeHarness({ content }: { content: string }) {
  const [viewMode, setViewMode] = useState<"preview" | "source">("source");
  const [searchOpen, setSearchOpen] = useState(true);
  return (
    <>
      <button type="button" onClick={() => setViewMode("preview")}>
        Preview
      </button>
      <button type="button" onClick={() => setViewMode("source")}>
        Source
      </button>
      <CodeViewer
        conversationId="conv_1"
        path="notes.md"
        fileQuery={makeFileQuery(content)}
        comments={[]}
        activeSelection={null}
        onSetActiveSelection={() => {}}
        panelOpen
        searchOpen={searchOpen}
        setSearchOpen={setSearchOpen}
        searchInputRef={searchInputRef}
        viewMode={viewMode}
      />
    </>
  );
}

function fireCopyEvent(): ReturnType<typeof vi.fn> {
  const setData = vi.fn();
  const event = new Event("copy", { bubbles: true, cancelable: true });
  Object.defineProperty(event, "clipboardData", {
    value: { setData, getData: vi.fn() },
    writable: false,
  });
  document.dispatchEvent(event);
  return setData;
}

beforeEach(() => {
  vi.mocked(permissions.useCanEdit).mockReturnValue(true);
});

afterEach(cleanup);

describe("CodeViewer source keyboard behavior", () => {
  it("copies raw source after Cmd+A", () => {
    const content = "const x = 1;\nconst y = 2;\nconst z = 3;";
    renderSource(content);
    fireEvent.keyDown(window, { key: "a", metaKey: true });
    expect(fireCopyEvent()).toHaveBeenCalledWith("text/plain", content);
  });

  it("copies raw source after Ctrl+A", () => {
    const content = "line1\nline2";
    renderSource(content);
    fireEvent.keyDown(window, { key: "a", ctrlKey: true });
    expect(fireCopyEvent()).toHaveBeenCalledWith("text/plain", content);
  });

  it("preserves embedded newlines", () => {
    const content = "function foo() {\n  return 42;\n}\n";
    renderSource(content);
    fireEvent.keyDown(window, { key: "a", metaKey: true });
    expect(fireCopyEvent()).toHaveBeenCalledWith("text/plain", content);
  });

  it("does not intercept copy without select-all", () => {
    renderSource("line1\nline2");
    expect(fireCopyEvent()).not.toHaveBeenCalled();
  });

  it("clears select-all interception on mousedown", () => {
    renderSource("line1\nline2");
    fireEvent.keyDown(window, { key: "a", metaKey: true });
    fireEvent.mouseDown(document.body);
    expect(fireCopyEvent()).not.toHaveBeenCalled();
  });

  it("does not intercept select-all while an input has focus", () => {
    renderSource("line1\nline2");
    const input = document.createElement("input");
    document.body.appendChild(input);
    input.focus();
    fireEvent.keyDown(window, { key: "a", metaKey: true });
    expect(fireCopyEvent()).not.toHaveBeenCalled();
    input.remove();
  });

  it("does not register shortcuts while the panel is closed", () => {
    renderSource("line1\nline2", false);
    fireEvent.keyDown(window, { key: "a", metaKey: true });
    expect(fireCopyEvent()).not.toHaveBeenCalled();
  });
});

describe("CodeViewer editor routing", () => {
  it("routes non-markdown files to Monaco", async () => {
    renderSource("const x = 1;", true, "src/index.ts");
    expect(await screen.findByTestId("monaco-editor-stub")).toBeDefined();
  });

  it("keeps markdown source on the Shiki path", () => {
    renderSource("# heading");
    expect(screen.queryByTestId("monaco-editor-stub")).toBeNull();
  });
});

describe("CodeViewer source search state", () => {
  it("matches the entered query without trimming it", () => {
    render(<SearchModeHarness content="foo" />);
    fireEvent.change(screen.getByPlaceholderText("Find…"), {
      target: { value: " foo " },
    });
    expect(screen.getByText("No results")).toBeDefined();
  });

  it("preserves the query across preview and source modes", () => {
    render(<SearchModeHarness content="needle" />);
    fireEvent.change(screen.getByPlaceholderText("Find…"), {
      target: { value: "needle" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Preview" }));
    fireEvent.click(screen.getByRole("button", { name: "Source" }));
    expect(screen.getByPlaceholderText("Find…")).toHaveValue("needle");
  });
});

describe("CodeViewer video rendering", () => {
  beforeEach(() => {
    vi.stubGlobal("URL", {
      createObjectURL: vi.fn(() => "blob:workspace-video"),
      revokeObjectURL: vi.fn(),
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("routes an MP4 filesystem response to the video viewer", async () => {
    render(
      <CodeViewer
        conversationId="conv_1"
        path="output/clip.mp4"
        fileQuery={makeVideoQuery()}
        comments={[]}
        activeSelection={null}
        onSetActiveSelection={() => {}}
        panelOpen={true}
        searchOpen={false}
        setSearchOpen={() => {}}
        searchInputRef={searchInputRef}
        viewMode="source"
      />,
    );

    expect(await screen.findByLabelText("clip.mp4")).toHaveAttribute("src", "blob:workspace-video");
    expect(screen.queryByText(/binary file/i)).toBeNull();
  });
});
