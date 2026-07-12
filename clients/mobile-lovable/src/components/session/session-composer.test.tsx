import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { SessionComposer } from "./session-composer";
import { SessionActionProvider } from "@/lib/session-action-context";

describe("SessionComposer", () => {
  it("sends an attachment without requiring text", async () => {
    const onSend = vi.fn().mockResolvedValue(undefined);
    const { container } = render(
      <SessionActionProvider value={{ onSend }}>
        <SessionComposer />
      </SessionActionProvider>,
    );
    const file = new File(["image"], "design.png", { type: "image/png" });
    const input = container.querySelector('input[type="file"]');

    fireEvent.change(input as HTMLInputElement, { target: { files: [file] } });
    fireEvent.click(screen.getByLabelText("发送"));

    await waitFor(() => expect(onSend).toHaveBeenCalledWith("", [file]));
  });
});
