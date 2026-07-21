import initWasm, { WasmRelayManager } from "agent-cloud-wasm";

let managerPromise: Promise<WasmRelayManager> | undefined;

export function getMobileRelayManager(): Promise<WasmRelayManager> {
  if (!managerPromise) {
    managerPromise = initWasm().then(() => new WasmRelayManager());
  }
  return managerPromise;
}
