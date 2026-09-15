INSERT INTO roles (name, description) VALUES
    ('admin',  'Full access: manage books, users, imports, and settings.'),
    ('editor', 'Can manage books, chapters, verses, and run imports.'),
    ('user',   'Registered reader: favorites, bookmarks, notes.'),
    ('guest',  'Unauthenticated read-only access.')
ON CONFLICT (name) DO NOTHING;

INSERT INTO languages (code, name, native_name) VALUES
    ('en', 'English', 'English'),
    ('id', 'Indonesian', 'Bahasa Indonesia')
ON CONFLICT (code) DO NOTHING;

INSERT INTO categories (slug, name_en, name_id, description) VALUES
    ('religion', 'Religion', 'Agama', 'Religious and scripture texts.'),
    ('fiction', 'Fiction', 'Fiksi', 'Fictional works and novels.'),
    ('history', 'History', 'Sejarah', 'Historical texts and references.'),
    ('philosophy', 'Philosophy', 'Filsafat', 'Philosophical works.')
ON CONFLICT (slug) DO NOTHING;

-- Bootstrap admin account. Password is "ChangeMe123!" (bcrypt hash below) — MUST be rotated
-- immediately in any non-local environment; see docs/deployment.md.
INSERT INTO users (role_id, name, email, password_hash, is_active, email_verified_at)
SELECT r.id, 'Administrator', 'admin@bookreader.local',
       '$2b$12$uFJUctLVptd5wE6v0fujY.3P1VjKayZKYGYLByLia9TTgYliNErBG',
       true, now()
FROM roles r
WHERE r.name = 'admin'
ON CONFLICT (email) DO NOTHING;
