use agentsmesh_protocol::{encode_json_message, MsgType};
use serde::Serialize;

use crate::types::ControlLeaseInfo;

#[derive(Serialize)]
struct ControlLeaseRequest<'a> {
    #[serde(rename = "type")]
    msg_type: &'static str,
    action: &'static str,
    #[serde(skip_serializing_if = "Option::is_none")]
    lease_id: Option<&'a str>,
    #[serde(skip_serializing_if = "Option::is_none")]
    client_label: Option<&'a str>,
}

pub(crate) fn encode_acquire(client_label: &str) -> Option<Vec<u8>> {
    encode_request("acquire", None, Some(client_label))
}

pub(crate) fn encode_renew(lease_id: &str) -> Option<Vec<u8>> {
    encode_request("renew", Some(lease_id), None)
}

pub(crate) fn encode_release(lease_id: &str) -> Option<Vec<u8>> {
    encode_request("release", Some(lease_id), None)
}

fn encode_request(
    action: &'static str,
    lease_id: Option<&str>,
    client_label: Option<&str>,
) -> Option<Vec<u8>> {
    encode_json_message(
        MsgType::Control,
        &ControlLeaseRequest {
            msg_type: "control_lease",
            action,
            lease_id,
            client_label,
        },
    )
    .ok()
}

pub(crate) fn parse_status(value: &serde_json::Value) -> Option<ControlLeaseInfo> {
    if value.get("type")?.as_str()? != "control_lease" {
        return None;
    }
    ControlLeaseInfo::from_wire(
        value.get("status")?.as_str()?,
        value.get("lease_id").and_then(serde_json::Value::as_str),
        value.get("expires_at").and_then(serde_json::Value::as_i64),
    )
}
