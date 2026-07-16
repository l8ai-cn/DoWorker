import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { InteractionModeSelector } from "./interaction-mode-selector";

describe("InteractionModeSelector", () => {
  it("only enables modes declared by the selected Worker", () => {
    const onChange = vi.fn();
    render(<InteractionModeSelector mode="pty" supportedModes={["pty"]} onChange={onChange} />);

    const acpButton = screen.getByRole("button", { name: /可视化对话/ }) as HTMLButtonElement;
    expect(acpButton.disabled).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: /命令行/ }));
    expect(onChange).toHaveBeenCalledWith("pty");
  });
});
