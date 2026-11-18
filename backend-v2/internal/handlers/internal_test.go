package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStoredContentType(t *testing.T) {
	cases := []struct{ mime, name, want string }{
		{"audio/wav", "rec.wav", "audio/wav"},
		{"audio/webm", "rec.webm", "audio/webm"},
		// Accepted via the extension fallback -- the declared type must not
		// be stored, or the object is served from the public bucket as HTML.
		{"text/html", "x.wav", "audio/wav"},
		{"text/html", "x.m4a", "audio/mp4"},
		{"image/svg+xml", "x.mp3", "audio/mpeg"},
		{"application/octet-stream", "blob", "application/octet-stream"},
	}
	for _, tc := range cases {
		if got := storedContentType(tc.mime, tc.name); got != tc.want {
			t.Errorf("storedContentType(%q, %q) = %q, want %q", tc.mime, tc.name, got, tc.want)
		}
	}
}

func parseAboutMe(t *testing.T, body string) ([]validationIssue, error) {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	_, issues, err := parseAboutMeBody(c)
	return issues, err
}

func issuePaths(issues []validationIssue) string {
	paths := make([]string, len(issues))
	for i, is := range issues {
		paths[i] = is.Path
	}
	return strings.Join(paths, ",")
}

func TestParseAboutMeBody(t *testing.T) {
	complete := `{"university_id":"IT1","education_level":"BSc","faculty":"Computing","major_field_of_study":"SE",
		"age":21,"living_situation":"Hostel","family_background":"x","cultural_background":"y",
		"hobbies_interests":"z","why_mindflow":"w","is_completed":true}`

	cases := []struct {
		name, body, wantPaths string
		wantErr               bool
	}{
		{"complete onboarding", complete, "", false},
		{"partial draft without is_completed", `{"faculty":"Computing"}`, "", false},
		{"explicit nulls are allowed", `{"faculty":null,"age":null}`, "", false},
		{"unknown keys are stripped, not rejected", `{"research_id":"X.ex"}`, "", false},
		{"is_completed requires fields", `{"is_completed":true,"faculty":"   "}`,
			"university_id,education_level,faculty,major_field_of_study,living_situation,family_background,cultural_background,hobbies_interests,why_mindflow,age", false},
		{"is_completed false needs nothing", `{"is_completed":false}`, "", false},
		{"is_completed null is invalid", `{"is_completed":null}`, "is_completed", false},
		{"non-integer age", `{"age":21.5}`, "age", false},
		{"age too high", `{"age":151}`, "age", false},
		{"age zero", `{"age":0}`, "age", false},
		{"wrong type", `{"faculty":5}`, "faculty", false},
		{"over max length", `{"faculty":"` + strings.Repeat("a", 201) + `"}`, "faculty", false},
		{"long field has 2000 limit", `{"why_mindflow":"` + strings.Repeat("a", 2000) + `"}`, "", false},
		{"not an object", `[1,2]`, "", true},
		{"malformed", `{`, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issues, err := parseAboutMe(t, tc.body)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got := issuePaths(issues); got != tc.wantPaths {
				t.Errorf("issue paths = %q, want %q", got, tc.wantPaths)
			}
		})
	}
}
