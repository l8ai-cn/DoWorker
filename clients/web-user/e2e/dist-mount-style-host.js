import { mountEmbeddedAgentWorkspace } from "../dist-embed/agent-cloud-embed.js";

const root = document.getElementById("agent-workspace");
if (!root) throw new Error("agent workspace root is missing");

mountEmbeddedAgentWorkspace(root, {
  access: {
    baseUrl: window.location.origin,
    getAccessToken: () => "invalid-qa-token",
    orgSlug: "dev-org",
    sessionId: "qa-session",
  },
  locale: "zh-CN",
});

window.setTimeout(() => {
  const probe = document.createElement("div");
  probe.className = "h-8 bg-card";
  probe.dataset.qaStyleProbe = "true";
  root.append(probe);
}, 500);
