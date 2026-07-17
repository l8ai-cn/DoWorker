import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { BlockProgrammingWorkbench } from "../BlockProgrammingWorkbench";

describe("BlockProgrammingWorkbench", () => {
  it("renders host-provided regions and labels", () => {
    render(
      <BlockProgrammingWorkbench
        canvas={<div>Canvas region</div>}
        editor={<div>Editor region</div>}
        messages={{
          canvasHint: "Double-click empty space to insert",
          canvasTitle: "Block canvas",
          editorMetadata: "Draft 3 · Semantic 2",
          editorTitle: "Program source",
        }}
        status={<div>Status region</div>}
        toolbar={<div>Toolbar region</div>}
      />,
    );

    expect(screen.getByText("Toolbar region")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Block canvas" })).toBeInTheDocument();
    expect(screen.getByText("Double-click empty space to insert")).toBeInTheDocument();
    expect(screen.getByText("Canvas region")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Program source" })).toBeInTheDocument();
    expect(screen.getByText("Draft 3 · Semantic 2")).toBeInTheDocument();
    expect(screen.getByText("Editor region")).toBeInTheDocument();
    expect(screen.getByText("Status region")).toBeInTheDocument();
  });
});
