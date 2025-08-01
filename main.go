package main

import (
	"fmt"

	"github.com/RogueAlmond70/code-review-challenge/endpoints"
	"github.com/RogueAlmond70/code-review-challenge/internal/middleware"
	"github.com/RogueAlmond70/code-review-challenge/services"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.Use(middleware.BasicAuth())

	db, err := services.OpenDB()
	if err != nil {
		fmt.Println("Unable to open database")
		fmt.Println(err)
		return
	}

	userStore := services.NewUserStore(db)
	router.POST("/register", endpoints.Register(userStore))
	router.POST("/login", endpoints.Login(userStore))

	router.GET("/notes", endpoints.GetNotes(db))
	router.POST("/note", endpoints.CreateNote(db))
	router.PATCH("/note/:id", endpoints.UpdateNote(db)) // This is incorrectly labelled as a PUT method in the README
	router.DELETE("/note/:id", endpoints.DeleteNote(db))

	router.Run("localhost:8080")
}
