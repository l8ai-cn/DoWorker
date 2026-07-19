-- Forward-only repair for deployments that have already reached 000229.
-- Down intentionally retains repaired contracts; reverting them would
-- reintroduce the lineage violation and can make workers non-dispatchable.
SELECT 1;
