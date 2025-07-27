package endpoints

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RogueAlmond70/code-review-challenge/internal/config"
	"github.com/RogueAlmond70/code-review-challenge/internal/datastore"
	"github.com/RogueAlmond70/code-review-challenge/services"
	"github.com/RogueAlmond70/code-review-challenge/types"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

type Server struct {
	DB     services.DBClient
	Cfg    *config.Config
	logger *zap.Logger
}

func (s Server) GetSingleNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		userID := userId(c)
		if userID == "" {
			s.logger.Warn("missing user ID in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		noteID := c.Param("noteId")
		if noteID == "" {
			s.logger.Warn("missing note ID in request URL")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "note ID must be provided"})
			return
		}

		note, err := s.DB.GetSingleNote(ctx, userID, noteID)
		if err != nil {
			if errors.Is(err, datastore.ErrNoteNoteFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "note not found"})
				return
			}

			s.logger.Error("failed to get single note", zap.String("userID", userID), zap.String("noteID", noteID), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve note"})
			return
		}

		c.JSON(http.StatusOK, note)
	}
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

func (s Server) CreateNote() gin.HandlerFunc {
	const maxTitleLen = 255
	const maxContentLen = 10000 // arbitrary sane max for content

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var newNote types.NoteDto
		if err := c.BindJSON(&newNote); err != nil {
			s.logger.Warn("invalid JSON body", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		userID := userId(c)
		if userID == "" {
			s.logger.Warn("missing user ID in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Validate and sanitize title
		title := ""
		if newNote.Title != nil {
			title = strings.TrimSpace(*newNote.Title)
			if title == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title cannot be empty"})
				return
			}
			if len(title) > maxTitleLen {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("title length cannot exceed %d characters", maxTitleLen)})
				return
			}
			title = sanitizeInput(title)
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title is required"})
			return
		}

		// Validate and sanitize content
		content := ""
		if newNote.Content != nil {
			content = strings.TrimSpace(*newNote.Content)
			if len(content) > maxContentLen {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("content length cannot exceed %d characters", maxContentLen)})
				return
			}
			content = sanitizeInput(content)
		}

		// Check for duplicate title for this user
		existingNotes, _, err := s.DB.GetNotes(ctx, userID, nil, 1, 0) // get first note, no filter
		if err != nil {
			s.logger.Error("failed to check for duplicate note", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		for _, note := range existingNotes {
			if note.Title == title {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "note with this title already exists"})
				return
			}
		}

		// Create note in DB
		createdNote, err := s.DB.CreateNote(ctx, userID, title, content)
		if err != nil {
			s.logger.Error("failed to create note", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
			return
		}

		// Return created note with 201 status
		c.JSON(http.StatusCreated, createdNote)
	}
}

// sanitizeInput uses a strict HTML sanitizer to remove potentially dangerous input.
// It prevents XSS by stripping out scripts, unsafe tags, and attributes.
func sanitizeInput(input string) string {
	// Use bluemonday's StrictPolicy, which allows only plain text
	policy := bluemonday.StrictPolicy()
	return policy.Sanitize(input)
}

func (s Server) DeleteNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		userID := userId(c)
		if userID == "" {
			s.logger.Warn("missing user ID in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		noteID := c.Param("noteId")
		if noteID == "" {
			s.logger.Warn("missing note ID in request URL")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "note ID must be provided"})
			return
		}

		err := s.DB.DeleteNote(ctx, userID, noteID)
		if err != nil {
			if errors.Is(err, datastore.ErrNoteNoteFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "note not found"})
				return
			}

			s.logger.Error("failed to delete note", zap.String("userID", userID), zap.String("noteID", noteID), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
			return
		}

		c.Status(http.StatusNoContent) // 204 No Content
	}
}
