package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/timeutil"
)

// fakeSupabase records PostgREST writes and emulates the Auth admin API:
// the participant already exists (create fails, found via list), the
// admin is created fresh.
func fakeSupabase(t *testing.T) (*httptest.Server, map[string][]map[string]any, *sync.Mutex) {
	t.Helper()
	var mu sync.Mutex
	writes := map[string][]map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("apikey") != "service-key" || r.Header.Get("Authorization") != "Bearer service-key" {
			t.Errorf("%s %s: missing service-role headers", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/auth/v1/admin/users":
			var in map[string]any
			_ = json.Unmarshal(body, &in)
			if in["email"] == participantEmail {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"msg":"A user with this email address has already been registered"}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":"admin-id"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/auth/v1/admin/users":
			_, _ = w.Write([]byte(`{"users":[{"id":"other","email":"x@y.z"},{"id":"participant-id","email":"Seed.Participant@mindflow.app"}]}`))
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/rest/v1/"):
			var rows []map[string]any
			if err := json.Unmarshal(body, &rows); err != nil {
				t.Errorf("%s: body is not a JSON array: %v", r.URL.Path, err)
			}
			mu.Lock()
			writes[strings.TrimPrefix(r.URL.Path, "/rest/v1/")] = rows
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, writes, &mu
}

func TestSeed(t *testing.T) {
	srv, writes, _ := fakeSupabase(t)
	sb := &supabase{baseURL: srv.URL, serviceKey: "service-key", http: &http.Client{Timeout: 5 * time.Second}}
	if err := seed(context.Background(), sb); err != nil {
		t.Fatalf("seed: %v", err)
	}

	want := map[string]int{
		"profiles": 1, "about_me_profiles": 1, "admins": 1,
		"daily_sliders": 42, "voice_recordings": 8,
		"questionnaire_pss10_responses": 6, "questionnaire_wemwbs14_responses": 3,
		"questionnaire_ffmq15_responses": 6, "calendar_events": 20,
	}
	for table, n := range want {
		if got := len(writes[table]); got != n {
			t.Errorf("%s: %d rows, want %d", table, got, n)
		}
	}
	if id := writes["profiles"][0]["id"]; id != "participant-id" {
		t.Errorf("existing participant not resolved via list: id = %v", id)
	}
	if id := writes["admins"][0]["id"]; id != "admin-id" {
		t.Errorf("admin id = %v", id)
	}

	// Daily rows: one per Sri Lanka calendar day (the table's unique constraint) and
	// mindfulness fields only when practiced.
	days := map[string]bool{}
	for _, r := range writes["daily_sliders"] {
		at, err := time.Parse(time.RFC3339, r["created_at"].(string))
		if err != nil {
			t.Fatal(err)
		}
		day := at.In(timeutil.SLT).Format("2006-01-02") // daily_sliders.entry_date
		if days[day] {
			t.Errorf("two daily rows on %s", day)
		}
		days[day] = true
		if (r["mindfulness_practice"] == "yes") != (r["practice_duration"] != nil) {
			t.Errorf("practice fields inconsistent: %v", r)
		}
	}

	// FFMQ facet scores follow the study mapping.
	for _, r := range writes["questionnaire_ffmq15_responses"] {
		q := func(n int) float64 { return r["q"+itoa(n)].(float64) }
		if r["describing_score"].(float64) != q(2)+q(7)+(6-q(12)) {
			t.Errorf("describing_score wrong: %v", r)
		}
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
