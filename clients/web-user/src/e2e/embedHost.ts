import {
  EMBED_OPEN_MESSAGE,
  isEmbedReadyMessage,
} from "@/embed-session/embedParentHandshake";

const params = new URLSearchParams(window.location.hash.slice(1));
const context = requiredParam(params, "embed_context");
const redemptionProof = requiredParam(params, "redemption_proof");
const frame = document.getElementById("agent-session");

if (!(frame instanceof HTMLIFrameElement)) {
  throw new Error("agent session iframe is missing");
}
const frameWindow = frame.contentWindow;
if (!frameWindow) throw new Error("agent session iframe window is missing");

applyFrameSize(frame, params);
window.history.replaceState(null, "", window.location.pathname);
window.addEventListener("message", (event) => {
  if (
    event.source !== frameWindow ||
    event.origin !== window.location.origin ||
    !isEmbedReadyMessage(event.data)
  ) {
    return;
  }
  frameWindow.postMessage(
    {
      type: EMBED_OPEN_MESSAGE,
      version: 1,
      redemptionProof,
    },
    window.location.origin,
  );
});
frame.src = `/iframe.html?embed_context=${encodeURIComponent(context)}`;

function requiredParam(params: URLSearchParams, key: string): string {
  const value = params.get(key);
  if (!value) throw new Error(`${key} is required`);
  return value;
}

function applyFrameSize(frame: HTMLIFrameElement, params: URLSearchParams): void {
  const width = Number(params.get("width"));
  const height = Number(params.get("height"));
  if (width > 0) frame.style.width = `${width}px`;
  if (height > 0) frame.style.height = `${height}px`;
}
