import "@testing-library/jest-dom/vitest";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { AgentWorkspaceLocaleProvider } from "../../AgentWorkspaceLocaleContext";
import { LazyArtifactPdfPreview } from "./LazyArtifactPdfPreview";

vi.mock("./ArtifactPdfPreview", () => {
  throw new Error("PDF viewer chunk unavailable");
});

it("shows the PDF error state when the viewer module cannot load", async () => {
  const consoleError = vi
    .spyOn(console, "error")
    .mockImplementation(() => undefined);

  render(
    <AgentWorkspaceLocaleProvider locale="en-US">
      <LazyArtifactPdfPreview filename="report.pdf" src="blob:report" />
    </AgentWorkspaceLocaleProvider>,
  );

  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Artifact loading failed. Try again.",
  );
  expect(consoleError).toHaveBeenCalledWith(
    "PDF preview failed",
    expect.any(Error),
  );
  consoleError.mockRestore();
});
