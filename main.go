package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gobeam/custom-validator"
	"golangblog/controllers"
	"golangblog/database"
	"golangblog/middleware"
	"net/http"
)

func main() {
	router := gin.Default()

	// connect to database
	database.InitDB()
	database.InitRedis()

	router.Use(validator.Errors())
	{
		router.POST("/users", controllers.CreateUser)
		router.POST("/login", controllers.LoginUser)
	}
	router.GET("/confirm-email", controllers.VerifyAccount)
	router.POST("/refresh", controllers.RefreshToken)
	router.DELETE("/access-token-revoke", middleware.TokenAuthMiddleware(), controllers.RevokeToken)

	router.POST("/article/create", middleware.TokenAuthMiddleware(), controllers.CreateArticle)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "golangblog",
		})
	})

	router.Run()
}
