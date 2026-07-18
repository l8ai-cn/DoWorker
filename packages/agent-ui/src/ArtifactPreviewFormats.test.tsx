import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { ActivityTimeline } from "./ActivityTimeline";
import type { AgentArtifactItem, AgentSessionRuntime } from "./contracts";
import { MAX_TEXT_PREVIEW_BYTES } from "./useArtifactBlobUrl";

const getDocument = vi.fn();

vi.mock("pdfjs-dist", () => ({
  GlobalWorkerOptions: { workerSrc: "" },
  getDocument,
}));

describe("generic artifact format previews", () => {
  beforeEach(() => {
    getDocument.mockReturnValue({
      destroy: vi.fn(),
      promise: Promise.resolve({
        cleanup: vi.fn(),
        destroy: vi.fn(),
        getPage: vi.fn(),
        numPages: 0,
      }),
    });
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:preview"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
    });
  });

  it("renders Markdown with headings and GFM tables", async () => {
    renderArtifact(
      artifact("report.md", "text/markdown"),
      runtime("# Report\n\n| A | B |\n|---|---|\n| 1 | 2 |", "text/markdown"),
    );

    expect(await screen.findByRole("heading", { name: "Report" })).toBeVisible();
    expect(screen.getByRole("table")).toBeVisible();
  });

  it("renders CSV as a bounded table", async () => {
    renderArtifact(
      artifact("results.csv", "text/csv"),
      runtime("region,revenue\nNorth,1200\nSouth,900", "text/csv"),
    );

    const table = await screen.findByRole("table", {
      name: "CSV preview for results.csv",
    });
    expect(table).toHaveTextContent("North");
    expect(table).toHaveTextContent("1200");
  });

  it("renders PDF in an inline reader", async () => {
    renderArtifact(
      artifact("report.pdf", "application/pdf"),
      runtime("%PDF", "application/pdf"),
    );

    expect(
      await screen.findByLabelText("PDF preview for report.pdf"),
    ).toBeVisible();
    expect(getDocument).toHaveBeenCalledWith({ url: "blob:preview" });
  });

  it("uses the loaded MIME type instead of trusting a PDF declaration", async () => {
    renderArtifact(
      artifact("report.pdf", "application/pdf"),
      runtime("<script>parent.document.body.dataset.compromised='true'</script>", "text/html"),
    );

    const preview = await screen.findByTitle("Preview report.pdf");
    expect(preview).toHaveAttribute("sandbox", "allow-scripts");
    expect(preview).toHaveAttribute("srcdoc");
    expect(screen.queryByTitle("PDF preview for report.pdf")).not.toBeInTheDocument();
  });

  it("limits text previews before parsing or rendering", async () => {
    const blob = new Blob(["# Report"], { type: "text/markdown" });
    Object.defineProperty(blob, "size", {
      configurable: true,
      value: MAX_TEXT_PREVIEW_BYTES + 1,
    });
    renderArtifact(
      artifact("report.md", "text/markdown"),
      runtimeBlob(blob),
    );

    expect(
      await screen.findByText("Preview limited to the first 2 MiB."),
    ).toBeVisible();
  });

  it("loads audio only after the user asks to preview it", async () => {
    renderArtifact(
      artifact("briefing.mp3", "audio/mpeg"),
      runtime("audio", "audio/mpeg"),
    );

    fireEvent.click(screen.getByRole("button", { name: "Load briefing.mp3" }));
    expect(
      await screen.findByLabelText("Audio preview for briefing.mp3"),
    ).toHaveAttribute("src", "blob:preview");
  });
});

function artifact(filename: string, mimeType: string): AgentArtifactItem {
  return {
    actions: [],
    artifactId: filename,
    filename,
    grants: [{
      actions: ["artifact.download"],
      grantId: "grant-download",
      representationIds: [],
    }],
    id: filename,
    kind: "artifact",
    manifest: null,
    mimeType,
    representations: [],
    revision: 1n,
    role: "preview",
    schemaVersion: "1",
    selectedRepresentationId: null,
    status: "completed",
  };
}

function runtime(content: string, mimeType: string): AgentSessionRuntime {
  return runtimeBlob(new Blob([content], { type: mimeType }));
}

function runtimeBlob(blob: Blob): AgentSessionRuntime {
  return {
    close: vi.fn(),
    getSnapshot: vi.fn(),
    interrupt: vi.fn(),
    loadArtifact: vi.fn(async () => blob),
    loadOlder: vi.fn(),
    open: vi.fn(),
    resolvePermission: vi.fn(),
    sendMessage: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    updateConfiguration: vi.fn(),
  };
}

function renderArtifact(
  item: AgentArtifactItem,
  agentRuntime: AgentSessionRuntime,
) {
  render(
    <ActivityTimeline
      items={[item]}
      runtime={agentRuntime}
      sessionId="session-1"
    />,
  );
}
