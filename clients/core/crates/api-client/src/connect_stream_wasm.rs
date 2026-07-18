#![cfg(target_arch = "wasm32")]

use futures::stream::Stream;
use prost::Message as _;
use wasm_bindgen::{JsCast, JsValue};
use wasm_bindgen_futures::{spawn_local, JsFuture};
use web_sys::{AbortController, Headers, ReadableStreamDefaultReader, RequestInit, Response};

use agentsmesh_types::proto_events_v1::{Event, SubscribeRequest};

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
        req: &SubscribeRequest,
    ) -> Result<(impl Stream<Item = Result<Event, ApiError>>, WasmAbortHandle), ApiError> {
        // Connect server-streaming wants the REQUEST body framed too —
        // even though there's only one request message. Frame is
        // `<flags=0 u8><len u32 BE><payload>`; without it the server
        // sees `protocol error: promised <random> bytes in enveloped
        // message` and the response we read back is an EOS trailer
        // carrying the error envelope (not the realtime events the
        // caller subscribed to).
        let payload = req.encode_to_vec();
        let len = u32::try_from(payload.len()).map_err(|_| {
            ApiError::Decode(format!("subscribe payload too large: {}", payload.len()))
        })?;
        let mut body_bytes = Vec::with_capacity(5 + payload.len());
        body_bytes.push(0u8); // flags: not compressed, not end-of-stream
        body_bytes.extend_from_slice(&len.to_be_bytes());
        body_bytes.extend_from_slice(&payload);
        let url = format!("{}/proto.events.v1.EventsService/Subscribe", self.base_url);

        let headers = Headers::new().map_err(js_err("Headers::new"))?;
        headers
            .set("Content-Type", "application/connect+proto")
            .map_err(js_err("Headers.set content-type"))?;
        headers
            .set("Connect-Protocol-Version", "1")
            .map_err(js_err("Headers.set connect-protocol-version"))?;
        if let Some(token) = self.auth_store.get_token() {
            headers
                .set("Authorization", &format!("Bearer {token}"))
                .map_err(js_err("Headers.set authorization"))?;
        }

        let abort_ctrl = AbortController::new().map_err(js_err("AbortController::new"))?;
        let signal = abort_ctrl.signal();

        // RequestInit: web-sys 0.3 exposes setters on the JS object via the
        // generated setter wrappers. body wants a Uint8Array (not a Rust
        // Vec) so that the browser keeps the request body buffer separately
        // from the fetch response stream — otherwise undici (Node) and
        // Chromium occasionally trip on ArrayBuffer detachment.
        let opts = RequestInit::new();
        opts.set_method("POST");
        opts.set_headers(&headers.into());
        opts.set_signal(Some(&signal));
        let body_u8 = js_sys::Uint8Array::new_with_length(body_bytes.len() as u32);
        body_u8.copy_from(&body_bytes);
        opts.set_body(&body_u8.into());

        let window =
            web_sys::window().ok_or_else(|| ApiError::Decode("wasm fetch: no window".into()))?;
        let resp_val = JsFuture::from(window.fetch_with_str_and_init(&url, &opts))
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

        let (tx, rx) = mpsc::unbounded::<Result<Bytes, ApiError>>();
        let reader = Rc::new(RefCell::new(reader));

        // Pump JS ReadableStream chunks into the mpsc channel on the
        // local task queue. The spawn_local task owns the reader; once
        // the receiver drops, send fails and the loop exits, releasing
        // the reader. AbortController triggers the JS-side error path
        // which surfaces here as a reader.read() rejection.
        let reader_pump = reader.clone();
        spawn_local(async move {
            loop {
                let read_promise = reader_pump.borrow().read();
                let chunk_obj = match JsFuture::from(read_promise).await {
                    Ok(v) => v,
                    Err(e) => {
                        let _ = tx.unbounded_send(Err(ApiError::Decode(format!(
                            "reader.read rejected: {e:?}"
                        ))));
                        return;
                    }
                };

                let done_v = js_sys::Reflect::get(&chunk_obj, &JsValue::from_str("done"))
                    .unwrap_or(JsValue::FALSE);
                if done_v.as_bool().unwrap_or(false) {
                    return;
                }

                let value_v = match js_sys::Reflect::get(&chunk_obj, &JsValue::from_str("value")) {
                    Ok(v) => v,
                    Err(e) => {
                        let _ = tx.unbounded_send(Err(ApiError::Decode(format!(
                            "reader.read value missing: {e:?}"
                        ))));
                        return;
                    }
                };
                let arr = js_sys::Uint8Array::new(&value_v);
                let mut buf = vec![0u8; arr.length() as usize];
                arr.copy_to(&mut buf);
                if tx.unbounded_send(Ok(Bytes::from(buf))).is_err() {
                    // Consumer dropped the stream — we're done.
                    return;
                }
            }
        });

        let frames = parse_connect_frames::<_, Event>(rx);
        Ok((frames, WasmAbortHandle { ctrl: abort_ctrl }))
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
