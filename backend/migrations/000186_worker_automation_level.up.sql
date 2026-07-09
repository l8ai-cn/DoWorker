-- Unified cross-agent worker automation/permission tier, configurable at
-- creation. Defaults to 'autonomous' so every Worker is automatable unless the
-- creator downgrades it. The per-agent adapter translates this into each
-- agent's native permission mechanism at pod-create time.

ALTER TABLE pods ADD COLUMN automation_level VARCHAR(20) NOT NULL DEFAULT 'autonomous';
ALTER TABLE experts ADD COLUMN automation_level VARCHAR(20) NOT NULL DEFAULT 'autonomous';

ALTER TABLE pods ADD CONSTRAINT pods_automation_level_check
  CHECK (automation_level IN ('interactive', 'auto_edit', 'autonomous'));
ALTER TABLE experts ADD CONSTRAINT experts_automation_level_check
  CHECK (automation_level IN ('interactive', 'auto_edit', 'autonomous'));
