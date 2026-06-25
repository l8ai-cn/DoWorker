"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, error, ...props }, ref) => {
    return (
      <div className="w-full">
        <input
          type={type}
          className={cn(
            "flex h-9 w-full rounded-md bg-surface-raised px-3 py-1 text-sm shadow-xs ring-1 ring-border/35 motion-interactive file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35 focus-visible:shadow-[0_0_0_3px_color-mix(in_srgb,var(--ring)_12%,transparent)] disabled:cursor-not-allowed disabled:opacity-50",
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

Input.displayName = "Input";

export { Input };
