pub trait AuthTokenStore: Send + Sync {
    fn get_token(&self) -> Option<String>;
    fn get_refresh_token(&self) -> Option<String>;
    /// `expires_in_secs` is the server-reported lifetime for the new
    /// access token (from the `expires_in` field in the refresh response).
    /// Implementations should compute `expires_at = now + expires_in_secs`
    /// and persist it. `None` means the server didn't provide it — fall
    /// back to a conservative default (e.g. 1h) but log a warning.
    fn set_tokens(&self, token: String, refresh_token: String, expires_in_secs: Option<i64>);
    fn clear_tokens(&self);
    fn get_current_org_slug(&self) -> Option<String>;
}
