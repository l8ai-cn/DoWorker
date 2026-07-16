import { createRoot } from "react-dom/client";
import { DoWorkerStandaloneApp } from "./standalone";
import type { DoWorkerAppProps } from "./embed";

export interface MountedDoWorkerApp {
  unmount(): void;
}

export function mountDoWorkerApp(
  element: Element,
  props: DoWorkerAppProps = {},
): MountedDoWorkerApp {
  const root = createRoot(element);
  root.render(<DoWorkerStandaloneApp {...props} />);

  return { unmount: () => root.unmount() };
}
