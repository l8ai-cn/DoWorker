use wasm_bindgen::prelude::*;

use crate::api::WasmApiClient;

#[wasm_bindgen]
impl WasmApiClient {
    pub fn create_pod_service(&self) -> crate::service_pod::WasmPodService {
        crate::service_pod::WasmPodService::new(self.client.clone())
    }

    pub fn create_events_manager(&self) -> crate::events_manager::WasmEventsManager {
        crate::events_manager::WasmEventsManager::from_shared(self.runtime.events.clone())
    }

    pub fn create_ticket_service(&self) -> crate::service_ticket::WasmTicketService {
        crate::service_ticket::WasmTicketService::new(self.client.clone())
    }

    pub fn create_channel_service(&self) -> crate::service_channel::WasmChannelService {
        crate::service_channel::WasmChannelService::new(self.client.clone())
    }

    pub fn create_runner_service(&self) -> crate::service_runner::WasmRunnerService {
        crate::service_runner::WasmRunnerService::new(self.client.clone())
    }

    pub fn create_workflow_service(&self) -> crate::service_workflow::WasmWorkflowService {
        crate::service_workflow::WasmWorkflowService::new(self.client.clone())
    }

    pub fn create_goal_loop_service(&self) -> crate::service_goal_loop::WasmGoalLoopService {
        crate::service_goal_loop::WasmGoalLoopService::new(self.client.clone())
    }

    pub fn create_autopilot_service(&self) -> crate::service_autopilot::WasmAutopilotService {
        crate::service_autopilot::WasmAutopilotService::new(self.client.clone())
    }

    pub fn create_mesh_service(&self) -> crate::service_mesh::WasmMeshService {
        crate::service_mesh::WasmMeshService::new(self.client.clone())
    }

    pub fn create_blockstore_service(&self) -> crate::service_blockstore::WasmBlockstoreService {
        let state = agentsmesh_state::blockstore_state::BlockstoreState::new();
        crate::service_blockstore::WasmBlockstoreService::new(self.client.clone(), state)
    }

    pub fn create_billing_service(&self) -> crate::service_billing::WasmBillingService {
        crate::service_billing::WasmBillingService::new(self.client.clone())
    }

    pub fn create_repository_service(&self) -> crate::service_repository::WasmRepositoryService {
        crate::service_repository::WasmRepositoryService::new(self.client.clone())
    }

    pub fn create_extension_service(&self) -> crate::service_extension::WasmExtensionService {
        crate::service_extension::WasmExtensionService::new(self.client.clone())
    }

    pub fn create_invitation_service(&self) -> crate::service_invitation::WasmInvitationService {
        crate::service_invitation::WasmInvitationService::new(self.client.clone())
    }

    pub fn create_grant_service(&self) -> crate::service_grant::WasmGrantService {
        crate::service_grant::WasmGrantService::new(self.client.clone())
    }

    pub fn create_apikey_service(&self) -> crate::service_apikey::WasmApiKeyService {
        crate::service_apikey::WasmApiKeyService::new(self.client.clone())
    }

    pub fn create_binding_service(&self) -> crate::service_binding::WasmBindingService {
        crate::service_binding::WasmBindingService::new(self.client.clone())
    }

    pub fn create_knowledgebase_service(&self) -> crate::service_kb::WasmKnowledgeBaseService {
        crate::service_kb::WasmKnowledgeBaseService::new(self.client.clone())
    }

    pub fn create_notification_service(
        &self,
    ) -> crate::service_notification::WasmNotificationService {
        crate::service_notification::WasmNotificationService::new(self.client.clone())
    }

    pub fn create_promocode_service(&self) -> crate::service_promocode::WasmPromoCodeService {
        crate::service_promocode::WasmPromoCodeService::new(self.client.clone())
    }

    pub fn create_token_usage_service(&self) -> crate::service_token_usage::WasmTokenUsageService {
        crate::service_token_usage::WasmTokenUsageService::new(self.client.clone())
    }

    pub fn create_sso_service(&self) -> crate::service_sso::WasmSSOService {
        crate::service_sso::WasmSSOService::new(self.client.clone())
    }

    pub fn create_user_api_service(&self) -> crate::service_user::WasmUserApiService {
        crate::service_user::WasmUserApiService::new(self.client.clone())
    }

    pub fn create_user_credential_service(
        &self,
    ) -> crate::service_user_credential::WasmUserCredentialService {
        crate::service_user_credential::WasmUserCredentialService::new(self.client.clone())
    }

    pub fn create_env_bundle_service(&self) -> crate::service_env_bundle::WasmEnvBundleService {
        crate::service_env_bundle::WasmEnvBundleService::new(self.client.clone())
    }

    pub fn create_org_api_service(&self) -> crate::service_org::WasmOrgApiService {
        crate::service_org::WasmOrgApiService::new(self.client.clone())
    }

    pub fn create_agent_service(&self) -> crate::service_agent::WasmAgentService {
        crate::service_agent::WasmAgentService::new(self.client.clone())
    }

    pub fn create_agent_workbench_service(
        &self,
    ) -> crate::service_agent_workbench::WasmAgentWorkbenchService {
        crate::service_agent_workbench::WasmAgentWorkbenchService::new(
            self.client.clone(),
            self.runtime.state.clone(),
        )
    }

    pub fn create_ai_resource_service(&self) -> crate::service_ai_resource::WasmAIResourceService {
        crate::service_ai_resource::WasmAIResourceService::new(self.client.clone())
    }

    pub fn create_orchestration_resource_service(
        &self,
    ) -> crate::service_orchestration_resource::WasmOrchestrationResourceService {
        crate::service_orchestration_resource::WasmOrchestrationResourceService::new(
            self.client.clone(),
        )
    }

    pub fn create_execution_cluster_service(
        &self,
    ) -> crate::service_execution_cluster::WasmExecutionClusterService {
        crate::service_execution_cluster::WasmExecutionClusterService::new(self.client.clone())
    }

    pub fn create_ticket_relations_service(
        &self,
    ) -> crate::service_ticket_relations::WasmTicketRelationsService {
        crate::service_ticket_relations::WasmTicketRelationsService::new(self.client.clone())
    }

    pub fn create_file_service(&self) -> crate::service_file::WasmFileService {
        crate::service_file::WasmFileService::new(self.client.clone())
    }

    pub fn create_support_ticket_service(
        &self,
    ) -> crate::service_support_ticket::WasmSupportTicketService {
        crate::service_support_ticket::WasmSupportTicketService::new(self.client.clone())
    }

    pub fn create_auth_connect_service(
        &self,
    ) -> crate::service_auth_connect::WasmAuthConnectService {
        crate::service_auth_connect::WasmAuthConnectService::new(self.client.clone())
    }
}
