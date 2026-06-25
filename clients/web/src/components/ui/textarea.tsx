"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: string;
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, ...props }, ref) => {
    return (
      <div className="w-full">
        <textarea
          className={cn(
            "flex min-h-[60px] w-full rounded-md bg-surface-raised px-3 py-2 text-sm shadow-xs ring-1 ring-border/35 motion-interactive placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35 focus-visible:shadow-[0_0_0_3px_color-mix(in_srgb,var(--ring)_12%,transparent)] disabled:cursor-not-allowed disabled:opacity-50",
            error && "ring-destructive/50 focus-visible:ring-destructive/35 focus-visible:shadow-[0_0_0_3px_color-mix(in_srgb,var(--destructive)_12%,transparent)]",
            className
          )}
          ref={ref}
          {...props}
        />
        {error && (
          <p className="mt-1 text-xs text-destructive">{error}</p>
        )}
      </div>
    );
  }
);

Textarea.displayName = "Textarea";

export { Textarea };
