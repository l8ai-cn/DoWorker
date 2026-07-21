import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AgentCloudStandaloneApp } from "./standalone";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Agent Cloud root element is missing");
}

createRoot(rootElement).render(
  <StrictMode>
    <AgentCloudStandaloneApp />
  </StrictMode>,
);
