package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pushfar/code-review-challenge/services"
	"github.com/pushfar/code-review-challenge/types"
)

// Needs error handling, and validation. Maybe also return a bool to say if the user is present.
func userId(c *gin.Context) string {
	return c.GetString("userId")
}

func GetNotes(db *sql.DB) gin.HandlerFunc { // This should really be a method of a database interface. Also it's lacking input validation.
	return func(c *gin.Context) {
		userId := userId(c)

		// This below is a bit much. c.DefaultQuery("includeArchived", "false") == "true" is more idiomatic
		includeArchived := strings.Compare(c.DefaultQuery("includeArchived", "false"), "true") == 0
		includeActive := strings.Compare(c.DefaultQuery("includeActive", "true"), "true") == 0

		var notes []types.Note
		var err error

		// There is no pagination - we're just returning all results at once. Bad idea (memory), also doesn't scale
		// Database queries really need to have context.Timeout so they don't potentially hang indefinitely.
		if includeActive && includeArchived {
			notes, err = services.AllNotes(db, userId)
		} else if includeArchived {
			notes, err = services.ArchivedNotes(db, userId)
		} else if includeActive {
			notes, err = services.UnarchivedNotes(db, userId)
		} else {
			fmt.Println("Nothing was included")      // We should have logging instead of print statements
			c.AbortWithStatus(http.StatusBadRequest) // We also really would benefit from some metrics for failure and success.
			return
		}

		if err != nil {
			fmt.Println(err) // Errors should be wrapped for traceability, and metrics should be tracked and incremented.
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.JSON(http.StatusOK, notes) // Return a 200 http code if successful. Increment / Capture Metrics
	}
}

func CreateNote(db *sql.DB) gin.HandlerFunc { // This does not seem to work...
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

func UpdateNote(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if len(id) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		var newNote types.NoteDto

		if err := c.BindJSON(&newNote); err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		note, err := services.UpdateNote(db, userId(c), id, newNote)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, note)
	}
}

func DeleteNote(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if len(id) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		err := services.DeleteNote(db, userId(c), id)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
