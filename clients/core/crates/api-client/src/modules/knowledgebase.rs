// proto.knowledgebase.v1.KnowledgeBaseService Connect-RPC client bindings.
// Procedure paths derive from `proto.knowledgebase.v1.KnowledgeBaseService.<Method>`.

use crate::ApiClient;
use crate::connect_call::connect_call;
use crate::error::ApiError;
use agentsmesh_types::proto_knowledgebase_v1 as kb;

const SVC: &str = "/proto.knowledgebase.v1.KnowledgeBaseService";

impl ApiClient {
    pub async fn list_knowledge_bases_connect(
        &self,
        req: &kb::ListKnowledgeBasesRequest,
    ) -> Result<kb::ListKnowledgeBasesResponse, ApiError> {
        connect_call(self, &format!("{SVC}/ListKnowledgeBases"), req).await
    }

    pub async fn get_knowledge_base_connect(
        &self,
        req: &kb::GetKnowledgeBaseRequest,
    ) -> Result<kb::KnowledgeBase, ApiError> {
        connect_call(self, &format!("{SVC}/GetKnowledgeBase"), req).await
    }

    pub async fn create_knowledge_base_connect(
        &self,
        req: &kb::CreateKnowledgeBaseRequest,
    ) -> Result<kb::KnowledgeBase, ApiError> {
        connect_call(self, &format!("{SVC}/CreateKnowledgeBase"), req).await
    }

    pub async fn update_knowledge_base_connect(
        &self,
        req: &kb::UpdateKnowledgeBaseRequest,
    ) -> Result<kb::KnowledgeBase, ApiError> {
        connect_call(self, &format!("{SVC}/UpdateKnowledgeBase"), req).await
    }

    pub async fn delete_knowledge_base_connect(
        &self,
        req: &kb::DeleteKnowledgeBaseRequest,
    ) -> Result<kb::DeleteKnowledgeBaseResponse, ApiError> {
        connect_call(self, &format!("{SVC}/DeleteKnowledgeBase"), req).await
    }

    pub async fn set_kb_agent_mounts_connect(
        &self,
        req: &kb::SetAgentMountsRequest,
    ) -> Result<kb::SetAgentMountsResponse, ApiError> {
        connect_call(self, &format!("{SVC}/SetAgentMounts"), req).await
    }

    pub async fn list_kb_agent_mounts_connect(
        &self,
        req: &kb::ListAgentMountsRequest,
    ) -> Result<kb::ListAgentMountsResponse, ApiError> {
        connect_call(self, &format!("{SVC}/ListAgentMounts"), req).await
    }

    pub async fn get_knowledge_base_file_connect(
        &self,
        req: &kb::GetKnowledgeBaseFileRequest,
    ) -> Result<kb::KnowledgeBaseFile, ApiError> {
        connect_call(self, &format!("{SVC}/GetKnowledgeBaseFile"), req).await
    }

    pub async fn list_knowledge_base_dir_connect(
        &self,
        req: &kb::ListKnowledgeBaseDirRequest,
    ) -> Result<kb::ListKnowledgeBaseDirResponse, ApiError> {
        connect_call(self, &format!("{SVC}/ListKnowledgeBaseDir"), req).await
    }
}
