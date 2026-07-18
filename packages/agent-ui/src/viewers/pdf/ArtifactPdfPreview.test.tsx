import "@testing-library/jest-dom/vitest";

import { render, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AgentWorkspaceLocaleProvider } from "../../AgentWorkspaceLocaleContext";
import { ArtifactPdfPreview } from "./ArtifactPdfPreview";

const getDocument = vi.fn();

vi.mock("pdfjs-dist", () => ({
  GlobalWorkerOptions: { workerSrc: "" },
  getDocument,
}));

describe("ArtifactPdfPreview lifecycle", () => {
  it("destroys a pending loading task when the preview closes", async () => {
    const destroy = vi.fn(async () => undefined);
    getDocument.mockReturnValue({
      destroy,
      promise: new Promise(() => undefined),
    });
    const view = render(
      <AgentWorkspaceLocaleProvider locale="en-US">
        <ArtifactPdfPreview filename="report.pdf" src="blob:report" />
      </AgentWorkspaceLocaleProvider>,
    );

    await waitFor(() => expect(getDocument).toHaveBeenCalled());
    view.unmount();

    await waitFor(() => expect(destroy).toHaveBeenCalledTimes(1));
  });
});
