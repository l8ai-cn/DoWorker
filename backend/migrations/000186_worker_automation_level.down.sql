ALTER TABLE experts DROP CONSTRAINT IF EXISTS experts_automation_level_check;
ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_automation_level_check;
ALTER TABLE experts DROP COLUMN IF EXISTS automation_level;
ALTER TABLE pods DROP COLUMN IF EXISTS automation_level;
