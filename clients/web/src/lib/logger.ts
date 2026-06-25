type Level = "trace" | "debug" | "info" | "warn" | "error";

// Single fan-out point for renderer-side log emission. Routes (in priority
// order):
//   1. Web console logger
//   2. Native console — on Web the wasm-side tracing subscriber renders
//      Rust events to console too, so the destination is the same.
// Why not also push to the wasm subscriber from here: doing so requires
// importing `wasm-core` (which depends on the `agentsmesh-wasm` package).
// Keeping logger console-only avoids pulling wasm into static routes.
function emit(level: Level, target: string, msg: string): void {
  const formatted = `[${target}] ${msg}`;
  switch (level) {
    case "error":
      console.error(formatted);
      break;
    case "warn":
      console.warn(formatted);
      break;
    case "info":
      console.info(formatted);
      break;
    default:
      console.debug(formatted);
  }
}

export const logger = {
  trace: (target: string, msg: string) => emit("trace", target, msg),
  debug: (target: string, msg: string) => emit("debug", target, msg),
  info: (target: string, msg: string) => emit("info", target, msg),
  warn: (target: string, msg: string) => emit("warn", target, msg),
  error: (target: string, msg: string) => emit("error", target, msg),
};
