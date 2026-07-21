use agentcloud_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_state::AgentWorkbenchError;

pub(crate) fn upsert_receipt(
    snapshot: &mut v2::SessionSnapshot,
    next: &v2::CommandReceipt,
    revision: u64,
) -> Result<(), AgentWorkbenchError> {
    if !valid_receipt(next, &snapshot.session_id, revision) {
        return Err(AgentWorkbenchError::InvalidPayload {
            reason: "receipt_invalid",
        });
    }
    let Some(current) = snapshot
        .command_receipts
        .iter_mut()
        .find(|item| item.command_id == next.command_id)
    else {
        snapshot.command_receipts.push(next.clone());
        return Ok(());
    };
    if current.payload_digest != next.payload_digest {
        return Err(AgentWorkbenchError::DigestConflict {
            key: format!("command:{}", next.command_id),
        });
    }
    if current.state == next.state && terminal(current.state) {
        return if current == next {
            Ok(())
        } else {
            Err(transition(next, current.state))
        };
    }
    if current.state != next.state && !allowed_transition(current.state, next.state) {
        return Err(transition(next, current.state));
    }
    *current = next.clone();
    Ok(())
}

pub(crate) fn valid_receipt(receipt: &v2::CommandReceipt, session_id: &str, revision: u64) -> bool {
    receipt.session_id == session_id
        && !receipt.command_id.is_empty()
        && !receipt.payload_digest.is_empty()
        && receipt.state != 0
        && v2::CommandReceiptState::try_from(receipt.state).is_ok()
        && receipt
            .resulting_revision
            .is_none_or(|value| value <= revision)
}

fn transition(next: &v2::CommandReceipt, from: i32) -> AgentWorkbenchError {
    AgentWorkbenchError::ReceiptTransition {
        command_id: next.command_id.clone(),
        from,
        to: next.state,
    }
}

fn terminal(state: i32) -> bool {
    use v2::CommandReceiptState::{Cancelled, Failed, Rejected, Succeeded};
    matches!(
        v2::CommandReceiptState::try_from(state),
        Ok(Succeeded | Failed | Rejected | Cancelled)
    )
}

fn allowed_transition(from: i32, to: i32) -> bool {
    use v2::CommandReceiptState::{
        Accepted, Cancelled, Failed, Received, Rejected, Running, Succeeded,
    };
    matches!(
        (
            v2::CommandReceiptState::try_from(from),
            v2::CommandReceiptState::try_from(to)
        ),
        (Ok(Received), Ok(Accepted | Rejected))
            | (Ok(Accepted), Ok(Running | Succeeded | Failed | Cancelled))
            | (Ok(Running), Ok(Succeeded | Failed | Cancelled))
    )
}
