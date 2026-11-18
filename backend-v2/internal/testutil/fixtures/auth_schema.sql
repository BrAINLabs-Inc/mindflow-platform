-- Minimal stand-in for the parts of Supabase's managed `auth` schema that
-- database/project_db.sql references (auth.users for FKs, auth.uid() in
-- RLS policies, the `authenticated` role for GRANTs). Supabase provisions
-- all of this itself in a real project; a vanilla postgres:17 container
-- used for integration tests needs it stubbed out first so project_db.sql
-- applies unmodified -- this file must run BEFORE project_db.sql.
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY
);

-- Mirrors Supabase's auth.uid(): the `sub` claim of the current
-- PostgREST request, NULL when there is none. backend-v2's own connections
-- never set it (it does its own application-level `WHERE user_id = $1`
-- filtering -- see plan §3), so it is NULL for them, exactly as for the
-- service role in production. RLS/trigger tests set it per transaction
-- with set_config('request.jwt.claim.sub', ...) to act as a participant.
CREATE OR REPLACE FUNCTION auth.uid() RETURNS UUID AS $$
    SELECT NULLIF(current_setting('request.jwt.claim.sub', true), '')::UUID;
$$ LANGUAGE sql STABLE;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
        CREATE ROLE authenticated NOLOGIN;
    END IF;
END
$$;

GRANT USAGE ON SCHEMA auth TO authenticated;
