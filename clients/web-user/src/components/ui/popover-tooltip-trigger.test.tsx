import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

describe("PopoverTrigger with a tooltip child", () => {
  it("opens the popover without giving the tooltip root a ref", async () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    render(
      <TooltipProvider>
        <Popover>
          <PopoverTrigger asChild>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button type="button">Agent tools and policies</Button>
              </TooltipTrigger>
              <TooltipContent>Tools tooltip</TooltipContent>
            </Tooltip>
          </PopoverTrigger>
          <PopoverContent>Tools popover</PopoverContent>
        </Popover>
      </TooltipProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Agent tools and policies" }));

    expect(await screen.findByText("Tools popover")).toBeInTheDocument();
    expect(
      consoleError.mock.calls.some(([message]) =>
        String(message).includes("Function components cannot be given refs"),
      ),
    ).toBe(false);
    consoleError.mockRestore();
  });
});
