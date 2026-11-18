package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimit(t *testing.T) {
	r := gin.New()
	r.POST("/echo", BodyLimit(16), func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(b))
	})

	t.Run("under limit passes body through", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"a":1}`)))
		if w.Code != http.StatusOK || w.Body.String() != `{"a":1}` {
			t.Fatalf("got %d %q", w.Code, w.Body.String())
		}
	})

	t.Run("over limit is 413 with Express's message", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(strings.Repeat("x", 17))))
		if w.Code != http.StatusRequestEntityTooLarge || w.Body.String() != `{"error":"Request body too large."}` {
			t.Fatalf("got %d %q", w.Code, w.Body.String())
		}
	})

	t.Run("over limit without Content-Length is still 413", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/echo", io.NopCloser(strings.NewReader(strings.Repeat("x", 17))))
		req.ContentLength = -1
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("got %d", w.Code)
		}
	})
}
