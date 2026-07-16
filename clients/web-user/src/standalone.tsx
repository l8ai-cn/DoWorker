import { HashRouter } from "react-router-dom";
import { DoWorkerApp, type DoWorkerAppProps } from "./embed";

export function DoWorkerStandaloneApp(props: DoWorkerAppProps = {}) {
  return (
    <HashRouter>
      <DoWorkerApp {...props} />
    </HashRouter>
  );
}
