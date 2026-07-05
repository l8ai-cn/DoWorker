// Connect-RPC bridge methods for WasmKnowledgeBaseService. Binary in,
// binary out (conventions §2.5): TS encodes via @bufbuild/protobuf
// .toBinary(), receives a Uint8Array back, decodes via .fromBinary().
// Split from service_kb.rs to honor the 200-line/file limit.

use agentsmesh_types::proto_knowledgebase_v1 as kb;
use prost::Message;
use wasm_bindgen::prelude::*;

use crate::service_kb::WasmKnowledgeBaseService;

#[wasm_bindgen]
impl WasmKnowledgeBaseService {
    #[wasm_bindgen(js_name = listKnowledgeBasesConnect)]
    pub async fn list_knowledge_bases_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::ListKnowledgeBasesRequest::decode(request)
            .map_err(|e| format!("decode ListKnowledgeBasesRequest: {e}"))?;
        let resp = self
            .client_ref()
            .list_knowledge_bases_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = getKnowledgeBaseConnect)]
    pub async fn get_knowledge_base_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::GetKnowledgeBaseRequest::decode(request)
            .map_err(|e| format!("decode GetKnowledgeBaseRequest: {e}"))?;
        let resp = self
            .client_ref()
            .get_knowledge_base_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = createKnowledgeBaseConnect)]
    pub async fn create_knowledge_base_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::CreateKnowledgeBaseRequest::decode(request)
            .map_err(|e| format!("decode CreateKnowledgeBaseRequest: {e}"))?;
        let resp = self
            .client_ref()
            .create_knowledge_base_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = updateKnowledgeBaseConnect)]
    pub async fn update_knowledge_base_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::UpdateKnowledgeBaseRequest::decode(request)
            .map_err(|e| format!("decode UpdateKnowledgeBaseRequest: {e}"))?;
        let resp = self
            .client_ref()
            .update_knowledge_base_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = deleteKnowledgeBaseConnect)]
    pub async fn delete_knowledge_base_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::DeleteKnowledgeBaseRequest::decode(request)
            .map_err(|e| format!("decode DeleteKnowledgeBaseRequest: {e}"))?;
        let resp = self
            .client_ref()
            .delete_knowledge_base_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = setAgentMountsConnect)]
    pub async fn set_agent_mounts_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::SetAgentMountsRequest::decode(request)
            .map_err(|e| format!("decode SetAgentMountsRequest: {e}"))?;
        let resp = self
            .client_ref()
            .set_kb_agent_mounts_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = listAgentMountsConnect)]
    pub async fn list_agent_mounts_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::ListAgentMountsRequest::decode(request)
            .map_err(|e| format!("decode ListAgentMountsRequest: {e}"))?;
        let resp = self
            .client_ref()
            .list_kb_agent_mounts_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = getKnowledgeBaseFileConnect)]
    pub async fn get_knowledge_base_file_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::GetKnowledgeBaseFileRequest::decode(request)
            .map_err(|e| format!("decode GetKnowledgeBaseFileRequest: {e}"))?;
        let resp = self
            .client_ref()
            .get_knowledge_base_file_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }

    #[wasm_bindgen(js_name = listKnowledgeBaseDirConnect)]
    pub async fn list_knowledge_base_dir_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = kb::ListKnowledgeBaseDirRequest::decode(request)
            .map_err(|e| format!("decode ListKnowledgeBaseDirRequest: {e}"))?;
        let resp = self
            .client_ref()
            .list_knowledge_base_dir_connect(&req)
            .await
            .map_err(agentsmesh_services::wire)?;
        Ok(resp.encode_to_vec())
    }
}
