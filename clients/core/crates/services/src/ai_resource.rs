use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_types::proto_ai_resource_v1 as ai;
use prost::Message;

pub struct AIResourceService {
    client: Arc<ApiClient>,
}

macro_rules! wire_rpc {
    ($name:ident, $request:ty, $client_method:ident) => {
        pub async fn $name(&self, bytes: &[u8]) -> Result<Vec<u8>, String> {
            let request = <$request>::decode(bytes)
                .map_err(|error| format!("decode {} request: {error}", stringify!($name)))?;
            let response = self
                .client
                .$client_method(&request)
                .await
                .map_err(crate::wire)?;
            Ok(response.encode_to_vec())
        }
    };
}

impl AIResourceService {
    pub fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    wire_rpc!(
        get_catalog_connect,
        ai::GetCatalogRequest,
        get_ai_resource_catalog_connect
    );
    wire_rpc!(
        list_personal_connections_connect,
        ai::ListPersonalConnectionsRequest,
        list_personal_connections_connect
    );
    wire_rpc!(
        list_organization_connections_connect,
        ai::ListOrganizationConnectionsRequest,
        list_organization_connections_connect
    );
    wire_rpc!(
        list_personal_effective_resources_connect,
        ai::ListPersonalEffectiveResourcesRequest,
        list_personal_effective_resources_connect
    );
    wire_rpc!(
        list_organization_effective_resources_connect,
        ai::ListOrganizationEffectiveResourcesRequest,
        list_organization_effective_resources_connect
    );
    wire_rpc!(
        create_personal_connection_connect,
        ai::CreatePersonalConnectionRequest,
        create_personal_connection_connect
    );
    wire_rpc!(
        create_organization_connection_connect,
        ai::CreateOrganizationConnectionRequest,
        create_organization_connection_connect
    );
    wire_rpc!(
        update_connection_connect,
        ai::UpdateConnectionRequest,
        update_ai_resource_connection_connect
    );
    wire_rpc!(
        rotate_connection_credentials_connect,
        ai::RotateConnectionCredentialsRequest,
        rotate_connection_credentials_connect
    );
    wire_rpc!(
        set_connection_enabled_connect,
        ai::SetConnectionEnabledRequest,
        set_connection_enabled_connect
    );
    wire_rpc!(
        validate_connection_connect,
        ai::ValidateConnectionRequest,
        validate_ai_resource_connection_connect
    );
    wire_rpc!(
        delete_connection_connect,
        ai::DeleteConnectionRequest,
        delete_ai_resource_connection_connect
    );
    wire_rpc!(
        create_resource_connect,
        ai::CreateResourceRequest,
        create_ai_resource_connect
    );
    wire_rpc!(
        update_resource_connect,
        ai::UpdateResourceRequest,
        update_ai_resource_connect
    );
    wire_rpc!(
        set_resource_enabled_connect,
        ai::SetResourceEnabledRequest,
        set_ai_resource_enabled_connect
    );
    wire_rpc!(
        delete_resource_connect,
        ai::DeleteResourceRequest,
        delete_ai_resource_connect
    );
    wire_rpc!(
        set_default_connect,
        ai::SetDefaultRequest,
        set_ai_resource_default_connect
    );
}
