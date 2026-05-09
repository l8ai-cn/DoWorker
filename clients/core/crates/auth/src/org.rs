use agentsmesh_types::Organization;

use crate::error::{parse_error_response, AuthError};
use crate::manager::AuthManager;

#[derive(serde::Deserialize)]
struct OrgsResponse {
    organizations: Vec<Organization>,
}

impl AuthManager {
    pub async fn fetch_organizations(&self) -> Result<Vec<Organization>, AuthError> {
        let auth = self.bearer_header()?;
        let resp = self
            .http
            .get(format!(
                "{}/api/v1/users/me/organizations",
                self.base_url
            ))
            .header("Authorization", auth)
            .send()
            .await?;

        if !resp.status().is_success() {
            return Err(parse_error_response(resp).await);
        }

        let wrapper: OrgsResponse = resp
            .json()
            .await
            .map_err(|e| AuthError::InvalidResponse(e.to_string()))?;
        let orgs = wrapper.organizations;

        self.replace_organizations(orgs.clone());
        Ok(orgs)
    }

    pub fn get_organizations(&self) -> Vec<Organization> {
        self.read_state().organizations.clone()
    }

    pub fn get_current_org(&self) -> Option<Organization> {
        self.read_state().current_org.clone()
    }

    pub fn switch_org(&self, slug: &str) -> Result<(), AuthError> {
        let org = self
            .read_state()
            .organizations
            .iter()
            .find(|o| o.slug == slug)
            .cloned()
            .ok_or_else(|| {
                AuthError::InvalidResponse(format!("organization '{slug}' not found"))
            })?;
        self.set_current_org(Some(org));
        Ok(())
    }
}
