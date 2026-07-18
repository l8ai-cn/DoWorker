import { describe, expect, it } from "vitest";
import { workspaceFileArtifacts } from "./workspaceFileArtifacts";

describe("workspaceFileArtifacts", () => {
  it("projects deliverables and ignores ordinary source files", () => {
    expect(
      workspaceFileArtifacts("tool-1", [
        { path: "deliverables/preview.png" },
        { path: "deliverables/deck.pptx" },
        { path: "outputs/report.docx" },
        { path: "outputs/data.xlsx" },
        { path: "outputs/results.csv" },
        { path: "outputs/briefing.mp3" },
        { path: "deliverables/generate-assets.mjs" },
        { path: "deliverables/README.md" },
        { path: "src/main.ts" },
      ]),
    ).toEqual([
      expect.objectContaining({
        artifactId: "workspace:deliverables/preview.png",
        filename: "preview.png",
        mimeType: "image/png",
      }),
      expect.objectContaining({
        artifactId: "workspace:deliverables/deck.pptx",
        filename: "deck.pptx",
        mimeType:
          "application/vnd.openxmlformats-officedocument.presentationml.presentation",
      }),
      expect.objectContaining({
        artifactId: "workspace:outputs/report.docx",
        filename: "report.docx",
        mimeType:
          "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      }),
      expect.objectContaining({
        artifactId: "workspace:outputs/data.xlsx",
        mimeType:
          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      }),
      expect.objectContaining({
        artifactId: "workspace:outputs/results.csv",
        mimeType: "text/csv",
      }),
      expect.objectContaining({
        artifactId: "workspace:outputs/briefing.mp3",
        mimeType: "audio/mpeg",
      }),
      expect.objectContaining({
        artifactId: "workspace:deliverables/generate-assets.mjs",
        filename: "generate-assets.mjs",
        mimeType: "text/javascript",
      }),
      expect.objectContaining({
        artifactId: "workspace:deliverables/README.md",
        filename: "README.md",
        mimeType: "text/markdown",
      }),
    ]);
  });

  it("ignores runtime files and deleted deliverables", () => {
    expect(
      workspaceFileArtifacts("tool-1", [
        { path: ".agent/skills/video/assets/hero.png", status: "created" },
        { path: "node_modules/jszip/graph.svg", status: "created" },
        { path: "output/removed.mp4", status: "deleted" },
        { path: "output/final.mp4", status: "created" },
      ]),
    ).toEqual([
      expect.objectContaining({
        artifactId: "workspace:output/final.mp4",
      }),
    ]);
  });
});
