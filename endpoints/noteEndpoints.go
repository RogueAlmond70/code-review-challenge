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

func userId(c *gin.Context) string {
	return c.GetString("userId")
}

func GetNotes(db *sql.DB) gin.HandlerFunc { // This should really be a method of a database interface.
	return func(c *gin.Context) {
		userId := userId(c)

		// This below is a bit much. c.DefaultQuery("includeArchived", "false") == "true" is more idiomatic
		includeArchived := strings.Compare(c.DefaultQuery("includeArchived", "false"), "true") == 0
		includeActive := strings.Compare(c.DefaultQuery("includeActive", "true"), "true") == 0

		var notes []types.Note
		var err error

		// There is no pagination - we're just returning all results at once. Bad idea, also doesn't scale
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

		c.JSON(http.StatusOK, notes) // Metrics
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

		if newNote.Title != nil {
			title = *newNote.Title
		}

		content := ""

		if newNote.Content != nil {
			content = *newNote.Content
		}

		note, err := services.CreateNote(db, userId(c), title, content)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, note)
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
