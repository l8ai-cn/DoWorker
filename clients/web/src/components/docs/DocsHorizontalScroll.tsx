import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface DocsHorizontalScrollProps {
  children: ReactNode;
  className?: string;
  hintPosition?: "below" | "top";
}

export function DocsHorizontalScroll({
  children,
  className,
  hintPosition = "below",
}: DocsHorizontalScrollProps) {
  return (
    <div className={cn("relative", className)}>
      <div className="overflow-x-auto pb-2 [scrollbar-width:thin]">
        {children}
      </div>
      <span
        aria-hidden
        className={cn(
          "pointer-events-none absolute flex h-6 w-6 items-center justify-center rounded-full bg-background/90 text-sm font-semibold text-muted-foreground ring-1 ring-border/60 shadow-sm sm:hidden",
          hintPosition === "top" ? "right-2 top-2" : "-bottom-7 right-2"
        )}
      >
        ↔
      </span>
    </div>
  );
}
