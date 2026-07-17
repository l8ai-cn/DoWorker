#![cfg(not(target_arch = "wasm32"))]

use bytes::Bytes;
use futures::stream::Stream;
use futures::TryStreamExt;
use prost::Message;
use reqwest::header::{HeaderName, HeaderValue};

use crate::client::ApiClient;
use crate::connect_stream_frames::parse_connect_frames;
use crate::connect_stream_request::frame_connect_stream_request;
use crate::error::ApiError;

impl ApiClient {
    pub async fn connect_server_stream_native<Req, Res>(
        &self,
        procedure: &str,
        request: &Req,
    ) -> Result<impl Stream<Item = Result<Res, ApiError>>, ApiError>
    where
        Req: Message,
        Res: Message + Default + 'static,
    {
        let bearer_token = self.auth_store.get_token();
        self.connect_server_stream_native_with_optional_bearer(
            procedure,
            request,
            bearer_token.as_deref(),
        )
        .await
    }

    pub(crate) async fn connect_server_stream_native_with_bearer<Req, Res>(
        &self,
        procedure: &str,
        request: &Req,
        bearer_token: &str,
    ) -> Result<impl Stream<Item = Result<Res, ApiError>>, ApiError>
    where
        Req: Message,
        Res: Message + Default + 'static,
    {
        self.connect_server_stream_native_with_optional_bearer(
            procedure,
            request,
            Some(bearer_token),
        )
        .await
    }

    async fn connect_server_stream_native_with_optional_bearer<Req, Res>(
        &self,
        procedure: &str,
        request: &Req,
        bearer_token: Option<&str>,
    ) -> Result<impl Stream<Item = Result<Res, ApiError>>, ApiError>
    where
        Req: Message,
        Res: Message + Default + 'static,
    {
        let url = format!("{}{}", self.base_url, procedure);
        let mut builder = self
            .http
            .post(&url)
            .header(
                HeaderName::from_static("content-type"),
                HeaderValue::from_static("application/connect+proto"),
            )
            .header(
                HeaderName::from_static("connect-protocol-version"),
                HeaderValue::from_static("1"),
            )
            .body(frame_connect_stream_request(request)?);

        if let Some(token) = bearer_token {
            let value = HeaderValue::from_str(&format!("Bearer {token}")).map_err(|error| {
                ApiError::Decode(format!("invalid authorization header: {error}"))
            })?;
            builder = builder.header(HeaderName::from_static("authorization"), value);
        }

        let response = builder.send().await?;
        let status = response.status();
        if !status.is_success() {
            if status.as_u16() == 401 {
                return Err(ApiError::AuthExpired);
            }
            let status_code = status.as_u16();
            let status_text = status.canonical_reason().unwrap_or("Unknown").to_string();
            let body = response.bytes().await.ok();
            let server_message = body
                .as_ref()
                .and_then(|value| std::str::from_utf8(value).ok())
                .filter(|value| !value.is_empty())
                .map(String::from);
            return Err(ApiError::Http {
                status: status_code,
                status_text,
                code: None,
                server_message,
                data: None,
                url: Some(url),
            });
        }

        let chunks = response
            .bytes_stream()
            .map_err(|error| ApiError::Http {
                status: 0,
                status_text: format!("transport: {error}"),
                code: None,
                server_message: None,
                data: None,
                url: None,
            })
            .map_ok(Bytes::from);
        Ok(parse_connect_frames::<_, Res>(Box::pin(chunks)))
    }
}
