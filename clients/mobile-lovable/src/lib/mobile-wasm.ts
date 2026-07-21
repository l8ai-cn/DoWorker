import initWasm, {
  WasmApiClient,
  type WasmAcpSessionManager,
  type WasmPodService,
} from "agent-cloud-wasm";
import { getMobileAuthManager, mobileAuthBaseUrl } from "./mobile-auth-manager";

let apiClientPromise: Promise<WasmApiClient> | undefined;
let podServicePromise: Promise<WasmPodService> | undefined;
let acpManagerPromise: Promise<WasmAcpSessionManager> | undefined;

export function getMobileApiClient(): Promise<WasmApiClient> {
  if (!apiClientPromise) {
    apiClientPromise = (async () => {
      await initWasm();
      return new WasmApiClient(mobileAuthBaseUrl(), await getMobileAuthManager());
    })();
  }
  return apiClientPromise;
}

export function getMobilePodService(): Promise<WasmPodService> {
  if (!podServicePromise) {
    podServicePromise = getMobileApiClient().then((client) => client.create_pod_service());
  }
  return podServicePromise;
}

export function getMobileAcpManager(): Promise<WasmAcpSessionManager> {
  if (!acpManagerPromise) {
    acpManagerPromise = getMobileApiClient().then((client) => client.get_acp_manager());
  }
  return acpManagerPromise;
}
