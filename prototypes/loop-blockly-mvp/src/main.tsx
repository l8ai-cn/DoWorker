import { createRoot } from "react-dom/client";

import { App } from "./app";
import "./styles/base.css";
import "./styles/workbench.css";
import "./styles/workbench-output.css";
import "./styles/workbench-dialog.css";
import "./styles/responsive.css";

const root = document.getElementById("root");
if (!root) throw new Error("Missing #root element.");
createRoot(root).render(<App />);
