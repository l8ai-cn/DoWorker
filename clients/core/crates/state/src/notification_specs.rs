#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ToastSpec {
    pub kind: String,
    pub title_key: String,
    #[serde(default)]
    pub title_params: serde_json::Value,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub duration_ms: u32,
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct NotificationSpec {
    pub title: String,
    pub body: String,
    #[serde(default)]
    pub icon: Option<String>,
    #[serde(default)]
    pub link: Option<String>,
}
