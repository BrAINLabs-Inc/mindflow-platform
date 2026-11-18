-- ==============================================================================
-- Patch 2026-11-18: RLS hardening + query indexes
-- ==============================================================================
-- Brings an EXISTING database in line with database/project_db.sql (which
-- already contains these changes for fresh bootstraps). Idempotent -- safe
-- to run more than once. See FINDINGS_AND_DEVELOPMENT.md §3 (D1-D6).
--
-- No participant client (mobile/, web-app/) writes to Supabase directly --
-- every participant write goes through the backend's service-role /
-- BYPASSRLS connection, which RLS doesn't apply to. These changes only
-- remove the ability to bypass the backend with the public anon key.
-- Admin (web-admin) access is unchanged.
-- ==============================================================================

BEGIN;

-- ------------------------------------------------------------------------------
-- D1: participants must not be able to set/change their own research_id
-- (study arm). RLS can't restrict columns, so a trigger enforces it: a
-- change made under a participant JWT is rejected unless the caller is an
-- admin. auth.uid() is NULL for the service role, the backend-v2 Postgres
-- role and Supabase Auth's own signup trigger, so those are unaffected.
-- ------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.guard_profiles_research_id()
RETURNS TRIGGER AS $$
BEGIN
    IF auth.uid() IS NOT NULL
       AND (
            (TG_OP = 'INSERT' AND NEW.research_id IS NOT NULL)
         OR (TG_OP = 'UPDATE' AND NEW.research_id IS DISTINCT FROM OLD.research_id)
       )
       AND NOT EXISTS (SELECT 1 FROM public.admins WHERE id = auth.uid())
    THEN
        RAISE EXCEPTION 'research_id can only be assigned by a researcher'
            USING ERRCODE = '42501';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, auth;

DROP TRIGGER IF EXISTS guard_profiles_research_id ON profiles;
CREATE TRIGGER guard_profiles_research_id
    BEFORE INSERT OR UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION public.guard_profiles_research_id();

-- ------------------------------------------------------------------------------
-- D2/D3: participant policies on research-data tables become read-only.
-- ------------------------------------------------------------------------------
DROP POLICY IF EXISTS "Users can insert own daily sliders" ON daily_sliders;
DROP POLICY IF EXISTS "Users can update own daily sliders" ON daily_sliders;

DROP POLICY IF EXISTS "Users can insert own recordings" ON voice_recordings;
DROP POLICY IF EXISTS "Users can update own recordings" ON voice_recordings;

DROP POLICY IF EXISTS "Users manage own PSS10 responses" ON questionnaire_pss10_responses;
DROP POLICY IF EXISTS "Users view own PSS10 responses" ON questionnaire_pss10_responses;
CREATE POLICY "Users view own PSS10 responses"
    ON questionnaire_pss10_responses FOR SELECT USING (user_id = auth.uid());

DROP POLICY IF EXISTS "Users manage own FFMQ15 responses" ON questionnaire_ffmq15_responses;
DROP POLICY IF EXISTS "Users view own FFMQ15 responses" ON questionnaire_ffmq15_responses;
CREATE POLICY "Users view own FFMQ15 responses"
    ON questionnaire_ffmq15_responses FOR SELECT USING (user_id = auth.uid());

DROP POLICY IF EXISTS "Users manage own WEMWBS14 responses" ON questionnaire_wemwbs14_responses;
DROP POLICY IF EXISTS "Users view own WEMWBS14 responses" ON questionnaire_wemwbs14_responses;
CREATE POLICY "Users view own WEMWBS14 responses"
    ON questionnaire_wemwbs14_responses FOR SELECT USING (user_id = auth.uid());

DROP POLICY IF EXISTS "Users see own, Admins see all" ON about_me_profiles;
CREATE POLICY "Users see own, Admins see all"
    ON about_me_profiles
    FOR SELECT
    USING (
        id = auth.uid()
        OR
        EXISTS (SELECT 1 FROM admins WHERE id = auth.uid())
    );
DROP POLICY IF EXISTS "Admins manage all about me profiles" ON about_me_profiles;
CREATE POLICY "Admins manage all about me profiles"
    ON about_me_profiles FOR ALL
    USING (EXISTS (SELECT 1 FROM admins WHERE id = auth.uid()))
    WITH CHECK (EXISTS (SELECT 1 FROM admins WHERE id = auth.uid()));

-- ------------------------------------------------------------------------------
-- D5: composite indexes matching every status/streak/journey query's
-- `WHERE user_id = $1 AND created_at >= $2 ORDER BY created_at DESC LIMIT n`.
-- ------------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_daily_sliders_user_created
    ON daily_sliders (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_voice_recordings_user_created
    ON voice_recordings (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_pss10_user_created
    ON questionnaire_pss10_responses (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ffmq15_user_created
    ON questionnaire_ffmq15_responses (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wemwbs14_user_created
    ON questionnaire_wemwbs14_responses (user_id, created_at DESC);

COMMIT;
