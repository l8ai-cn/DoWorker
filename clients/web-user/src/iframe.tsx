import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { EmbeddedSessionIframe } from "./embed-session/EmbeddedSessionIframe";
import "./index.css";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Agent Cloud iframe root element is missing");
}

createRoot(rootElement).render(
  <StrictMode>
    <EmbeddedSessionIframe />
  </StrictMode>,
);
