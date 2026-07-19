"use client";

import * as React from "react";
import { Popover as PopoverPrimitive } from "radix-ui";

import { Tooltip, TooltipTrigger } from "@/components/ui/tooltip";
import { getEmbedRoot } from "@/lib/host";
import { cn } from "@/lib/utils";

function Popover({ ...props }: React.ComponentProps<typeof PopoverPrimitive.Root>) {
  return <PopoverPrimitive.Root data-slot="popover" {...props} />;
}

function isTooltipRoot(
  node: React.ReactNode,
): node is React.ReactElement<React.ComponentPropsWithoutRef<typeof Tooltip>> {
  return React.isValidElement(node) && node.type === Tooltip;
}

function isTooltipTrigger(
  node: React.ReactNode,
): node is React.ReactElement<React.ComponentPropsWithoutRef<typeof TooltipTrigger>> {
  return React.isValidElement(node) && node.type === TooltipTrigger;
}

const PopoverTrigger = React.forwardRef<
  React.ElementRef<typeof PopoverPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Trigger>
>(({ asChild, children, ...props }, ref) => {
  if (asChild && isTooltipRoot(children)) {
    const tooltipChildren = React.Children.map(children.props.children, (node) => {
      if (!isTooltipTrigger(node)) return node;
      return (
        <TooltipTrigger {...node.props}>
          <PopoverPrimitive.Trigger ref={ref} asChild data-slot="popover-trigger" {...props}>
            {node.props.children}
          </PopoverPrimitive.Trigger>
        </TooltipTrigger>
      );
    });
    return React.cloneElement(children, undefined, tooltipChildren);
  }
  return (
    <PopoverPrimitive.Trigger ref={ref} asChild={asChild} data-slot="popover-trigger" {...props}>
      {children}
    </PopoverPrimitive.Trigger>
  );
});
PopoverTrigger.displayName = "PopoverTrigger";

function PopoverContent({
  className,
  align = "center",
  sideOffset = 4,
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Content>) {
  return (
    <PopoverPrimitive.Portal container={getEmbedRoot() ?? undefined}>
      <PopoverPrimitive.Content
        data-slot="popover-content"
        align={align}
        sideOffset={sideOffset}
        className={cn(
          "z-50 flex w-72 origin-(--radix-popover-content-transform-origin) flex-col gap-2.5 rounded-lg bg-popover p-2.5 text-sm text-popover-foreground shadow-md ring-1 ring-foreground/10 outline-hidden duration-150 ease-[cubic-bezier(0.16,1,0.3,1)] data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95",
          className,
        )}
        {...props}
      />
    </PopoverPrimitive.Portal>
  );
}

function PopoverAnchor({ ...props }: React.ComponentProps<typeof PopoverPrimitive.Anchor>) {
  return <PopoverPrimitive.Anchor data-slot="popover-anchor" {...props} />;
}

function PopoverHeader({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="popover-header"
      className={cn("flex flex-col gap-0.5 text-sm", className)}
      {...props}
    />
  );
}

function PopoverTitle({ className, ...props }: React.ComponentProps<"h2">) {
  return <div data-slot="popover-title" className={cn("font-medium", className)} {...props} />;
}

function PopoverDescription({ className, ...props }: React.ComponentProps<"p">) {
  return (
    <p
      data-slot="popover-description"
      className={cn("text-muted-foreground", className)}
      {...props}
    />
  );
}

export {
  Popover,
  PopoverAnchor,
  PopoverContent,
  PopoverDescription,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
};
