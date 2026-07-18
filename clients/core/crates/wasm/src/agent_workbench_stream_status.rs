#[derive(Debug, Default)]
pub(crate) struct AgentWorkbenchStreamStatus {
    terminal: StreamTerminal,
    error: Option<String>,
}

#[derive(Debug, Default)]
enum StreamTerminal {
    #[default]
    Open,
    ClientClosed,
    RemoteClosed,
    Failed,
}

impl AgentWorkbenchStreamStatus {
    pub(crate) fn code(&self) -> &'static str {
        match self.terminal {
            StreamTerminal::Open => "open",
            StreamTerminal::ClientClosed => "client_closed",
            StreamTerminal::RemoteClosed => "remote_closed",
            StreamTerminal::Failed => "failed",
        }
    }

    pub(crate) fn error(&self) -> Option<&str> {
        self.error.as_deref()
    }

    pub(crate) fn mark_client_closed(&mut self) -> bool {
        self.transition(StreamTerminal::ClientClosed, None)
    }

    pub(crate) fn mark_remote_closed(&mut self) -> bool {
        self.transition(StreamTerminal::RemoteClosed, None)
    }

    pub(crate) fn mark_failed(&mut self, error: String) -> bool {
        self.transition(StreamTerminal::Failed, Some(error))
    }

    fn transition(&mut self, terminal: StreamTerminal, error: Option<String>) -> bool {
        if !matches!(self.terminal, StreamTerminal::Open) {
            return false;
        }
        self.terminal = terminal;
        self.error = error;
        true
    }
}
