-- name: GetProfile :one
SELECT username, research_id FROM profiles WHERE id = $1;

-- name: GetResearchID :one
SELECT research_id FROM profiles WHERE id = $1;

-- name: GetAboutMe :one
SELECT * FROM about_me_profiles WHERE id = $1;

-- name: UpsertAboutMe :one
-- POST /api/profile/about-me (profileService.ts updateAboutMe). Partial-upsert
-- semantics, matching PostgREST's upsert of an allowlisted payload: only
-- fields present in the request body (set_<field> = true) are written --
-- an explicit null clears the column, an absent field keeps its stored
-- value on update and gets the column default on insert (NULL for the
-- text/int fields, FALSE for is_completed). updated_at is always bumped,
-- as the TS source sets it explicitly on every call.
INSERT INTO about_me_profiles (
    id, university_id, education_level, faculty, major_field_of_study, age, living_situation, family_background, cultural_background, hobbies_interests, personal_goals, why_mindflow, is_completed, updated_at
) VALUES (
    sqlc.arg(id),
    CASE WHEN sqlc.arg(set_university_id)::bool THEN sqlc.narg(university_id)::text END,
    CASE WHEN sqlc.arg(set_education_level)::bool THEN sqlc.narg(education_level)::text END,
    CASE WHEN sqlc.arg(set_faculty)::bool THEN sqlc.narg(faculty)::text END,
    CASE WHEN sqlc.arg(set_major_field_of_study)::bool THEN sqlc.narg(major_field_of_study)::text END,
    CASE WHEN sqlc.arg(set_age)::bool THEN sqlc.narg(age)::int4 END,
    CASE WHEN sqlc.arg(set_living_situation)::bool THEN sqlc.narg(living_situation)::text END,
    CASE WHEN sqlc.arg(set_family_background)::bool THEN sqlc.narg(family_background)::text END,
    CASE WHEN sqlc.arg(set_cultural_background)::bool THEN sqlc.narg(cultural_background)::text END,
    CASE WHEN sqlc.arg(set_hobbies_interests)::bool THEN sqlc.narg(hobbies_interests)::text END,
    CASE WHEN sqlc.arg(set_personal_goals)::bool THEN sqlc.narg(personal_goals)::text END,
    CASE WHEN sqlc.arg(set_why_mindflow)::bool THEN sqlc.narg(why_mindflow)::text END,
    CASE WHEN sqlc.arg(set_is_completed)::bool THEN sqlc.arg(is_completed)::bool ELSE FALSE END,
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    university_id = CASE WHEN sqlc.arg(set_university_id)::bool THEN EXCLUDED.university_id ELSE about_me_profiles.university_id END,
    education_level = CASE WHEN sqlc.arg(set_education_level)::bool THEN EXCLUDED.education_level ELSE about_me_profiles.education_level END,
    faculty = CASE WHEN sqlc.arg(set_faculty)::bool THEN EXCLUDED.faculty ELSE about_me_profiles.faculty END,
    major_field_of_study = CASE WHEN sqlc.arg(set_major_field_of_study)::bool THEN EXCLUDED.major_field_of_study ELSE about_me_profiles.major_field_of_study END,
    age = CASE WHEN sqlc.arg(set_age)::bool THEN EXCLUDED.age ELSE about_me_profiles.age END,
    living_situation = CASE WHEN sqlc.arg(set_living_situation)::bool THEN EXCLUDED.living_situation ELSE about_me_profiles.living_situation END,
    family_background = CASE WHEN sqlc.arg(set_family_background)::bool THEN EXCLUDED.family_background ELSE about_me_profiles.family_background END,
    cultural_background = CASE WHEN sqlc.arg(set_cultural_background)::bool THEN EXCLUDED.cultural_background ELSE about_me_profiles.cultural_background END,
    hobbies_interests = CASE WHEN sqlc.arg(set_hobbies_interests)::bool THEN EXCLUDED.hobbies_interests ELSE about_me_profiles.hobbies_interests END,
    personal_goals = CASE WHEN sqlc.arg(set_personal_goals)::bool THEN EXCLUDED.personal_goals ELSE about_me_profiles.personal_goals END,
    why_mindflow = CASE WHEN sqlc.arg(set_why_mindflow)::bool THEN EXCLUDED.why_mindflow ELSE about_me_profiles.why_mindflow END,
    is_completed = CASE WHEN sqlc.arg(set_is_completed)::bool THEN EXCLUDED.is_completed ELSE about_me_profiles.is_completed END,
    updated_at = NOW()
RETURNING *;

-- name: UpsertProfileUsername :exec
-- Signup's initial username assignment. Only touches username -- never
-- research_id (that's researcher-assigned only, never set at signup --
-- see CLAUDE.md's research-group note). A collision with another user's
-- username surfaces as a unique_violation (SQLSTATE 23505) on the
-- profiles_username_key constraint, which authController.ts's caller
-- catches to retry once with an id-suffixed username.
INSERT INTO profiles (id, username) VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username;
