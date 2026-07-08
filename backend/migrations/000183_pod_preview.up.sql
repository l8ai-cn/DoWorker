-- Gateway HTTP preview: expose a pod's loopback HTTP service through the
-- Gateway /preview/{podKey}/* data plane. preview_port==0 disables preview.
ALTER TABLE pods ADD COLUMN preview_port INT NOT NULL DEFAULT 0;
ALTER TABLE pods ADD COLUMN preview_path VARCHAR(255) NOT NULL DEFAULT '';
