use agentcloud_types::proto_events_v1::{Event, SubscribeRequest};
use futures::stream::Stream;

use crate::{ApiClient, ApiError};

const EVENTS_SUBSCRIBE: &str = "/proto.events.v1.EventsService/Subscribe";

impl ApiClient {
    #[cfg(not(target_arch = "wasm32"))]
    pub async fn subscribe_events_connect_native(
        &self,
        request: &SubscribeRequest,
    ) -> Result<impl Stream<Item = Result<Event, ApiError>>, ApiError> {
        self.connect_server_stream_native(EVENTS_SUBSCRIBE, request)
            .await
    }

    #[cfg(target_arch = "wasm32")]
    pub async fn subscribe_events_connect_wasm(
        &self,
        request: &SubscribeRequest,
    ) -> Result<
        (
            impl Stream<Item = Result<Event, ApiError>>,
            crate::WasmAbortHandle,
        ),
        ApiError,
    > {
        self.connect_server_stream_wasm(EVENTS_SUBSCRIBE, request)
            .await
    }
}
