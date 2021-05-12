package middleware

import (
	"github.com/gin-gonic/gin"
	"golangblog/config"
	"net/http"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := config.TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}
	}
}
