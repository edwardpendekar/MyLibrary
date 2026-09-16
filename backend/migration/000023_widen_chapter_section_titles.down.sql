-- Only safe to roll back if no title longer than 255 characters was
-- inserted while this migration was active — Postgres will reject the
-- ALTER otherwise, same as any other lossy down-migration.
ALTER TABLE sections ALTER COLUMN title_id TYPE VARCHAR(255);
ALTER TABLE sections ALTER COLUMN title_en TYPE VARCHAR(255);
ALTER TABLE chapters ALTER COLUMN title_id TYPE VARCHAR(255);
ALTER TABLE chapters ALTER COLUMN title_en TYPE VARCHAR(255);
