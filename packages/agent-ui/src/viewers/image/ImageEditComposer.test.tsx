import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { ImageEditComposer } from "./ImageEditComposer";

describe("ImageEditComposer", () => {
  it("disables submission for an empty or whitespace-only instruction", () => {
    render(
      <ImageEditComposer
        actionSchemaVersion="1"
        artifactId="image-1"
        baseRevision={7n}
        onSubmit={vi.fn()}
        sourceDimensions={{ height: 900, width: 1600 }}
      />,
    );

    const submit = screen.getByRole("button", { name: "提交编辑" });
    expect(submit).toBeDisabled();

    fireEvent.change(screen.getByRole("textbox", { name: "编辑说明" }), {
      target: { value: "   " },
    });
    expect(submit).toBeDisabled();
  });

  it("emits an image.edit action with normalized geometry", () => {
    const onSubmit = vi.fn();
    const randomUUID = vi
      .spyOn(globalThis.crypto, "randomUUID")
      .mockReturnValue("00000000-0000-4000-8000-000000000001");
    const alert = vi.spyOn(window, "alert");
    const prompt = vi.spyOn(window, "prompt");

    render(
      <ImageEditComposer
        actionSchemaVersion="1"
        artifactId="image-1"
        baseRevision={12n}
        normalizedRegion={{ height: 0.4, width: 0.3, x: 0.1, y: 0.2 }}
        onSubmit={onSubmit}
        representationId="source"
        sourceDimensions={{ height: 1080, width: 1920 }}
      />,
    );

    fireEvent.change(screen.getByRole("textbox", { name: "编辑说明" }), {
      target: { value: "  删除选区中的文字  " },
    });
    fireEvent.click(screen.getByRole("button", { name: "提交编辑" }));

    expect(onSubmit).toHaveBeenCalledWith({
      actionSchemaVersion: "1",
      actionType: "image.edit",
      artifactId: "image-1",
      baseRevision: 12n,
      commandId: "00000000-0000-4000-8000-000000000001",
      payload: {
        instruction: "删除选区中的文字",
        normalizedRegion: {
          height: 0.4,
          width: 0.3,
          x: 0.1,
          y: 0.2,
        },
        sourceDimensions: { height: 1080, width: 1920 },
      },
      representationId: "source",
    });
    expect(alert).not.toHaveBeenCalled();
    expect(prompt).not.toHaveBeenCalled();

    randomUUID.mockRestore();
    alert.mockRestore();
    prompt.mockRestore();
  });

  it("omits normalized geometry when no region is selected", () => {
    const onSubmit = vi.fn();
    render(
      <ImageEditComposer
        actionSchemaVersion="1"
        artifactId="image-1"
        baseRevision={3n}
        onSubmit={onSubmit}
        sourceDimensions={{ height: 600, width: 800 }}
      />,
    );

    fireEvent.change(screen.getByRole("textbox", { name: "编辑说明" }), {
      target: { value: "增强清晰度" },
    });
    fireEvent.click(screen.getByRole("button", { name: "提交编辑" }));

    expect(onSubmit.mock.calls[0][0].payload).toEqual({
      instruction: "增强清晰度",
      sourceDimensions: { height: 600, width: 800 },
    });
  });
});
