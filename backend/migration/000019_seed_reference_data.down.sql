DELETE FROM users WHERE email = 'admin@bookreader.local';
DELETE FROM categories WHERE slug IN ('religion', 'fiction', 'history', 'philosophy');
DELETE FROM languages WHERE code IN ('en', 'id');
DELETE FROM roles WHERE name IN ('admin', 'editor', 'user', 'guest');
