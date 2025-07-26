package endpoints

import (
	"context"
	"net/http"
	"strings"

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

func (s Server) GetNotes(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := userId(c)

		// This below is a bit much. c.DefaultQuery("includeArchived", "false") == "true" is more idiomatic
		includeArchived := strings.Compare(c.DefaultQuery("includeArchived", "false"), "true") == 0
		includeActive := strings.Compare(c.DefaultQuery("includeActive", "true"), "true") == 0

		var notes []types.Note
		var err error

		// There is no pagination - we're just returning all results at once. Bad idea (memory), also doesn't scale
		// Database queries really need to have context.Timeout so they don't potentially hang indefinitely.

		// Our GetNotes implementation makes use of an ArchiveFilter, which is a pointer to a boolean.
		t := true
		f := false

		if includeActive && includeArchived {
			notes, err = s.DB.GetNotes(ctx, userId, nil)
		} else if includeArchived {
			notes, err = s.DB.GetNotes(ctx, userId, &t)
		} else if includeActive {
			notes, err = s.DB.GetNotes(ctx, userId, &f)
		} else {
			s.logger.Warn("Nothing was included",
				zap.String("userId", userId))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err != nil {
			s.logger.Error("unable to get notes", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.JSON(http.StatusOK, notes) // Return a 200 http code if successful. Increment / Capture Metrics
	}
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
