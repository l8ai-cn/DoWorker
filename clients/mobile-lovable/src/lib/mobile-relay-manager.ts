import initWasm, { WasmRelayManager } from "do-worker-wasm";

let managerPromise: Promise<WasmRelayManager> | undefined;

export function getMobileRelayManager(): Promise<WasmRelayManager> {
  if (!managerPromise) {
    managerPromise = initWasm().then(() => new WasmRelayManager());
  }
  return managerPromise;
}
