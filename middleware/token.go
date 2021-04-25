package middleware

import (
	"github.com/gin-gonic/gin"
	"golangblog/database"
	"net/http"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := database.TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}
	}
}
