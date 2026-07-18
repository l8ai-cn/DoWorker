import "@testing-library/jest-dom/vitest";

import { render, screen, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import { AgentWorkspaceLocaleProvider } from "./AgentWorkspaceLocaleContext";
import { ArtifactPreview } from "./ArtifactPreview";

const pdfViewerModuleLoaded = vi.hoisted(() => vi.fn());

vi.mock("./viewers/pdf/ArtifactPdfPreview", () => {
  pdfViewerModuleLoaded();
  return {
    ArtifactPdfPreview: ({ filename }: { filename: string }) => (
      <div aria-label={`Loaded PDF preview for ${filename}`} />
    ),
  };
});

it("loads the PDF viewer module only after a PDF preview opens", async () => {
  const view = renderPreview("image");

  expect(screen.getByRole("img", { name: "preview.png" })).toBeVisible();
  expect(pdfViewerModuleLoaded).not.toHaveBeenCalled();

  view.rerender(preview("pdf"));

  expect(screen.getByRole("status")).toHaveTextContent("Loading preview.pdf");
  await waitFor(() => expect(pdfViewerModuleLoaded).toHaveBeenCalledTimes(1));
  expect(
    await screen.findByLabelText("Loaded PDF preview for preview.pdf"),
  ).toBeVisible();
});

function renderPreview(kind: "image" | "pdf") {
  return render(preview(kind));
}

function preview(kind: "image" | "pdf") {
  return (
    <AgentWorkspaceLocaleProvider locale="en-US">
      <ArtifactPreview
        filename={`preview.${kind === "pdf" ? "pdf" : "png"}`}
        kind={kind}
        onLoad={vi.fn()}
        onRetry={vi.fn()}
        state={{
          status: "ready",
          url: "blob:preview",
          mimeType: kind === "pdf" ? "application/pdf" : "image/png",
          text: null,
          textTruncated: false,
        }}
      />
    </AgentWorkspaceLocaleProvider>
  );
}
