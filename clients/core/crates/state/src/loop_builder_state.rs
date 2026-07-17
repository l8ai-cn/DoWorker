use agentsmesh_types::proto_goalloop_v1 as lp;

pub struct LoopBuilderState {
    source: String,
    canonical_source: String,
    program: Option<lp::LoopProgram>,
    diagnostics: Vec<lp::LoopDiagnostic>,
    parse_status: String,
    active_editor: String,
    revision: u64,
    semantic_revision: u64,
    run: Option<lp::GoalLoop>,
}

impl LoopBuilderState {
    pub fn new() -> Self {
        Self {
            source: String::new(),
            canonical_source: String::new(),
            program: None,
            diagnostics: Vec::new(),
            parse_status: "empty".into(),
            active_editor: "blocks".into(),
            revision: 0,
            semantic_revision: 0,
            run: None,
        }
    }

    pub fn set_source(&mut self, source: String, active_editor: String) {
        self.source = source;
        self.active_editor = active_editor;
        self.parse_status = "parsing".into();
        self.revision += 1;
    }

    pub fn set_active_editor(&mut self, active_editor: String) {
        self.active_editor = active_editor;
    }

    pub fn apply_compile(&mut self, response: lp::CompileLoopProgramResponse) {
        if response.revision != self.revision {
            return;
        }
        self.diagnostics = response.diagnostics;
        let is_valid = self.diagnostics.is_empty() && response.program.is_some();
        if !is_valid {
            self.parse_status = "syntax-error".into();
            return;
        }

        self.canonical_source = response.canonical_source;
        self.program = response.program;
        self.parse_status = "valid".into();
        self.semantic_revision += 1;
    }

    pub fn apply_ai_draft(&mut self, response: lp::CompileLoopProgramResponse) -> bool {
        if response.revision != self.revision
            || !response.diagnostics.is_empty()
            || response.program.is_none()
            || response.canonical_source.trim().is_empty()
        {
            return false;
        }

        self.source = response.canonical_source.clone();
        self.canonical_source = response.canonical_source;
        self.program = response.program;
        self.diagnostics.clear();
        self.parse_status = "valid".into();
        self.revision += 1;
        self.semantic_revision += 1;
        true
    }

    pub fn apply_run(&mut self, run: lp::GoalLoop) {
        self.run = Some(run);
    }

    pub fn snapshot(&self) -> lp::LoopDraftSnapshot {
        lp::LoopDraftSnapshot {
            source: self.source.clone(),
            canonical_source: self.canonical_source.clone(),
            program: self.program.clone(),
            diagnostics: self.diagnostics.clone(),
            parse_status: self.parse_status.clone(),
            active_editor: self.active_editor.clone(),
            revision: self.revision,
            semantic_revision: self.semantic_revision,
            run: self.run.clone(),
        }
    }

    pub fn reset(&mut self) {
        *self = Self::new();
    }
}

impl Default for LoopBuilderState {
    fn default() -> Self {
        Self::new()
    }
}
