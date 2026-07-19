import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/dialog";

describe("Dialog primitives", () => {
  it("opens content without Radix ref warnings", async () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    render(
      <Dialog>
        <DialogTrigger asChild>
          <Button type="button">Open fork dialog</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogTitle>Fork session</DialogTitle>
          <p>Configure the fork.</p>
        </DialogContent>
      </Dialog>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open fork dialog" }));

    expect(await screen.findByText("Configure the fork.")).toBeInTheDocument();
    expect(
      consoleError.mock.calls.some(([message]) =>
        String(message).includes("Function components cannot be given refs"),
      ),
    ).toBe(false);
    consoleError.mockRestore();
  });
});
