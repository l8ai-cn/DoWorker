import { createRoot } from "react-dom/client";
import { AgentCloudStandaloneApp } from "./standalone";
import type { AgentCloudAppProps } from "./embed";

export interface MountedAgentCloudApp {
  unmount(): void;
}

export function mountAgentCloudApp(
  element: Element,
  props: AgentCloudAppProps = {},
): MountedAgentCloudApp {
  const root = createRoot(element);
  root.render(<AgentCloudStandaloneApp {...props} />);

  return { unmount: () => root.unmount() };
}
