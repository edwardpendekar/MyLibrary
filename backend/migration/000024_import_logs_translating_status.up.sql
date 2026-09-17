ALTER TABLE import_logs DROP CONSTRAINT chk_import_logs_status;
ALTER TABLE import_logs ADD CONSTRAINT chk_import_logs_status CHECK (
    status IN ('pending', 'validating', 'ready', 'translating', 'importing', 'completed', 'failed', 'rolled_back')
);
