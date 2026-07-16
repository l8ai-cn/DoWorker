DROP TRIGGER IF EXISTS agent_workbench_events_immutable
    ON agent_workbench_events;
DROP TRIGGER IF EXISTS agent_workbench_source_events_immutable
    ON agent_workbench_source_events;
DROP FUNCTION IF EXISTS prevent_agent_workbench_append_only_mutation();
DROP TABLE IF EXISTS agent_workbench_command_receipts;
DROP TABLE IF EXISTS agent_workbench_source_events;
DROP TABLE IF EXISTS agent_workbench_events;
DROP TABLE IF EXISTS agent_workbench_session_states;
