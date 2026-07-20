import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

const ROOT = process.cwd();
const RUNTIME_BACKEND_ENV = join(ROOT, "deploy", "dev", "runtime", "backend", "pattern-locked-backend.env");
const DEV_ENV = join(ROOT, "deploy", "dev", ".env");

function readEnvFile(path) {
  if (!existsSync(path)) return {};
  return Object.fromEntries(
    readFileSync(path, "utf8")
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter((line) => line && !line.startsWith("#") && line.includes("="))
      .map((line) => {
        const idx = line.indexOf("=");
        return [line.slice(0, idx), line.slice(idx + 1)];
      }),
  );
}

const runtimeEnv = { ...readEnvFile(DEV_ENV), ...readEnvFile(RUNTIME_BACKEND_ENV), ...process.env };

function urlFromHostPort(host, port, fallback) {
  if (!port) return fallback;
  return `http://${host}:${port}`;
}

function portFromAddress(value) {
  const match = String(value ?? "").match(/:(\d+)$/);
  return match?.[1] ?? "";
}

function primaryPort() {
  if (!runtimeEnv.PRIMARY_DOMAIN) return "";
  return portFromAddress(runtimeEnv.PRIMARY_DOMAIN);
}

function slotPort(slot) {
  const port = Number(primaryPort());
  return Number.isFinite(port) && port > 0 ? String(port + slot) : "";
}

export function devUrl(name, fallback) {
  if (runtimeEnv[name]) return runtimeEnv[name];
  switch (name) {
    case "WEB_URL":
      if (runtimeEnv.PUBLIC_WEB_URL) return runtimeEnv.PUBLIC_WEB_URL;
      return urlFromHostPort("127.0.0.1", slotPort(7), urlFromHostPort("127.0.0.1", runtimeEnv.WEB_PORT, fallback));
    case "WEB_USER_URL":
      return urlFromHostPort("127.0.0.1", slotPort(20), urlFromHostPort("127.0.0.1", runtimeEnv.WEB_USER_PORT, fallback));
    case "TRAEFIK_API_URL":
      if (runtimeEnv.PRIMARY_DOMAIN) return `http://${runtimeEnv.PRIMARY_DOMAIN}`;
      return urlFromHostPort("127.0.0.1", runtimeEnv.HTTP_PORT, fallback);
    case "WEB_USER_AUTH_URL":
      return runtimeEnv.VITE_DO_WORKER_API_URL ?? runtimeEnv.VITE_AGENTSMESH_API_URL ?? fallback;
    case "SESSION_COMPAT_API_URL":
      return urlFromHostPort(
        "localhost",
        portFromAddress(runtimeEnv.SERVER_ADDRESS) || slotPort(15),
        urlFromHostPort("localhost", runtimeEnv.BACKEND_HTTP_PORT, fallback),
      );
    default:
      return fallback;
  }
}
