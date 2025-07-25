package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pushfar/code-review-challenge/endpoints"
	"github.com/pushfar/code-review-challenge/middleware"
	"github.com/pushfar/code-review-challenge/services"
)

func main() {
	router := gin.Default()

	router.Use(middleware.BasicAuth()) // Sending credentials with every request is silly. Use JWT or something.

	db, err := services.OpenDB()
	if err != nil {
		fmt.Println("Unable to open database")
		fmt.Println(err)
		return
	}

	router.GET("/notes", endpoints.GetNotes(db))
	router.POST("/note", endpoints.CreateNote(db))
	router.PATCH("/note/:id", endpoints.UpdateNote(db)) // This is incorrectly labelled as a PUT method in the README
	router.DELETE("/note/:id", endpoints.DeleteNote(db))

	router.Run("localhost:8080") // We should be using a context and running this in a goroutine
}
