#![cfg(target_arch = "wasm32")]

use futures::stream::Stream;
use prost::Message;
use wasm_bindgen::JsCast;
use wasm_bindgen_futures::JsFuture;
use web_sys::{AbortController, Headers, ReadableStreamDefaultReader, RequestInit, Response};

use crate::client::ApiClient;
use crate::connect_stream_frames::parse_connect_frames;
use crate::connect_stream_request::frame_connect_stream_request;
use crate::connect_stream_wasm_reader::{js_error, read_chunks};
use crate::error::ApiError;

pub struct WasmAbortHandle {
    controller: AbortController,
}

impl WasmAbortHandle {
    pub fn abort(&self) {
        self.controller.abort();
    }
}

impl Drop for WasmAbortHandle {
    fn drop(&mut self) {
        self.controller.abort();
    }
}

impl ApiClient {
    pub async fn connect_server_stream_wasm<Req, Res>(
        &self,
        procedure: &str,
        request: &Req,
    ) -> Result<(impl Stream<Item = Result<Res, ApiError>>, WasmAbortHandle), ApiError>
    where
        Req: Message,
        Res: Message + Default + 'static,
    {
        let bearer_token = self.auth_store.get_token();
        self.connect_server_stream_wasm_with_optional_bearer(
            procedure,
            request,
            bearer_token.as_deref(),
        )
        .await
    }

    pub(crate) async fn connect_server_stream_wasm_with_bearer<Req, Res>(
        &self,
        procedure: &str,
        request: &Req,
        bearer_token: &str,
    ) -> Result<(impl Stream<Item = Result<Res, ApiError>>, WasmAbortHandle), ApiError>
    where
        Req: Message,
        Res: Message + Default + 'static,
    {
        self.connect_server_stream_wasm_with_optional_bearer(procedure, request, Some(bearer_token))
            .await
    }

    async fn connect_server_stream_wasm_with_optional_bearer<Req, Res>(
        &self,
        procedure: &str,
        request: &Req,
        bearer_token: Option<&str>,
    ) -> Result<(impl Stream<Item = Result<Res, ApiError>>, WasmAbortHandle), ApiError>
    where
        Req: Message,
        Res: Message + Default + 'static,
    {
        let url = format!("{}{}", self.base_url, procedure);
        let headers = stream_headers(bearer_token)?;
        let abort_controller = AbortController::new().map_err(js_error("AbortController::new"))?;
        let options = RequestInit::new();
        options.set_method("POST");
        options.set_headers(&headers.into());
        options.set_signal(Some(&abort_controller.signal()));
        let body = frame_connect_stream_request(request)?;
        let body_array = js_sys::Uint8Array::new_with_length(body.len() as u32);
        body_array.copy_from(&body);
        options.set_body(&body_array.into());

        let window =
            web_sys::window().ok_or_else(|| ApiError::Decode("wasm fetch: no window".into()))?;
        let value = JsFuture::from(window.fetch_with_str_and_init(&url, &options))
            .await
            .map_err(js_error("fetch"))?;
        let response: Response = value
            .dyn_into()
            .map_err(|_| ApiError::Decode("wasm fetch: response not a Response".into()))?;
        if !response.ok() {
            if response.status() == 401 {
                return Err(ApiError::AuthExpired);
            }
            return Err(ApiError::Http {
                status: response.status(),
                status_text: response.status_text(),
                code: None,
                server_message: None,
                data: None,
                url: Some(url),
            });
        }

        let body = response
            .body()
            .ok_or_else(|| ApiError::Decode("wasm fetch: response has no body".into()))?;
        let reader: ReadableStreamDefaultReader = body.get_reader().dyn_into().map_err(|_| {
            ApiError::Decode("wasm fetch: get_reader did not return a default reader".into())
        })?;
        let chunks = read_chunks(reader);
        let frames = parse_connect_frames::<_, Res>(chunks);
        Ok((
            frames,
            WasmAbortHandle {
                controller: abort_controller,
            },
        ))
    }
}

fn stream_headers(bearer_token: Option<&str>) -> Result<Headers, ApiError> {
    let headers = Headers::new().map_err(js_error("Headers::new"))?;
    headers
        .set("Content-Type", "application/connect+proto")
        .map_err(js_error("Headers.set content-type"))?;
    headers
        .set("Connect-Protocol-Version", "1")
        .map_err(js_error("Headers.set connect-protocol-version"))?;
    if let Some(token) = bearer_token {
        headers
            .set("Authorization", &format!("Bearer {token}"))
            .map_err(js_error("Headers.set authorization"))?;
    }
    Ok(headers)
}
