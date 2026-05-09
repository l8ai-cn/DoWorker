use std::collections::HashMap;

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NotificationPreference {
    pub source: Option<String>,
    pub entity_id: Option<String>,
    pub is_muted: Option<bool>,
    pub channels: Option<HashMap<String, bool>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SetNotificationPreferenceRequest {
    pub source: String,
    pub entity_id: Option<String>,
    pub is_muted: Option<bool>,
    pub channels: Option<HashMap<String, bool>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NotificationPreferenceListResponse {
    pub preferences: Vec<NotificationPreference>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn preference_decodes_backend_channels_map() {
        let backend = r#"{
            "preferences": [{
                "source": "channel",
                "entity_id": "general",
                "is_muted": false,
                "channels": {"toast": true, "browser": false}
            }]
        }"#;
        let resp: NotificationPreferenceListResponse = serde_json::from_str(backend).unwrap();
        let ch = resp.preferences[0].channels.as_ref().unwrap();
        assert_eq!(ch.get("toast"), Some(&true));
        assert_eq!(ch.get("browser"), Some(&false));
    }

    #[test]
    fn preference_relay_preserves_channels_map() {
        let backend = r#"{"source":"org","is_muted":true,"channels":{"email":true}}"#;
        let typed: NotificationPreference = serde_json::from_str(backend).unwrap();
        let relayed = serde_json::to_string(&typed).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&relayed).unwrap();
        assert_eq!(parsed["channels"]["email"], serde_json::json!(true));
    }
}
