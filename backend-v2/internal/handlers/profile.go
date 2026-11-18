package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"unicode/utf16"

	"github.com/gin-gonic/gin"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/db"
	appmiddleware "github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/middleware"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/services"
)

// Port of backend/src/controllers/profileController.ts.
type ProfileHandler struct {
	svc *services.ProfileService
}

func NewProfileHandler(svc *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	user, ok := appmiddleware.AuthUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := db.ParseUUID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		log.Printf("getProfile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var email *string
	if user.Email != "" {
		email = &user.Email
	}
	c.JSON(http.StatusOK, gin.H{
		"username":    profile.Username,
		"research_id": profile.ResearchID,
		"email":       email,
	})
}

func (h *ProfileHandler) GetAboutMe(c *gin.Context) {
	user, ok := appmiddleware.AuthUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := db.ParseUUID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	aboutMe, err := h.svc.GetAboutMe(c.Request.Context(), userID)
	if err != nil {
		log.Printf("getAboutMe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, aboutMe)
}

// aboutMeTextFields is profileController.ts's aboutMeSchema string fields
// with their .max() limits, in schema order.
var aboutMeTextFields = []struct {
	name string
	max  int
}{
	{"university_id", 200},
	{"education_level", 200},
	{"faculty", 200},
	{"major_field_of_study", 200},
	{"living_situation", 200},
	{"family_background", 2000},
	{"cultural_background", 200},
	{"hobbies_interests", 2000},
	{"personal_goals", 2000},
	{"why_mindflow", 2000},
}

// aboutMeRequiredWhenCompleted mirrors the superRefine `required` list:
// is_completed=true is rejected unless each of these is present and
// non-empty in the same request (personal_goals is deliberately optional).
var aboutMeRequiredWhenCompleted = []string{
	"university_id", "education_level", "faculty", "major_field_of_study",
	"living_situation", "family_background", "cultural_background",
	"hobbies_interests", "why_mindflow",
}

// UpdateAboutMe ports updateAboutMe: server-side mirror of AboutMeScreen's
// required-field validation (so a client can't mark onboarding complete
// while skipping fields), then a partial upsert of the allowlisted fields.
func (h *ProfileHandler) UpdateAboutMe(c *gin.Context) {
	user, ok := appmiddleware.AuthUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	update, issues, err := parseAboutMeBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission", "details": []validationIssue{{Path: "body", Message: "Malformed request body"}}})
		return
	}
	if len(issues) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission", "details": issues})
		return
	}
	userID, err := db.ParseUUID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	result, err := h.svc.UpdateAboutMe(c.Request.Context(), userID, update)
	if err != nil {
		log.Printf("updateAboutMe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// parseAboutMeBody decodes and validates the request body with the same
// rules as aboutMeSchema. A non-nil error means the body isn't a JSON
// object at all; validation failures come back as issues instead.
// Unknown keys are ignored, like Zod's default object stripping.
func parseAboutMeBody(c *gin.Context) (services.AboutMeUpdate, []validationIssue, error) {
	var raw map[string]json.RawMessage
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&raw); err != nil || raw == nil {
		return services.AboutMeUpdate{}, nil, fmt.Errorf("body must be a JSON object")
	}

	var issues []validationIssue
	text := make(map[string]services.Field[string], len(aboutMeTextFields))
	for _, f := range aboutMeTextFields {
		v, present := raw[f.name]
		if !present {
			continue
		}
		field := services.Field[string]{Set: true}
		if !isJSONNull(v) {
			var str string
			if err := json.Unmarshal(v, &str); err != nil {
				issues = append(issues, validationIssue{Path: f.name, Message: f.name + " must be a string"})
				continue
			}
			str = strings.TrimSpace(str)
			// Zod's .max() counts UTF-16 code units (JS string length).
			if len(utf16.Encode([]rune(str))) > f.max {
				issues = append(issues, validationIssue{Path: f.name, Message: fmt.Sprintf("%s must be at most %d characters", f.name, f.max)})
				continue
			}
			field.Value = &str
		}
		text[f.name] = field
	}

	var age services.Field[int32]
	if v, present := raw["age"]; present {
		age.Set = true
		if !isJSONNull(v) {
			var n float64
			if err := json.Unmarshal(v, &n); err != nil || n != math.Trunc(n) || n <= 0 || n > 150 {
				issues = append(issues, validationIssue{Path: "age", Message: "age must be a positive integer no greater than 150"})
			} else {
				a := int32(n)
				age.Value = &a
			}
		}
	}

	var completed services.Field[bool]
	if v, present := raw["is_completed"]; present {
		// z.boolean().optional() -- unlike the other fields, null is invalid.
		var b bool
		if isJSONNull(v) || json.Unmarshal(v, &b) != nil {
			issues = append(issues, validationIssue{Path: "is_completed", Message: "is_completed must be a boolean"})
		} else {
			completed = services.Field[bool]{Set: true, Value: &b}
		}
	}

	// superRefine only runs once the base schema parsed cleanly.
	if len(issues) == 0 && completed.Value != nil && *completed.Value {
		for _, name := range aboutMeRequiredWhenCompleted {
			if f := text[name]; f.Value == nil || *f.Value == "" {
				issues = append(issues, validationIssue{Path: name, Message: name + " is required to complete onboarding"})
			}
		}
		if age.Value == nil {
			issues = append(issues, validationIssue{Path: "age", Message: "A valid age is required to complete onboarding"})
		}
	}

	return services.AboutMeUpdate{
		UniversityID:       text["university_id"],
		EducationLevel:     text["education_level"],
		Faculty:            text["faculty"],
		MajorFieldOfStudy:  text["major_field_of_study"],
		Age:                age,
		LivingSituation:    text["living_situation"],
		FamilyBackground:   text["family_background"],
		CulturalBackground: text["cultural_background"],
		HobbiesInterests:   text["hobbies_interests"],
		PersonalGoals:      text["personal_goals"],
		WhyMindflow:        text["why_mindflow"],
		IsCompleted:        completed,
	}, issues, nil
}

func isJSONNull(v json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(v), []byte("null"))
}
