import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";

import { GenericArtifactCard } from "./GenericArtifactCard";
import type { AgentArtifactItem, AgentSessionRuntime } from "./contracts";

it("downloads the original representation separately from its preview", async () => {
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, "click")
    .mockImplementation(() => undefined);
  const loadArtifact = vi.fn(async (
    _sessionId: string,
    _artifactId: string,
    representationId?: string,
  ) => new Blob([representationId ?? "preview"], { type: "text/plain" }));
  Object.defineProperty(URL, "createObjectURL", {
    configurable: true,
    value: vi.fn(() => "blob:artifact"),
  });
  Object.defineProperty(URL, "revokeObjectURL", {
    configurable: true,
    value: vi.fn(),
  });

  render(
    <GenericArtifactCard
      filename="report.docx"
      item={artifact()}
      runtime={{ loadArtifact } as unknown as AgentSessionRuntime}
      sessionId="session-1"
    />,
  );

  fireEvent.click(
    await screen.findByRole("button", { name: "Download report.docx" }),
  );

  await waitFor(() =>
    expect(loadArtifact).toHaveBeenCalledWith(
      "session-1",
      "artifact-1",
      "original",
    ),
  );
  expect(click).toHaveBeenCalledOnce();
  click.mockRestore();
});

function artifact(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "artifact-1",
    filename: "report.docx",
    grants: [downloadGrant()],
    id: "artifact-item-1",
    kind: "artifact",
    manifest: null,
    mimeType: "text/plain",
    representations: [
      {
        filename: "report.docx",
        mediaType:
          "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        representationId: "original",
        revision: 1n,
        status: "ready",
      },
      {
        filename: "report-preview.txt",
        mediaType: "text/plain",
        representationId: "preview-text",
        revision: 1n,
        status: "ready",
      },
    ],
    revision: 1n,
    role: "preview",
    schemaVersion: "1",
    selectedRepresentationId: "preview-text",
    status: "completed",
  };
}

function downloadGrant() {
  return {
    actions: ["artifact.download"],
    grantId: "grant-download",
    representationIds: [],
  };
}
