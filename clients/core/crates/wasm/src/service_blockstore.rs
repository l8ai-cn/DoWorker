use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::BlockstoreService;
use agentcloud_state::blockstore_state::BlockstoreState;
use agentcloud_types::proto_blockstore_state_v1::ApplyRemoteOpRequest;
use prost::Message;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmBlockstoreService(pub(crate) BlockstoreService);

fn decode_err<E: std::fmt::Display>(e: E) -> String {
    format!("decode: {e}")
}

#[wasm_bindgen]
impl WasmBlockstoreService {
    pub(crate) fn new(client: Arc<ApiClient>, state: BlockstoreState) -> Self {
        Self(BlockstoreService::new(client, state))
    }

    // -------- Connect-RPC (binary wire) --------
    //
    // TS encodes the request via @bufbuild/protobuf .toBinary(), passes the
    // Uint8Array in, receives a Uint8Array back, decodes via .fromBinary().
    // No JSON intermediate; conventions §2.5 forbids it on the client.
    //
    // js_name is camelCase to match JS conventions; the `_connect` suffix
    // marks the migration lane so the legacy JSON methods can coexist until
    // the full 26-service migration ships.

    #[wasm_bindgen(js_name = applyOpsConnect)]
    pub async fn apply_ops_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.apply_ops_connect(request).await
    }

    #[wasm_bindgen(js_name = listWorkspacesConnect)]
    pub async fn list_workspaces_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_workspaces_connect(request).await
    }

    #[wasm_bindgen(js_name = ensureDefaultWorkspaceConnect)]
    pub async fn ensure_default_workspace_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.ensure_default_workspace_connect(request).await
    }

    #[wasm_bindgen(js_name = createWorkspaceConnect)]
    pub async fn create_workspace_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_workspace_connect(request).await
    }

    #[wasm_bindgen(js_name = deleteWorkspaceConnect)]
    pub async fn delete_workspace_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.delete_workspace_connect(request).await
    }

    #[wasm_bindgen(js_name = getBlockConnect)]
    pub async fn get_block_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_block_connect(request).await
    }

    #[wasm_bindgen(js_name = listChildrenConnect)]
    pub async fn list_children_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_children_connect(request).await
    }

    #[wasm_bindgen(js_name = listBacklinksConnect)]
    pub async fn list_backlinks_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_backlinks_connect(request).await
    }

    #[wasm_bindgen(js_name = getSubtreeConnect)]
    pub async fn get_subtree_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_subtree_connect(request).await
    }

    #[wasm_bindgen(js_name = streamOpsConnect)]
    pub async fn stream_ops_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.stream_ops_connect(request).await
    }

    #[wasm_bindgen(js_name = exportWorkspaceConnect)]
    pub async fn export_workspace_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.export_workspace_connect(request).await
    }

    #[wasm_bindgen(js_name = listTypeDefsConnect)]
    pub async fn list_type_defs_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_type_defs_connect(request).await
    }

    #[wasm_bindgen(js_name = getBlockAtConnect)]
    pub async fn get_block_at_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_block_at_connect(request).await
    }

    #[wasm_bindgen(js_name = semanticSearchConnect)]
    pub async fn semantic_search_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.semantic_search_connect(request).await
    }

    #[wasm_bindgen(js_name = memoryRetrieveConnect)]
    pub async fn memory_retrieve_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.memory_retrieve_connect(request).await
    }

    // -------- Legacy JSON methods (preserved during dual-track) --------
    //
    // All block/ref/workspace data lives in the local Rust state cache. The
    // legacy methods bridge to the existing REST path; the Connect methods
    // above run in parallel for the proto-migration's binary wire.

    pub async fn list_workspaces(&self) -> Result<String, String> {
        self.0.list_workspaces().await
    }

    pub async fn ensure_default_workspace(&self) -> Result<String, String> {
        self.0.ensure_default_workspace().await
    }

    pub async fn load_subtree(&self, workspace_id: String, root_id: String) -> Result<(), String> {
        self.0.load_subtree(&workspace_id, &root_id).await
    }

    pub async fn load_type_defs(&self, workspace_id: String) -> Result<(), String> {
        self.0.load_type_defs(&workspace_id).await
    }

    pub async fn catchup(&self, workspace_id: String) -> Result<(), String> {
        self.0.catchup(&workspace_id).await
    }

    pub fn apply_remote_op(&self, req_bytes: &[u8]) -> Result<(), String> {
        let req = ApplyRemoteOpRequest::decode(req_bytes).map_err(decode_err)?;
        self.0.apply_remote_op(&req.op_json)
    }

    // ── Sync getters ──

    pub fn workspaces_json(&self) -> String {
        self.0.workspaces_json()
    }

    pub fn get_block_json(&self, id: &str) -> JsValue {
        match self.0.get_block_json(id) {
            Some(s) => JsValue::from_str(&s),
            None => JsValue::NULL,
        }
    }

    pub fn list_children_json(&self, parent_id: &str) -> String {
        self.0.list_children_json(parent_id)
    }

    pub fn list_backlinks_json(&self, target_id: &str) -> String {
        self.0.list_backlinks_json(target_id)
    }

    pub fn type_defs_json(&self, workspace_id: &str) -> String {
        self.0.type_defs_json(workspace_id)
    }

    pub fn last_op_id(&self, workspace_id: &str) -> i64 {
        self.0.last_op_id(workspace_id)
    }

    pub fn set_last_op_id(&self, workspace_id: &str, id: i64) {
        self.0.set_last_op_id(workspace_id, id);
    }

    // ── Bulk state population (proto-bytes envelopes; JSON internals carried
    // through opaquely per blockstore_state.proto YAGNI rationale).

    pub fn replace_workspaces(&self, req_bytes: &[u8]) -> Result<(), String> {
        self.0.replace_workspaces(req_bytes)
    }

    pub fn upsert_workspace(&self, req_bytes: &[u8]) -> Result<(), String> {
        self.0.upsert_workspace(req_bytes)
    }

    pub fn upsert_blocks(&self, req_bytes: &[u8]) -> Result<(), String> {
        self.0.upsert_blocks(req_bytes)
    }

    pub fn upsert_refs(&self, req_bytes: &[u8]) -> Result<(), String> {
        self.0.upsert_refs(req_bytes)
    }

    pub fn project_local_ops(&self, req_bytes: &[u8]) -> Result<(), String> {
        self.0.project_local_ops(req_bytes)
    }

    pub fn blocks_json(&self) -> String {
        self.0.blocks_json()
    }
    pub fn refs_json(&self) -> String {
        self.0.refs_json()
    }
    pub fn nest_children_json(&self) -> String {
        self.0.nest_children_json()
    }
    pub fn backlinks_json(&self) -> String {
        self.0.backlinks_json()
    }
    pub fn last_op_ids_json(&self) -> String {
        self.0.last_op_ids_json()
    }
}
