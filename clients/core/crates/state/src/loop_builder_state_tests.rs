use agentsmesh_types::proto_goalloop_v1 as lp;

use crate::app_state::AppState;
use crate::loop_builder_state::LoopBuilderState;

fn valid_program() -> lp::LoopProgram {
    lp::LoopProgram {
        schema_version: 1,
        r#loop: Some(lp::LoopNodeIdentity {
            node_id: "n-checkout-fix".into(),
            local_id: "checkout-fix".into(),
        }),
        ..Default::default()
    }
}

#[test]
fn valid_compile_replaces_semantics() {
    let mut state = LoopBuilderState::new();
    state.set_source("loop source".into(), "code".into());

    state.apply_compile(lp::CompileLoopProgramResponse {
        canonical_source: "canonical source".into(),
        program: Some(valid_program()),
        diagnostics: vec![],
        revision: 1,
    });

    let snapshot = state.snapshot();
    assert_eq!(snapshot.source, "loop source");
    assert_eq!(snapshot.canonical_source, "canonical source");
    assert_eq!(snapshot.parse_status, "valid");
    assert_eq!(snapshot.active_editor, "code");
    assert_eq!(snapshot.revision, 1);
    assert_eq!(snapshot.semantic_revision, 1);
    assert_eq!(
        snapshot.program.unwrap().r#loop.unwrap().node_id,
        "n-checkout-fix"
    );
}

#[test]
fn invalid_compile_keeps_last_valid_program() {
    let mut state = LoopBuilderState::new();
    state.set_source("valid".into(), "code".into());
    state.apply_compile(lp::CompileLoopProgramResponse {
        canonical_source: "canonical".into(),
        program: Some(valid_program()),
        diagnostics: vec![],
        revision: 1,
    });
    state.set_source("loop broken {".into(), "code".into());

    state.apply_compile(lp::CompileLoopProgramResponse {
        canonical_source: String::new(),
        program: None,
        diagnostics: vec![lp::LoopDiagnostic {
            code: "loop.syntax.unexpected-token".into(),
            message: "expected }".into(),
            line: 1,
            column: 13,
            ..Default::default()
        }],
        revision: 2,
    });

    let snapshot = state.snapshot();
    assert_eq!(snapshot.source, "loop broken {");
    assert_eq!(snapshot.canonical_source, "canonical");
    assert_eq!(snapshot.parse_status, "syntax-error");
    assert_eq!(snapshot.semantic_revision, 1);
    assert_eq!(snapshot.diagnostics.len(), 1);
    assert!(snapshot.program.is_some());
}

#[test]
fn stale_compile_does_not_replace_current_source_state() {
    let mut state = LoopBuilderState::new();
    state.set_source("first".into(), "code".into());
    state.set_source("second".into(), "code".into());

    state.apply_compile(lp::CompileLoopProgramResponse {
        canonical_source: "stale".into(),
        program: Some(valid_program()),
        diagnostics: vec![],
        revision: 1,
    });

    let snapshot = state.snapshot();
    assert_eq!(snapshot.source, "second");
    assert_eq!(snapshot.parse_status, "parsing");
    assert_eq!(snapshot.semantic_revision, 0);
    assert!(snapshot.program.is_none());
}

#[test]
fn run_projection_and_org_reset_clear_state() {
    let mut app = AppState::new();
    app.loop_builder.set_source("valid".into(), "blocks".into());
    app.loop_builder.apply_run(lp::GoalLoop {
        slug: "checkout-fix".into(),
        status: "active".into(),
        ..Default::default()
    });

    assert_eq!(app.loop_builder.snapshot().run.unwrap().status, "active");
    app.reset_for_org_switch();

    let snapshot = app.loop_builder.snapshot();
    assert!(snapshot.source.is_empty());
    assert!(snapshot.run.is_none());
    assert_eq!(snapshot.revision, 0);
}
