import type { KeyboardEvent } from "react";

export function focusAdjacentTab(event: KeyboardEvent<HTMLElement>): void {
  if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
  const tabs = Array.from(
    event.currentTarget.querySelectorAll<HTMLElement>('[role="tab"]'),
  ).filter((tab) => !tab.hasAttribute("disabled"));
  if (tabs.length === 0) return;
  const current = tabs.indexOf(document.activeElement as HTMLElement);
  const next =
    event.key === "Home"
      ? 0
      : event.key === "End"
        ? tabs.length - 1
        : (Math.max(current, 0) + (event.key === "ArrowRight" ? 1 : -1) + tabs.length) %
          tabs.length;
  event.preventDefault();
  tabs[next]?.focus();
}
