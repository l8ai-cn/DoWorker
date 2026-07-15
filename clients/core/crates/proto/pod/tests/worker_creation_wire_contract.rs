use pod_proto::proto::pod::v1::{WorkerToolModelRequirement, WorkerTypeOption};
use prost::Message;

#[test]
fn preserves_requires_model_resource_on_worker_type_options() {
    let encoded = [0x40, 0x01];
    let option = WorkerTypeOption::decode(encoded.as_slice())
        .expect("worker type option should decode");

    assert!(
        option.requires_model_resource,
        "the Worker creation model-resource contract must survive Rust Core"
    );
    assert_eq!(option.encode_to_vec(), encoded);
}

#[test]
fn preserves_tool_model_requirements_on_worker_type_options() {
    let option = WorkerTypeOption {
        tool_model_requirements: vec![WorkerToolModelRequirement {
            role: "seedance-video".into(),
            provider_keys: vec!["doubao".into()],
            protocol_adapters: vec!["openai-compatible".into()],
            modality: "video".into(),
            capability: "video-generation".into(),
        }],
        ..Default::default()
    };

    let decoded = WorkerTypeOption::decode(option.encode_to_vec().as_slice())
        .expect("worker tool model requirement should decode");

    assert_eq!(decoded.tool_model_requirements, option.tool_model_requirements);
}
