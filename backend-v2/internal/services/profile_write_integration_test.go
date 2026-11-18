package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/services"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/testutil"
)

func str(s string) *string { return &s }

// TestUpdateAboutMe_PartialUpsert covers profileService.ts's updateAboutMe
// semantics: only fields present in the request are written, explicit
// nulls clear, absent fields keep their stored values.
func TestUpdateAboutMe_PartialUpsert(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := services.NewProfileService(q)

	userID := uuid.New()
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	// The handle_new_user_about_me trigger creates the row, as on signup.
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("seeding user: %v", err)
	}

	age := int32(22)
	got, err := svc.UpdateAboutMe(ctx, pgUserID, services.AboutMeUpdate{
		Faculty:       services.Field[string]{Set: true, Value: str("Computing")},
		PersonalGoals: services.Field[string]{Set: true, Value: str("Sleep better")},
		Age:           services.Field[int32]{Set: true, Value: &age},
	})
	if err != nil {
		t.Fatalf("first update: %v", err)
	}
	if got.ID != userID.String() || got.Faculty == nil || *got.Faculty != "Computing" || got.Age == nil || *got.Age != 22 || got.IsCompleted {
		t.Fatalf("after first update: %+v", got)
	}

	// Second call touches only personal_goals (explicit null) and
	// is_completed -- faculty and age must survive untouched.
	done := true
	got, err = svc.UpdateAboutMe(ctx, pgUserID, services.AboutMeUpdate{
		PersonalGoals: services.Field[string]{Set: true, Value: nil},
		IsCompleted:   services.Field[bool]{Set: true, Value: &done},
	})
	if err != nil {
		t.Fatalf("second update: %v", err)
	}
	if got.Faculty == nil || *got.Faculty != "Computing" {
		t.Errorf("faculty = %v, want preserved \"Computing\"", got.Faculty)
	}
	if got.Age == nil || *got.Age != 22 {
		t.Errorf("age = %v, want preserved 22", got.Age)
	}
	if got.PersonalGoals != nil {
		t.Errorf("personal_goals = %q, want cleared by explicit null", *got.PersonalGoals)
	}
	if !got.IsCompleted {
		t.Error("is_completed = false, want true")
	}

	// GET reflects the same row.
	read, err := svc.GetAboutMe(ctx, pgUserID)
	if err != nil || !read.IsCompleted || read.Faculty == nil || *read.Faculty != "Computing" {
		t.Fatalf("GetAboutMe after updates = %+v, %v", read, err)
	}
}

// TestUpdateAboutMe_InsertsWhenNoRow covers the upsert's insert path (row
// missing, e.g. trigger never ran): absent fields take column defaults.
func TestUpdateAboutMe_InsertsWhenNoRow(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := services.NewProfileService(q)

	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("seeding user: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM about_me_profiles WHERE id = $1`, userID); err != nil {
		t.Fatalf("removing trigger row: %v", err)
	}

	got, err := svc.UpdateAboutMe(ctx, pgtype.UUID{Bytes: userID, Valid: true}, services.AboutMeUpdate{
		Faculty: services.Field[string]{Set: true, Value: str("Engineering")},
	})
	if err != nil {
		t.Fatalf("UpdateAboutMe: %v", err)
	}
	if got.Faculty == nil || *got.Faculty != "Engineering" || got.IsCompleted || got.Age != nil {
		t.Errorf("inserted row = %+v", got)
	}
}

// TestSubmitDailyEntry_NormalisesFalsyToNull mirrors dailyService.ts's
// `value || null`: "" and 0 are stored as NULL, not verbatim.
func TestSubmitDailyEntry_NormalisesFalsyToNull(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := services.NewDailyService(q)

	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("seeding user: %v", err)
	}
	zero := int32(0)
	row, err := svc.SubmitDailyEntry(ctx, pgtype.UUID{Bytes: userID, Valid: true}, services.DailyEntryInput{
		StressLevel: 3, CalmBefore: 2, CalmAfter: 4, SleepQuality: 3,
		Feelings: str(""), SleepStartTime: str(""), WakeUpTime: str("06:30"),
		MindfulnessPractice: str("no"), PracticeDuration: &zero,
	})
	if err != nil {
		t.Fatalf("SubmitDailyEntry: %v", err)
	}
	if row.Feelings.Valid || row.SleepStartTime.Valid || row.PracticeDuration.Valid {
		t.Errorf("falsy values stored verbatim: feelings=%v sleep_start=%v duration=%v", row.Feelings, row.SleepStartTime, row.PracticeDuration)
	}
	if !row.WakeUpTime.Valid || row.WakeUpTime.String != "06:30" || row.MindfulnessPractice.String != "no" {
		t.Errorf("non-empty values altered: wake=%v practice=%v", row.WakeUpTime, row.MindfulnessPractice)
	}
}
