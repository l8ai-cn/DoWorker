import { HashRouter } from "react-router-dom";
import { AgentCloudApp, type AgentCloudAppProps } from "./embed";

export function AgentCloudStandaloneApp(props: AgentCloudAppProps = {}) {
  return (
    <HashRouter>
      <AgentCloudApp {...props} />
    </HashRouter>
  );
}
