use std::collections::{HashMap, HashSet, VecDeque};

use agentsmesh_types::proto_agent_workbench_v2 as v2;
use thiserror::Error;

use crate::agent_workbench_reducer::apply_batch;
use crate::agent_workbench_snapshot::same_canonical_content;
use crate::agent_workbench_validation::{invalid, validate_batch, validate_snapshot};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ProjectionStatus {
    Ready,
    ResyncRequired,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ResyncReason {
    StreamEpochChanged,
    BaseRevisionMismatch,
    SequenceGap,
    DigestConflict,
}

#[derive(Debug, Error, Clone, PartialEq, Eq)]
pub enum AgentWorkbenchError {
    #[error("session {session_id} requires resync: {reason:?}")]
    ResyncRequired {
        session_id: String,
        reason: ResyncReason,
    },
    #[error("digest conflict for {key}")]
    DigestConflict { key: String },
    #[error("invalid receipt transition for {command_id}: {from} -> {to}")]
    ReceiptTransition {
        command_id: String,
        from: i32,
        to: i32,
    },
    #[error("invalid agent workbench payload: {reason}")]
    InvalidPayload { reason: &'static str },
}

#[derive(Debug, Clone)]
pub struct AgentWorkbenchSession {
    pub snapshot: v2::SessionSnapshot,
    pub commit_revision: u64,
    pub status: ProjectionStatus,
    pub resync_reason: Option<ResyncReason>,
    applied_batches: VecDeque<((String, u64, u64, u64), String)>,
    seen_epochs: HashSet<String>,
}

#[derive(Debug, Default)]
pub struct AgentWorkbenchState {
    sessions: HashMap<String, AgentWorkbenchSession>,
}

impl AgentWorkbenchState {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn get_session(&self, session_id: &str) -> Option<&AgentWorkbenchSession> {
        self.sessions.get(session_id)
    }

    pub fn revision(&self, session_id: &str) -> Option<u64> {
        Some(self.sessions.get(session_id)?.commit_revision)
    }

    pub fn apply_snapshot(
        &mut self,
        snapshot: &v2::SessionSnapshot,
    ) -> Result<(), AgentWorkbenchError> {
        validate_snapshot(snapshot)?;
        let Some(current) = self.sessions.get_mut(&snapshot.session_id) else {
            let session = session(snapshot, 1, HashSet::from([snapshot.stream_epoch.clone()]));
            self.sessions.insert(snapshot.session_id.clone(), session);
            return Ok(());
        };
        let epoch_changed = current.snapshot.stream_epoch != snapshot.stream_epoch;
        if epoch_changed {
            if current.seen_epochs.contains(&snapshot.stream_epoch) {
                return Err(invalid("snapshot_epoch_stale"));
            }
        } else {
            if snapshot.revision < current.snapshot.revision
                || snapshot.latest_sequence < current.snapshot.latest_sequence
            {
                return Err(invalid("snapshot_stale"));
            }
            if snapshot.revision == current.snapshot.revision
                && snapshot.latest_sequence == current.snapshot.latest_sequence
            {
                if !same_canonical_content(&current.snapshot, snapshot) {
                    return Err(invalid("snapshot_cursor_conflict"));
                }
                if current.status == ProjectionStatus::Ready {
                    if current.snapshot != *snapshot {
                        current.commit_revision = current
                            .commit_revision
                            .checked_add(1)
                            .ok_or(invalid("commit_revision_overflow"))?;
                        current.snapshot = snapshot.clone();
                    }
                    return Ok(());
                }
            }
        }
        let commit_revision = current
            .commit_revision
            .checked_add(1)
            .ok_or(invalid("commit_revision_overflow"))?;
        let mut seen_epochs = current.seen_epochs.clone();
        if epoch_changed {
            seen_epochs.insert(snapshot.stream_epoch.clone());
        }
        *current = session(snapshot, commit_revision, seen_epochs);
        Ok(())
    }

    pub fn apply_delta_batch(
        &mut self,
        batch: &v2::SessionDeltaBatch,
    ) -> Result<(), AgentWorkbenchError> {
        validate_batch(batch)?;
        let session = self.sessions.get_mut(&batch.session_id);
        let session = session.ok_or(invalid("session_snapshot_missing"))?;
        let identity = (
            batch.stream_epoch.clone(),
            batch.revision,
            batch.first_sequence,
            batch.last_sequence,
        );
        let batches = &session.applied_batches;
        let applied = batches.iter().find(|(id, _)| id == &identity);
        if let Some((_, digest)) = applied {
            return if digest == &batch.digest {
                Ok(())
            } else {
                session.status = ProjectionStatus::ResyncRequired;
                session.resync_reason = Some(ResyncReason::DigestConflict);
                Err(AgentWorkbenchError::DigestConflict {
                    key: format!("batch:{identity:?}"),
                })
            };
        }
        if let Some(reason) = session.resync_reason {
            return mark_resync(session, reason);
        }
        if batch.stream_epoch != session.snapshot.stream_epoch {
            return mark_resync(session, ResyncReason::StreamEpochChanged);
        }
        if batch.base_revision != session.snapshot.revision {
            return mark_resync(session, ResyncReason::BaseRevisionMismatch);
        }
        if session.snapshot.latest_sequence.checked_add(1) != Some(batch.first_sequence) {
            return mark_resync(session, ResyncReason::SequenceGap);
        }

        let mut snapshot = session.snapshot.clone();
        apply_batch(&mut snapshot, batch)?;
        session.commit_revision = session
            .commit_revision
            .checked_add(1)
            .ok_or(invalid("commit_revision_overflow"))?;
        snapshot.revision = batch.revision;
        snapshot.latest_sequence = batch.last_sequence;
        session.snapshot = snapshot;
        let digest = batch.digest.clone();
        session.applied_batches.push_back((identity, digest));
        if session.applied_batches.len() > 256 {
            session.applied_batches.pop_front();
        }
        Ok(())
    }
}

fn session(
    snapshot: &v2::SessionSnapshot,
    commit_revision: u64,
    seen_epochs: HashSet<String>,
) -> AgentWorkbenchSession {
    AgentWorkbenchSession {
        snapshot: snapshot.clone(),
        commit_revision,
        status: ProjectionStatus::Ready,
        resync_reason: None,
        applied_batches: VecDeque::new(),
        seen_epochs,
    }
}

fn mark_resync(
    session: &mut AgentWorkbenchSession,
    reason: ResyncReason,
) -> Result<(), AgentWorkbenchError> {
    session.status = ProjectionStatus::ResyncRequired;
    session.resync_reason = Some(reason);
    Err(AgentWorkbenchError::ResyncRequired {
        session_id: session.snapshot.session_id.clone(),
        reason,
    })
}
