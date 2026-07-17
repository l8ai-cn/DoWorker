import { useLayoutEffect, useRef, useState } from "react";

export type WorkbenchContainerMode = "narrow" | "medium" | "wide";

export function workbenchContainerMode(width: number): WorkbenchContainerMode {
  if (width >= 960) return "wide";
  if (width >= 640) return "medium";
  return "narrow";
}

export function useWorkbenchContainerMode() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [mode, setMode] = useState<WorkbenchContainerMode>("narrow");

  useLayoutEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    const measure = () => {
      setMode(workbenchContainerMode(container.getBoundingClientRect().width));
    };
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(container);
    return () => observer.disconnect();
  }, []);

  return { containerRef, mode };
}
