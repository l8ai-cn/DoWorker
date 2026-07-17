import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { DoWorkerStandaloneApp } from "./standalone";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Do Worker root element is missing");
}

createRoot(rootElement).render(
  <StrictMode>
    <DoWorkerStandaloneApp />
  </StrictMode>,
);
