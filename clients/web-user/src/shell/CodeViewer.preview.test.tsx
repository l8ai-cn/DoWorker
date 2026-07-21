import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "@agent-cloud/agent-ui";
import { ImageLightboxProvider } from "@/components/ImageLightbox";
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

const PNG_BASE64 =
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==";
const searchInputRef = { current: null };

function makeFileQuery(
  content: string,
  truncated = false,
  contentType?: string,
): ReturnType<typeof useFileContent> {
  return {
    data: {
      content,
      encoding: contentType ? "base64" : "utf-8",
      content_type: contentType,
      truncated,
    },
    isLoading: false,
    isError: false,
    isSuccess: true,
    error: null,
  } as unknown as ReturnType<typeof useFileContent>;
}

function renderPreview(content: string, path: string, truncated = false) {
  return render(
    <CodeViewer
      conversationId="conv_1"
      path={path}
      fileQuery={makeFileQuery(content, truncated)}
      comments={[]}
      activeSelection={null}
      onSetActiveSelection={() => {}}
      panelOpen
      searchOpen={false}
      setSearchOpen={() => {}}
      searchInputRef={searchInputRef}
      viewMode="preview"
    />,
  );
}

function renderImage(path = "logo.png", truncated = false) {
  const viewer = (
    <CodeViewer
      conversationId="conv_1"
      path={path}
      fileQuery={makeFileQuery(PNG_BASE64, truncated, "image/png")}
      comments={[]}
      activeSelection={null}
      onSetActiveSelection={() => {}}
      panelOpen
      searchOpen={false}
      setSearchOpen={() => {}}
      searchInputRef={searchInputRef}
      viewMode="source"
    />
  );
  return render(<ImageLightboxProvider>{viewer}</ImageLightboxProvider>);
}

beforeEach(() => {
  vi.mocked(permissions.useCanEdit).mockReturnValue(true);
});

afterEach(cleanup);

describe("CodeViewer truncated preview", () => {
  it("shows the warning for truncated markdown", () => {
    renderPreview("# big file", "notes.md", true);
    expect(screen.getByText(/too large to load fully/)).toBeDefined();
  });

  it("does not show the warning for complete markdown", () => {
    renderPreview("# full file", "notes.md");
    expect(screen.queryByText(/too large to load fully/)).toBeNull();
  });
});

describe("CodeViewer markdown preview", () => {
  it("renders headings", () => {
    const { container } = renderPreview("# Title\n\n## Subtitle", "doc.md");
    expect(container.querySelector("h1")?.textContent).toBe("Title");
    expect(container.querySelector("h2")?.textContent).toBe("Subtitle");
  });

  it("renders bullet and ordered lists", () => {
    const { container } = renderPreview("- one\n- two\n\n1. first\n2. second", "doc.md");
    expect(container.querySelectorAll("ul li")).toHaveLength(2);
    expect(container.querySelectorAll("ol li")).toHaveLength(2);
  });

  it("renders GFM tables", () => {
    const { container } = renderPreview("| A | B |\n| - | - |\n| 1 | 2 |", "doc.md");
    expect(container.querySelectorAll("th")).toHaveLength(2);
    expect(container.querySelectorAll("tbody td")).toHaveLength(2);
  });

  it("renders fenced code blocks", () => {
    const { container } = renderPreview("```js\nconst x = 1;\n```", "doc.md");
    expect(container.querySelector("pre code")?.textContent).toContain("const x = 1;");
  });

  it("renders blockquotes", () => {
    const { container } = renderPreview("> quoted text", "doc.md");
    expect(container.querySelector("blockquote")?.textContent).toContain("quoted text");
  });

  it("renders GFM task list state", () => {
    const { container } = renderPreview("- [x] done\n- [ ] todo", "doc.md");
    const boxes = container.querySelectorAll<HTMLInputElement>('input[type="checkbox"]');
    expect(boxes).toHaveLength(2);
    expect(boxes[0].checked).toBe(true);
    expect(boxes[1].checked).toBe(false);
  });

  it("renders emoji shortcodes", () => {
    const { container } = renderPreview("Ship it :tada: :rocket:", "doc.md");
    expect(container.textContent).toContain("🎉");
    expect(container.textContent).toContain("🚀");
  });

  it("renders supported raw HTML", () => {
    const { container } = renderPreview(
      "<details><summary>More</summary>Hidden</details>\n\nH<sub>2</sub>O\n\npress <kbd>Enter</kbd>",
      "doc.md",
    );
    expect(container.querySelector("details summary")?.textContent).toBe("More");
    expect(container.querySelector("sub")?.textContent).toBe("2");
    expect(container.querySelector("kbd")?.textContent).toBe("Enter");
  });

  it("sanitizes executable raw HTML", () => {
    const { container } = renderPreview(
      '<script>bad()</script>\n<img src="x" onerror="bad()" alt="x">\n<a href="javascript:bad()">click</a>',
      "doc.md",
    );
    expect(container.querySelector("script")).toBeNull();
    expect(container.querySelector("img")).toBeNull();
    expect(container.innerHTML).not.toContain("onerror");
    expect(container.querySelector("a")?.getAttribute("href")).toBeNull();
  });

  it("renders GitHub alerts as typed callouts", () => {
    const { container } = renderPreview(
      "> [!NOTE]\n> Useful information.\n\n> [!WARNING]\n> Careful here.",
      "doc.md",
    );
    expect(container.querySelector(".markdown-alert-note")).not.toBeNull();
    expect(container.querySelector(".markdown-alert-warning")).not.toBeNull();
    expect(container.textContent).not.toContain("[!NOTE]");
  });

  it("keeps explicit image dimensions", () => {
    const { container } = renderPreview(
      '<img src="data:image/png;base64,AA==" alt="logo" width="200" height="100">',
      "doc.md",
    );
    const image = container.querySelector<HTMLImageElement>('img[alt="logo"]');
    expect(image?.style.width).toBe("200px");
    expect(image?.style.height).toBe("100px");
  });
});

describe("CodeViewer HTML preview", () => {
  it("uses the shared non-executable sandbox profile", () => {
    const content = "<script>window.pwned=true</script><form></form>";
    const { container } = renderPreview(content, "page.html");
    const frame = container.querySelector('iframe[title="HTML preview"]');
    expect(frame?.getAttribute("sandbox")).toBe(STATIC_HTML_SANDBOX);
    expect(frame?.getAttribute("referrerpolicy")).toBe(STATIC_HTML_REFERRER_POLICY);
    expect(frame?.getAttribute("srcdoc")).toBe(staticHtmlDocument(content));
    expect(frame?.getAttribute("srcdoc")).not.toContain("omni-html-comment");
  });
});

describe("CodeViewer image preview", () => {
  let createdBlob: Blob | null;

  beforeEach(() => {
    createdBlob = null;
    vi.stubGlobal("URL", {
      createObjectURL: vi.fn((blob: Blob) => {
        createdBlob = blob;
        return "blob:mock-object-url";
      }),
      revokeObjectURL: vi.fn(),
    });
  });

  afterEach(() => vi.unstubAllGlobals());

  it("renders image bytes through a blob URL", async () => {
    renderImage("assets/logo.png");
    const image = (await screen.findByAltText("logo.png")) as HTMLImageElement;
    expect(image.getAttribute("src")).toBe("blob:mock-object-url");
    expect(screen.queryByText(/binary file/i)).toBeNull();
    expect(createdBlob?.type).toBe("image/png");
    expect(createdBlob?.size).toBe(atob(PNG_BASE64).length);
  });

  it("shows the truncated warning", () => {
    renderImage("logo.png", true);
    expect(screen.getByText(/too large to load fully/)).toBeDefined();
  });

  it("routes by MIME type over extension", async () => {
    renderImage("data.txt");
    expect(await screen.findByAltText("data.txt")).toBeDefined();
  });

  it("revokes the blob URL when the image viewer unmounts", async () => {
    const view = renderImage();
    await screen.findByAltText("logo.png");
    view.unmount();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock-object-url");
  });

  it("opens the shared lightbox", async () => {
    renderImage("assets/logo.png");
    fireEvent.click(await screen.findByAltText("logo.png"));
    expect(await screen.findByRole("dialog")).toBeDefined();
    expect(screen.getByLabelText("Zoom in")).toBeDefined();
    expect(screen.getByLabelText("Zoom out")).toBeDefined();
  });
});
