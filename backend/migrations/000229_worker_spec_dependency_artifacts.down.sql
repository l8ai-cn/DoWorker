DROP TRIGGER IF EXISTS worker_spec_dependency_artifacts_immutable
    ON worker_spec_dependency_artifacts;
DROP FUNCTION IF EXISTS prevent_worker_spec_dependency_artifact_update();
DROP TABLE IF EXISTS worker_spec_dependency_artifacts;
