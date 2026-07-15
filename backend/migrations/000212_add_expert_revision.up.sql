ALTER TABLE experts
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1;

ALTER TABLE experts
    ADD CONSTRAINT chk_experts_revision_positive
    CHECK (revision > 0);
