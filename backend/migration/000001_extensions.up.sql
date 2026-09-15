-- Required extensions.
-- pgcrypto: gen_random_uuid() for tokens/opaque identifiers.
-- pg_trgm: trigram indexes for fuzzy ILIKE search (author/title autocomplete).
-- unaccent: accent-insensitive search normalization.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;
-- citext: case-insensitive text, used for email uniqueness.
CREATE EXTENSION IF NOT EXISTS citext;
