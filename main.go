package main

/*
- Define a database interface and refactor the functions to be methods of that interface.
- Cleanup the endpoint code
- Replace the basic authentication with a JWT implementation - add expiry of tokens
- Create a metrics file where I define Prometheus metrics for success and errors
- Increment these metrics appropriately in the endpoints (for example)
- Refine the containerisation by moving credentials to environment variables -done
- Add a prometheus instance and create a docker compose
- Update the readme file
- Other general cleanup
- Check the Bruno tests work, make them work if needed
- Implement pagination for the get endpoints - done!
- If there's time, implement duplicate databases with a slave master config and load balancing
*/

import (
	"fmt"

	"github.com/RogueAlmond70/code-review-challenge/endpoints"
	"github.com/RogueAlmond70/code-review-challenge/internal/middleware"
	"github.com/RogueAlmond70/code-review-challenge/services"
	"github.com/gin-gonic/gin"
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

	// Create instance of userStore to pass in to register and login routes
	userStore := services.NewUserStore(db)
	router.POST("/register", endpoints.Register(userStore))
	router.POST("/login", endpoints.Login(userStore))

	router.GET("/notes", endpoints.GetNotes(db))
	router.POST("/note", endpoints.CreateNote(db))
	router.PATCH("/note/:id", endpoints.UpdateNote(db)) // This is incorrectly labelled as a PUT method in the README
	router.DELETE("/note/:id", endpoints.DeleteNote(db))

	router.Run("localhost:8080") // We should be using a context and running this in a goroutine
}
