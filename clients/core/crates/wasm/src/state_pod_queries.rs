use agentsmesh_types::proto_pod_state_v1::ReplaceCachedPodsRequest;
use agentsmesh_types::proto_pod_v1::ListPodsResponse;
use prost::Message;
use wasm_bindgen::prelude::*;

use crate::state_pod::WasmPodState;

fn decode_err<E: std::fmt::Display>(error: E) -> JsValue {
    JsValue::from_str(&format!("decode: {error}"))
}

#[wasm_bindgen]
impl WasmPodState {
    pub fn query_pods_bytes(&self, query_key: &str) -> Vec<u8> {
        let pods = self
            .state
            .read()
            .pod_query_snapshots
            .get(query_key)
            .to_vec();
        ReplaceCachedPodsRequest { pods }.encode_to_vec()
    }

    pub fn apply_fetched_pod_query(
        &self,
        query_key: &str,
        response_bytes: &[u8],
    ) -> Result<(), JsValue> {
        let response = ListPodsResponse::decode(response_bytes).map_err(decode_err)?;
        self.state
            .write()
            .pod_query_snapshots
            .replace(query_key, response.items);
        Ok(())
    }
}
