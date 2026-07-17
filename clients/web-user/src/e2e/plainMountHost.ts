import { redeemEmbedContext } from "@/embed-context";
import { mountEmbeddedAgentWorkspace } from "@/mountEmbeddedAgentWorkspace";
import "@/index.css";

const params = new URLSearchParams(window.location.hash.slice(1));
const context = requiredParam(params, "embed_context");
const redemptionProof = requiredParam(params, "redemption_proof");
const root = document.getElementById("agent-workspace");

if (!root) throw new Error("agent workspace root is missing");
window.history.replaceState(null, "", window.location.pathname);

const access = await redeemEmbedContext(context, redemptionProof);
mountEmbeddedAgentWorkspace(root, {
  access: {
    baseUrl: window.location.origin,
    getAccessToken: () => access.accessToken,
    orgSlug: access.orgSlug,
    sessionId: access.sessionId,
  },
  locale: "zh-CN",
});

function requiredParam(params: URLSearchParams, key: string): string {
  const value = params.get(key);
  if (!value) throw new Error(`${key} is required`);
  return value;
}
