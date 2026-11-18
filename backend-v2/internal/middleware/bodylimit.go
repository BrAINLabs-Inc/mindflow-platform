package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSONBodyLimit is express.json({ limit: '1mb' })'s cap.
const JSONBodyLimit = 1 << 20

// BodyLimit rejects request bodies larger than limit with the same 413
// response app.ts's error handler returns for body-parser's
// entity.too.large. The body is read up front (bounded at limit+1 bytes)
// so the rejection doesn't depend on a client-supplied Content-Length.
// Not applied to the multipart upload route, which enforces its own 10MB
// cap via http.MaxBytesReader.
func BodyLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body too large."})
			return
		}
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, limit+1))
		_ = c.Request.Body.Close()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Malformed JSON in request body."})
			return
		}
		if int64(len(body)) > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body too large."})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
