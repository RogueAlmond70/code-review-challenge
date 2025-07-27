package middleware

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type user struct {
	id       string
	name     string
	password string
}

var users = []user{
	{id: "1", name: "user1", password: "1234"},
	{id: "2", name: "user2", password: "2345"},
}

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, pass, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userIndex, found := sort.Find(len(users), func(i int) int {
			return strings.Compare(username, users[i].name)
		})

		if !found || strings.Compare(users[userIndex].password, pass) != 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userId", users[userIndex].id)

		c.Next()
	}
}

// TODO: review
