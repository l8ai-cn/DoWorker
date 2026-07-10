"use client";

import { useEffect, useRef, type RefObject } from "react";

const focusable = "a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex='-1'])";

export function useDialogFocus(open: boolean, overlay: RefObject<HTMLElement | null>) {
  const priorFocus = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return;
    priorFocus.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const focusables = () => Array.from(overlay.current?.querySelectorAll<HTMLElement>(focusable) ?? []).filter((element) => !element.hidden);
    const frame = requestAnimationFrame(() => focusables()[0]?.focus());
    const trap = (event: KeyboardEvent) => {
      if (event.key !== "Tab") return;
      const elements = focusables();
      if (elements.length === 0) return;
      const first = elements[0];
      const last = elements[elements.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };
    document.addEventListener("keydown", trap);
    return () => {
      cancelAnimationFrame(frame);
      document.removeEventListener("keydown", trap);
      priorFocus.current?.focus();
    };
  }, [open, overlay]);
}
