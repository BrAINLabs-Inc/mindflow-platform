package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/r2"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/services"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/testutil"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/timeutil"
)

// weeklyKeyForTest mirrors weekly.go's unexported weeklyVoiceKey format
// exactly, for constructing a key SubmitWeeklyEntry will accept.
func weeklyKeyForTest(year, week int, userID string) string {
	return fmt.Sprintf("WeeklyVoice/weekly-%d-W%02d-%s.wav", year, week, userID)
}

// testR2 builds an R2 client that is never contacted -- SubmitWeeklyEntry
// only needs its PublicURL derivation.
func testR2(t *testing.T) *r2.Client {
	t.Helper()
	c, err := r2.New(context.Background(), "acct", "key", "secret", "bucket", "https://cdn.example.com")
	if err != nil {
		t.Fatalf("r2.New: %v", err)
	}
	return c
}

func TestSubmitWeeklyEntry_FileURLMustMatchKey(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("seeding user: %v", err)
	}
	svc := services.NewWeeklyService(q, testR2(t))
	year, week := timeutil.GetISOWeekNumber(time.Now())
	fileKey := weeklyKeyForTest(year, week, userID.String())

	for _, badURL := range []string{
		"javascript:alert(document.cookie)",
		"https://phish.example.net/" + fileKey,
		"https://cdn.example.com/WeeklyVoice/other.wav",
	} {
		if _, err := svc.SubmitWeeklyEntry(ctx, pgUserID, badURL, fileKey, nil); !errors.Is(err, services.ErrWeeklyFileKeyMismatch) {
			t.Errorf("file_url %q: error = %v, want ErrWeeklyFileKeyMismatch", badURL, err)
		}
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM voice_recordings WHERE user_id = $1`, userID).Scan(&n); err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	if n != 0 {
		t.Errorf("rows = %d, want 0 (mismatched URLs must not be stored)", n)
	}
}

func TestSubmitWeeklyEntry_KeyMismatchRejected(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("seeding user: %v", err)
	}

	svc := services.NewWeeklyService(q, testR2(t))

	_, err := svc.SubmitWeeklyEntry(ctx, pgUserID, "https://cdn.example.com/somewhere-else.wav", "WeeklyVoice/weekly-2000-W01-someone-elses-id.wav", nil)
	if !errors.Is(err, services.ErrWeeklyFileKeyMismatch) {
		t.Fatalf("error = %v, want ErrWeeklyFileKeyMismatch", err)
	}
}

func TestSubmitWeeklyEntry_UpsertsOnMatchingKey(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("seeding user: %v", err)
	}

	svc := services.NewWeeklyService(q, testR2(t))
	year, week := timeutil.GetISOWeekNumber(time.Now())
	fileKey := weeklyKeyForTest(year, week, userID.String())
	duration := int32(42)

	row, err := svc.SubmitWeeklyEntry(ctx, pgUserID, "https://cdn.example.com/"+fileKey, fileKey, &duration)
	if err != nil {
		t.Fatalf("SubmitWeeklyEntry: %v", err)
	}
	if row.FileKey != fileKey {
		t.Errorf("FileKey = %q, want %q", row.FileKey, fileKey)
	}

	// Re-submitting the same week (matching key) should update, not error
	// or duplicate -- see voice_recordings_user_week_year_unique.
	newDuration := int32(99)
	row2, err := svc.SubmitWeeklyEntry(ctx, pgUserID, "https://cdn.example.com/"+fileKey, fileKey, &newDuration)
	if err != nil {
		t.Fatalf("re-submit: %v", err)
	}
	if row2.Duration.Int32 != newDuration {
		t.Errorf("Duration after resubmit = %d, want %d (upsert should update)", row2.Duration.Int32, newDuration)
	}

	var rowCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM voice_recordings WHERE user_id = $1`, userID).Scan(&rowCount); err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	if rowCount != 1 {
		t.Errorf("row count = %d, want 1 (upsert, not duplicate)", rowCount)
	}
}

func TestGetWeeklyVideo_GroupMatchAndFallback(t *testing.T) {
	pool, q := testutil.NewTestDB(t)
	ctx := context.Background()

	_, week := timeutil.GetISOWeekNumber(time.Now())

	exUser := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, exUser); err != nil {
		t.Fatalf("seeding ex user: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE profiles SET research_id = $2 WHERE id = $1`, exUser, "MF-TEST.ex"); err != nil {
		t.Fatalf("setting ex research_id: %v", err)
	}

	cgUser := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO auth.users (id) VALUES ($1)`, cgUser); err != nil {
		t.Fatalf("seeding cg user: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE profiles SET research_id = $2 WHERE id = $1`, cgUser, "MF-TEST.cg"); err != nil {
		t.Fatalf("setting cg research_id: %v", err)
	}

	// A group-agnostic (global) video and an ex-only video for this week.
	if _, err := pool.Exec(ctx, `INSERT INTO weekly_recordings (week_no, title, youtube_id, target_group, published_at) VALUES ($1, 'Global Video', 'yt-global', NULL, NOW())`, week); err != nil {
		t.Fatalf("seeding global video: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO weekly_recordings (week_no, title, youtube_id, target_group, published_at) VALUES ($1, 'EX Video', 'yt-ex', 'ex', NOW())`, week); err != nil {
		t.Fatalf("seeding ex video: %v", err)
	}

	svc := services.NewWeeklyService(q, nil)

	exVideo, err := svc.GetWeeklyVideo(ctx, pgtype.UUID{Bytes: exUser, Valid: true})
	if err != nil {
		t.Fatalf("GetWeeklyVideo(ex): %v", err)
	}
	if exVideo == nil || exVideo.Title != "EX Video" {
		t.Errorf("ex user's video = %v, want the group-matched EX Video", exVideo)
	}

	cgVideo, err := svc.GetWeeklyVideo(ctx, pgtype.UUID{Bytes: cgUser, Valid: true})
	if err != nil {
		t.Fatalf("GetWeeklyVideo(cg): %v", err)
	}
	if cgVideo == nil || cgVideo.Title != "Global Video" {
		t.Errorf("cg user's video = %v, want fallback to the group-agnostic Global Video", cgVideo)
	}
}
