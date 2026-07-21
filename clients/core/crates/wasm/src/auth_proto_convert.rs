// auth.rs proto-bytes mutator support. AuthManager's renderer-facing JSON
// mutators (apply_session / set_organizations / set_current_org) have been
// flipped to proto-bytes; these helpers down-project the proto wire shape
// to AuthManager's serde DTOs (auth_types::*) so the in-memory cache stays
// on the legacy DTO and the on-disk persistence format doesn't churn.

use agentcloud_state::auth_types::{Organization as AuthOrg, User as AuthUser};
use agentcloud_types::{proto_auth_v1 as auth_proto, proto_org_v1 as org_proto};

pub(crate) fn user_from_proto(u: &auth_proto::User) -> AuthUser {
    AuthUser {
        id: u.id,
        email: u.email.clone(),
        username: u.username.clone(),
        name: u.name.clone(),
        avatar_url: u.avatar_url.clone(),
        is_email_verified: u.is_email_verified,
    }
}

pub(crate) fn org_from_proto(o: &org_proto::Organization) -> AuthOrg {
    let subscription_plan = if o.subscription_plan.is_empty() {
        None
    } else {
        Some(o.subscription_plan.clone())
    };
    let subscription_status = if o.subscription_status.is_empty() {
        None
    } else {
        Some(o.subscription_status.clone())
    };
    AuthOrg {
        id: o.id,
        name: o.name.clone(),
        slug: o.slug.clone(),
        role: o.role.clone(),
        logo_url: o.logo_url.clone(),
        subscription_plan,
        subscription_status,
    }
}
