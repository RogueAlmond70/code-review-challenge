package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RogueAlmond70/code-review-challenge/internal/config"
	"github.com/RogueAlmond70/code-review-challenge/services"
	"github.com/RogueAlmond70/code-review-challenge/types"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	DB     services.DBClient
	Cfg    *config.Config
	logger *zap.Logger
}

func (s Server) GetSingleNote(ctx context.Context) gin.HandlerFunc {

}
func (s Server) GetNotes() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		userID := userId(c)
		if userID == "" {
			s.logger.Warn("missing user ID in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		includeArchived := c.DefaultQuery("includeArchived", "false") == "true"
		includeActive := c.DefaultQuery("includeActive", "true") == "true"

		// Validate filters
		if !includeArchived && !includeActive {
			s.logger.Warn("no notes included", zap.String("userID", userID))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "must include at least one of active or archived notes"})
			return
		}

		// Parse pagination parameters
		limit, offset, err := parsePagination(c)
		if err != nil {
			s.logger.Warn("invalid pagination params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid pagination parameters"})
			return
		}

		var archiveFilter *bool
		switch {
		case includeArchived && includeActive:
			archiveFilter = nil
		case includeArchived:
			archiveFilter = ptr(true)
		case includeActive:
			archiveFilter = ptr(false)
		}

		notes, totalCount, err := s.DB.GetNotes(ctx, userID, archiveFilter, limit, offset)
		if err != nil {
			s.logger.Error("failed to get notes", zap.String("userID", userID), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notes"})
			return
		}

		hasMore := offset+len(notes) < totalCount

		c.JSON(http.StatusOK, types.NotesResponse{
			Notes:      notes,
			Offset:     offset,
			Limit:      limit,
			TotalNotes: totalCount,
			HasMore:    hasMore,
		})
	}
}

// Helper to return a pointer to a bool
func ptr(b bool) *bool {
	return &b
}

// parsePagination parses "limit" and "offset" query parameters with sane defaults and validation.
func parsePagination(c *gin.Context) (limit int, offset int, err error) {
	const (
		defaultLimit = 50
		maxLimit     = 200
	)

	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > maxLimit {
		return 0, 0, fmt.Errorf("limit must be between 1 and %d", maxLimit)
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, fmt.Errorf("offset must be non-negative")
	}

	return limit, offset, nil
}

func (s Server) CreateNote(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newNote types.NoteDto

		if err := c.BindJSON(&newNote); err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		title := ""

		if newNote.Title != nil { // Validate (not a blank string or too long > 255 characters)
			title = *newNote.Title // Also input sanitisation to prevent injection / XSS before saving
		}

		content := ""

		if newNote.Content != nil { // Validate (not a blank string or too long >255 characters)
			content = *newNote.Content // Also input sanitisation to prevent injection / XSS before saving
		}

		// Use context with timeouts for database operations
		// Check for duplicates before creating a note
		note, err := services.CreateNote(db, userId(c), title, content)
		if err != nil {
			fmt.Println(err) // proper error handling. Wrap errors, use logging, also metrics
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, note) // Return a 200 http status code on creation. Increment / Capture metrics
	}
}
