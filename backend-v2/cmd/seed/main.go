// Command seed fills a Supabase project with realistic dev/test data. It is
// a port of backend/src/seed.ts and self-provisions its own participant and
// admin auth users, so it works in any environment.
//
// Like the TS original it goes through Supabase's HTTP APIs (Auth admin +
// PostgREST) with the service-role key, NOT through DATABASE_URL: the API's
// least-privilege Postgres role deliberately can't write admins or
// calendar_events. Never run it against production.
//
//	make seed     (or: go run ./cmd/seed)
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/timeutil"
)

const (
	participantEmail    = "seed.participant@mindflow.app"
	participantPassword = "SeedParticipant123!"
	adminEmail          = "seed.admin@mindflow.app"
	adminPassword       = "SeedAdmin123!"
)

type supabase struct {
	baseURL    string
	serviceKey string
	http       *http.Client
}

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("reading .env: %v", err)
	}
	url, key := os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if url == "" || key == "" {
		log.Fatal("FATAL: SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY are required for seeding.")
	}
	sb := &supabase{baseURL: strings.TrimSuffix(url, "/"), serviceKey: key, http: &http.Client{Timeout: 30 * time.Second}}
	if err := seed(context.Background(), sb); err != nil {
		log.Fatalf("Seed failed: %v", err)
	}
}

func seed(ctx context.Context, sb *supabase) error {
	fmt.Println("Starting database seed...")
	fmt.Println()

	fmt.Println("  -> Creating/finding participant auth user...")
	participantID, err := sb.findOrCreateAuthUser(ctx, participantEmail, participantPassword)
	if err != nil {
		return err
	}
	fmt.Printf("     participant id: %s\n", participantID)

	// Table-level failures are reported and seeding continues, like seed.ts
	// (e.g. re-running hits the one-entry-per-day constraint on daily_sliders).
	step := func(label string, n int, err error) {
		if err != nil {
			fmt.Printf("    %s error: %v\n", label, err)
		} else {
			fmt.Printf("    %s: %d entries\n", label, n)
		}
	}

	fmt.Println("  -> Seeding profiles...")
	step("profiles", 1, sb.upsert(ctx, "profiles", []map[string]any{{
		"id": participantID, "username": "seed_participant", "research_id": "MF-SEED-001.ex",
	}}))

	fmt.Println("  -> Seeding about_me_profiles...")
	step("about_me_profiles", 1, sb.upsert(ctx, "about_me_profiles", []map[string]any{{
		"id":                   participantID,
		"university_id":        "UNI-SEED-001",
		"education_level":      "Bachelor's Degree",
		"major_field_of_study": "Computer Science",
		"age":                  22,
		"living_situation":     "Living alone",
		"family_background":    "Supportive family environment",
		"cultural_background":  "Sri Lankan",
		"hobbies_interests":    "Meditation, yoga, hiking, reading",
		"personal_goals":       "Improve daily mindfulness practice and reduce stress",
		"why_mindflow":         "To track my mindfulness journey and contribute to research",
		"is_completed":         true,
		"updated_at":           time.Now().UTC().Format(time.RFC3339),
	}}))

	fmt.Println("  -> Creating/finding admin auth user...")
	adminID, err := sb.findOrCreateAuthUser(ctx, adminEmail, adminPassword)
	if err != nil {
		return err
	}
	step("admins", 1, sb.upsert(ctx, "admins", []map[string]any{{
		"id": adminID, "username": "seed_admin", "email": adminEmail,
	}}))

	tables := []struct {
		name string
		rows []map[string]any
	}{
		{"daily_sliders", dailySliders(participantID, 42)},
		{"voice_recordings", voiceRecordings(participantID, 8)},
		{"questionnaire_pss10_responses", questionnaire(participantID, 6, 10, 7, 3, 120, 600)},
		{"questionnaire_wemwbs14_responses", questionnaire(participantID, 3, 14, 14, 5, 180, 900)},
		{"questionnaire_ffmq15_responses", ffmq15(participantID, 6)},
		{"calendar_events", calendarEvents(20)},
	}
	for _, t := range tables {
		fmt.Printf("  -> Seeding %s...\n", t.name)
		step(t.name, len(t.rows), sb.insert(ctx, t.name, t.rows))
	}

	fmt.Println()
	fmt.Println("Seed complete!")
	fmt.Println()
	fmt.Printf("  Participant: %s / %s (%s)\n", participantEmail, participantPassword, participantID)
	fmt.Printf("  Admin:       %s / %s (%s)\n\n", adminEmail, adminPassword, adminID)
	return nil
}

// --- data generators (same shapes and ranges as seed.ts) --------------------

func randInt(lo, hi int) int { return lo + rand.IntN(hi-lo+1) }

func pick[T any](xs []T) T { return xs[rand.IntN(len(xs))] }

// daysAgo returns an RFC 3339 timestamp n days ago at a random daytime hour
// in Sri Lanka time. (seed.ts used the machine's local time; on a UTC
// machine two consecutive days could then fall on the same SLT
// entry_date and fail daily_sliders' one-per-day constraint.)
func daysAgo(n int) string {
	d := time.Now().In(timeutil.SLT).AddDate(0, 0, -n)
	d = time.Date(d.Year(), d.Month(), d.Day(), randInt(6, 22), randInt(0, 59), randInt(0, 59), 0, timeutil.SLT)
	return d.UTC().Format(time.RFC3339)
}

func dailySliders(userID string, count int) []map[string]any {
	feelings := []any{
		"Felt calm and focused today.",
		"Had a restless morning but improved after meditation.",
		"Grateful for a good night's sleep.",
		"Struggled with anxiety in the afternoon.",
		"Energized after a walk in nature.",
		"Felt overwhelmed with work deadlines.",
		"Peaceful evening with family.",
		"Managed stress well through breathing exercises.",
		nil,
	}
	rows := make([]map[string]any, count)
	for i := range rows {
		practiced := pick([]string{"yes", "no"})
		var duration, location any
		if practiced == "yes" {
			duration, location = randInt(5, 30), pick([]string{"At University", "Outside University"})
		}
		rows[i] = map[string]any{
			"user_id":              userID,
			"stress_level":         randInt(1, 5),
			"calm_before":          randInt(1, 5),
			"calm_after":           randInt(1, 5),
			"sleep_quality":        randInt(1, 5),
			"sleep_start_time":     fmt.Sprintf("%d:%02d", randInt(21, 23), randInt(0, 59)),
			"wake_up_time":         fmt.Sprintf("%d:%02d", randInt(5, 8), randInt(0, 59)),
			"feelings":             pick(feelings),
			"mindfulness_practice": practiced,
			"practice_duration":    duration,
			"practice_location":    location,
			"video_play_seconds":   randInt(0, 300),
			"created_at":           daysAgo(i),
		}
	}
	return rows
}

// voiceRecordings uses the API's own Sri Lanka ISO-week helper, so seeded
// rows line up with what GET /api/roadmap/weekly/status looks for.
func voiceRecordings(userID string, count int) []map[string]any {
	rows := make([]map[string]any, count)
	for i := range rows {
		at := time.Now().AddDate(0, 0, -7*i)
		year, week := timeutil.GetISOWeekNumber(at)
		key := fmt.Sprintf("WeeklyVoice/weekly-%d-W%02d-%s.wav", year, week, userID)
		rows[i] = map[string]any{
			"user_id":     userID,
			"week_number": week,
			"year":        year,
			"file_url":    "https://storage.example.com/" + key,
			"file_key":    key,
			"duration":    randInt(15, 45),
			"created_at":  at.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

// questionnaire generates PSS-10 / WEMWBS-14 style rows: `items` answers of
// 1-5, one row roughly every `spacingDays` (+0..jitter days).
func questionnaire(userID string, count, items, spacingDays, jitter, durLo, durHi int) []map[string]any {
	rows := make([]map[string]any, count)
	for i := range rows {
		row := map[string]any{
			"user_id":    userID,
			"created_at": daysAgo(i*spacingDays + randInt(0, jitter)),
			"duration":   randInt(durLo, durHi),
		}
		for q := 1; q <= items; q++ {
			row[fmt.Sprintf("q%d", q)] = randInt(1, 5)
		}
		rows[i] = row
	}
	return rows
}

// ffmq15 includes facet scores computed with the study's exact item mapping
// (same as internal/services/mindful.go).
func ffmq15(userID string, count int) []map[string]any {
	rev := func(v int) int { return 6 - v }
	rows := make([]map[string]any, count)
	for i := range rows {
		a := make([]int, 16) // 1-indexed
		row := map[string]any{"user_id": userID}
		for q := 1; q <= 15; q++ {
			a[q] = randInt(1, 5)
			row[fmt.Sprintf("q%d", q)] = a[q]
		}
		row["observing_score"] = a[1] + a[6] + a[11]
		row["describing_score"] = a[2] + a[7] + rev(a[12])
		row["awareness_score"] = rev(a[3]) + rev(a[8]) + rev(a[13])
		row["non_judging_score"] = rev(a[4]) + rev(a[9]) + rev(a[14])
		row["non_reactivity_score"] = a[5] + a[10] + a[15]
		row["duration"] = randInt(200, 800)
		row["created_at"] = daysAgo(i*7 + randInt(0, 3))
		rows[i] = row
	}
	return rows
}

func calendarEvents(count int) []map[string]any {
	titles := []string{
		"Daily Check-in", "Weekly Voice Recording", "Stress Assessment",
		"Wellbeing Check", "Mindfulness Assessment", "Meditation Session",
		"Yoga Practice", "Breathing Exercise", "Body Scan", "Gratitude Journal",
	}
	descs := []any{
		"Completed morning routine", "Afternoon mindfulness break",
		"Evening reflection session", "Quick breathing exercise",
		"Full meditation session", nil,
	}
	rows := make([]map[string]any, count)
	for i := range rows {
		rows[i] = map[string]any{
			"title":        pick(titles),
			"description":  pick(descs),
			"event_date":   time.Now().AddDate(0, 0, -i).Format("2006-01-02"),
			"is_completed": rand.Float64() > 0.2,
		}
	}
	return rows
}

// --- Supabase HTTP helpers ---------------------------------------------------

func (s *supabase) do(ctx context.Context, method, path string, body any, extraHeaders map[string]string) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

func (s *supabase) insert(ctx context.Context, table string, rows []map[string]any) error {
	_, err := s.do(ctx, http.MethodPost, "/rest/v1/"+table, rows, map[string]string{"Prefer": "return=minimal"})
	return err
}

func (s *supabase) upsert(ctx context.Context, table string, rows []map[string]any) error {
	_, err := s.do(ctx, http.MethodPost, "/rest/v1/"+table, rows,
		map[string]string{"Prefer": "resolution=merge-duplicates,return=minimal"})
	return err
}

// findOrCreateAuthUser creates a confirmed auth user, or finds the existing
// one by email if creation fails because it already exists.
func (s *supabase) findOrCreateAuthUser(ctx context.Context, email, password string) (string, error) {
	body, err := s.do(ctx, http.MethodPost, "/auth/v1/admin/users", map[string]any{
		"email": email, "password": password, "email_confirm": true,
	}, nil)
	if err == nil {
		var u struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(body, &u) == nil && u.ID != "" {
			return u.ID, nil
		}
	}

	const perPage = 200
	for page := 1; ; page++ {
		body, err := s.do(ctx, http.MethodGet, fmt.Sprintf("/auth/v1/admin/users?page=%d&per_page=%d", page, perPage), nil, nil)
		if err != nil {
			return "", err
		}
		var list struct {
			Users []struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"users"`
		}
		if err := json.Unmarshal(body, &list); err != nil {
			return "", err
		}
		for _, u := range list.Users {
			if strings.EqualFold(u.Email, email) {
				return u.ID, nil
			}
		}
		if len(list.Users) < perPage {
			return "", fmt.Errorf("could not find or create auth user for %s", email)
		}
	}
}
