use agentsmesh_types::proto_agent_workbench_v2 as v2;

#[test]
fn command_envelope_exposes_v2_core_command_oneof() {
    let command = v2::CommandEnvelope {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        command_id: "command-1".into(),
        payload_digest: "sha256:abc".into(),
        issued_at: "2026-07-16T00:00:00Z".into(),
        command: Some(v2::command_envelope::Command::SendPrompt(
            v2::SendPromptCommand {
                text: "build it".into(),
                ..Default::default()
            },
        )),
        ..Default::default()
    };

    assert!(matches!(
        command.command,
        Some(v2::command_envelope::Command::SendPrompt(_))
    ));
}

#[test]
fn unsupported_content_preserves_exact_identity_and_payload() {
    let unsupported = v2::UnsupportedValue {
        identity: Some(v2::ContentIdentity {
            namespace: "vendor.content".into(),
            semantic_key: "future-block".into(),
            schema_version: "3".into(),
            source_type: Some("future_block".into()),
        }),
        reason: v2::UnsupportedReason::Unknown as i32,
        payload: Some(v2::StructuredPayload {
            media_type: "application/json".into(),
            data: br#"{"exact":9007199254740993}"#.to_vec(),
        }),
    };

    assert_eq!(
        unsupported.identity.as_ref().unwrap().semantic_key,
        "future-block"
    );
    assert_eq!(
        unsupported.payload.as_ref().unwrap().data,
        br#"{"exact":9007199254740993}"#
    );
}

#[test]
fn tool_execution_keeps_exact_identity_and_results() {
    let execution = v2::ToolExecution {
        execution_id: "tool-1".into(),
        identity: Some(v2::ToolIdentity {
            namespace: "agentsmesh.acp".into(),
            semantic_key: "shell".into(),
            schema_version: "1".into(),
            source_tool_name: Some("shell".into()),
        }),
        phase: v2::ToolPhase::Running as i32,
        input: Some(v2::StructuredPayload {
            media_type: "application/json".into(),
            data: br#"{"command":"pwd"}"#.to_vec(),
        }),
        results: vec![v2::ToolResult {
            result_id: "result-1".into(),
            blocks: vec![v2::ContentBlock::default()],
            ..Default::default()
        }],
        ..Default::default()
    };

    assert_eq!(execution.identity.as_ref().unwrap().semantic_key, "shell");
    assert_eq!(execution.phase, v2::ToolPhase::Running as i32);
}

#[test]
fn artifacts_expose_revisions_provenance_grants_and_media_manifests() {
    let descriptor = v2::ArtifactDescriptor {
        artifact_id: "artifact-1".into(),
        revision: 7,
        filename: "deck.pptx".into(),
        media_type: "application/vnd.openxmlformats-officedocument.presentationml.presentation"
            .into(),
        provenance: Some(v2::ArtifactProvenance {
            command_id: Some("command-1".into()),
            ..Default::default()
        }),
        representations: vec![v2::ArtifactRepresentation {
            representation_id: "preview".into(),
            revision: 7,
            media_type: "application/pdf".into(),
            ..Default::default()
        }],
        revisions: vec![v2::ArtifactRevision {
            revision: 7,
            representation_ids: vec!["preview".into()],
            ..Default::default()
        }],
        grants: vec![v2::ArtifactGrant {
            grant_id: "grant-1".into(),
            actions: vec!["presentation.replace_slide".into()],
            ..Default::default()
        }],
        manifest: Some(v2::ArtifactManifest {
            manifest: Some(v2::artifact_manifest::Manifest::Presentation(
                v2::PresentationManifest {
                    deck_revision: 7,
                    slides: vec![v2::PresentationSlide {
                        slide_id: "slide-1".into(),
                        position: 0,
                        ..Default::default()
                    }],
                    ..Default::default()
                },
            )),
        }),
        ..Default::default()
    };

    let manifests = [
        v2::artifact_manifest::Manifest::ImageEdit(v2::ImageEditManifest {
            source_width: 1920,
            source_height: 1080,
            ..Default::default()
        }),
        v2::artifact_manifest::Manifest::Video(v2::VideoManifest {
            duration_millis: Some(12_000),
            ..Default::default()
        }),
        descriptor.manifest.unwrap().manifest.unwrap(),
    ];

    assert_eq!(descriptor.revision, 7);
    assert_eq!(manifests.len(), 3);
}

#[test]
fn session_contract_uses_v2_events_and_u64_cursors() {
    let receipt = v2::CommandReceipt {
        session_id: "session-1".into(),
        command_id: "command-1".into(),
        state: v2::CommandReceiptState::Running as i32,
        payload_digest: "sha256:abc".into(),
        ..Default::default()
    };
    let event = v2::AgentEvent {
        envelope: Some(v2::EventEnvelope {
            session_id: "session-1".into(),
            stream_epoch: "epoch-1".into(),
            revision: u64::MAX - 1,
            sequence: u64::MAX,
            item_id: "item-1".into(),
            created_at: "2026-07-16T00:00:00Z".into(),
            ..Default::default()
        }),
        event: Some(v2::agent_event::Event::CommandReceiptChanged(
            v2::CommandReceiptChanged {
                receipt: Some(receipt.clone()),
            },
        )),
    };
    let snapshot = v2::SessionSnapshot {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        revision: u64::MAX - 1,
        latest_sequence: u64::MAX,
        command_receipts: vec![receipt],
        capabilities: Some(v2::SupportCapabilities {
            protocol_version: "2".into(),
            ..Default::default()
        }),
        resources: vec![v2::SessionResource::default()],
        ..Default::default()
    };
    let batch = v2::SessionDeltaBatch {
        session_id: snapshot.session_id.clone(),
        stream_epoch: snapshot.stream_epoch.clone(),
        base_revision: snapshot.revision,
        revision: u64::MAX,
        first_sequence: u64::MAX,
        last_sequence: u64::MAX,
        events: vec![event],
        ..Default::default()
    };
    let cursor = v2::SessionCursor {
        session_id: batch.session_id,
        stream_epoch: batch.stream_epoch,
        revision: batch.revision,
        sequence: batch.last_sequence,
    };

    assert_eq!(cursor.sequence, u64::MAX);
}
