import { pathToFileURL } from "node:url";
import {
  ACP_COMMAND,
  ACP_EVENT,
  ACP_SNAPSHOT,
  binaryFrame,
  CONTROL,
  TERMINAL_INPUT,
  TERMINAL_OUTPUT,
  TERMINAL_SNAPSHOT,
  browserRelayUrl,
  jsonFrame,
  messageBytes,
  parseJson,
  resizeFrame,
} from "./mobile-relay-frame-codec.mjs";
const RENEW_AHEAD_MS = 10_000;
const MIN_RENEW_DELAY_MS = 5_000;

export async function runMobileRelayDataPlaneSmoke({
  marker,
  mode,
  relayUrl,
  token,
  timeoutMs = 90_000,
  WebSocketImpl = globalThis.WebSocket,
  debug = null,
  schedule = setTimeout,
  cancelSchedule = clearTimeout,
}) {
  if (mode !== "acp" && mode !== "pty") throw new Error("mode must be acp or pty");
  if (!relayUrl || !token || !WebSocketImpl) throw new Error("relay connection is unavailable");
  if (!marker) throw new Error(`${mode.toUpperCase()} smoke requires a response marker`);

  return new Promise((resolve, reject) => {
    const log = typeof debug === "function" ? debug : () => {};
    let completed = false;
    let dataReady = false;
    let leaseRequested = false;
    let interactionSent = false;
    let renewalTimer = null;
    let assistantText = "";
    const socket = new WebSocketImpl(browserRelayUrl(relayUrl, token));
    socket.binaryType = "arraybuffer";
    const timeout = setTimeout(() => finish(new Error(`${mode} Relay smoke timed out`)), timeoutMs);

    socket.onerror = () => finish(new Error(`${mode} Relay WebSocket failed`));
    socket.onclose = () => {
      if (!completed) finish(new Error(`${mode} Relay WebSocket closed before verification`));
    };
    socket.onmessage = (event) => {
      void handleMessage(event.data).catch((error) => finish(error));
    };

    async function handleMessage(data) {
      const bytes = await messageBytes(data);
      if (bytes.length < 1) throw new Error("Relay returned an empty frame");
      const type = bytes[0];
      const payloadBytes = bytes.subarray(1);
      const payload = parseJson(payloadBytes);

      if (!dataReady && type === (mode === "acp" ? ACP_SNAPSHOT : TERMINAL_SNAPSHOT)) {
        dataReady = true;
        log(`received ${mode.toUpperCase()} snapshot`);
        leaseRequested = true;
        log(`requested ${mode.toUpperCase()} control lease`);
        socket.send(jsonFrame(CONTROL, {
          action: "acquire",
          client_label: "mobile-release-smoke",
          type: "control_lease",
        }));
        return;
      }
      if (type === CONTROL && payload?.type === "control_lease") {
        log(`${mode.toUpperCase()} control lease status=${payload.status ?? "unknown"}`);
        if (payload.status === "granted" && leaseRequested) {
          log(`${mode.toUpperCase()} control lease granted`);
          scheduleRenewal(payload);
          if (!interactionSent) {
            interactionSent = true;
            if (mode === "pty") {
              socket.send(resizeFrame(120, 40));
              socket.send(binaryFrame(TERMINAL_INPUT, new TextEncoder().encode(`${marker}\n`)));
              log("sent PTY resize and input");
              return;
            }
            socket.send(jsonFrame(ACP_COMMAND, {
              prompt: `Reply with exactly ${marker}.`,
              type: "prompt",
            }));
            log("sent ACP prompt");
          }
          return;
        }
        if (
          payload.status === "busy" ||
          payload.status === "control_required" ||
          payload.status === "expired" ||
          payload.status === "released"
        ) {
          throw new Error(`${mode} Relay control lease was not granted`);
        }
        return;
      }
      if (mode === "pty" && type === TERMINAL_OUTPUT && text(payloadBytes).includes(marker)) {
        log("received matching PTY output");
        finish();
      }
      if (
        mode === "acp" &&
        type === ACP_EVENT &&
        payload?.type === "contentChunk" &&
        payload.role === "assistant" &&
        typeof payload.text === "string"
      ) {
        assistantText += payload.text;
        if (assistantText.includes(marker)) {
          log("received matching ACP assistant chunk");
          finish();
        }
      }
    }

    function finish(error) {
      if (completed) return;
      completed = true;
      clearTimeout(timeout);
      if (renewalTimer !== null) cancelSchedule(renewalTimer);
      socket.close();
      if (error) reject(error);
      else resolve();
    }

    function scheduleRenewal(payload) {
      const leaseId = payload.lease_id;
      const expiresAt = Number(payload.expires_at);
      if (!leaseId || !Number.isFinite(expiresAt)) {
        throw new Error(`${mode.toUpperCase()} Relay granted an invalid control lease`);
      }
      if (renewalTimer !== null) cancelSchedule(renewalTimer);
      const delay = Math.max(MIN_RENEW_DELAY_MS, expiresAt - Date.now() - RENEW_AHEAD_MS);
      renewalTimer = schedule(() => {
        renewalTimer = null;
        log(`renewed ${mode.toUpperCase()} control lease`);
        socket.send(jsonFrame(CONTROL, {
          action: "renew",
          lease_id: leaseId,
          type: "control_lease",
        }));
      }, delay);
    }
  });
}

function text(bytes) {
  return new TextDecoder().decode(bytes);
}

async function main() {
  const input = await readInput();
  const debug = process.env.MOBILE_RELAY_SMOKE_DEBUG === "1"
    ? (message) => process.stderr.write(`[mobile-relay-smoke] ${message}\n`)
    : undefined;
  await runMobileRelayDataPlaneSmoke({ ...input, debug });
  process.stdout.write(`${input.mode} Relay data-plane smoke passed\n`);
}

async function readInput() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch((error) => {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 1;
  });
}
