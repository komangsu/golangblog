package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gobeam/custom-validator"
	"golangblog/config"
	"golangblog/controllers"
	"golangblog/middleware"
	"net/http"
)

func main() {
	router := gin.Default()

	// connect to database
	config.InitDB()
	config.InitRedis()

	router.GET("/", controllers.HandleMain)
	router.Use(validator.Errors())
	{
		router.POST("/register", controllers.CreateUser)
		router.POST("/login", controllers.LoginUser)
	}
	router.GET("/users", controllers.GetUsers)
	router.GET("/confirm-email", controllers.VerifyAccount)
	router.POST("/refresh", controllers.RefreshToken)
	router.DELETE("/access-token-revoke", middleware.TokenAuthMiddleware(), controllers.RevokeToken)
	router.POST("/send-password/reset", controllers.SendPasswordReset)

	router.GET("/login/google", controllers.GoogleLogin)
	router.GET("/login/google/authorized", controllers.GoogleAuthorized)
	router.GET("/login/facebook", controllers.FacebookLogin)
	router.GET("/login/facebook/authorized", controllers.FacebookAuthorized)

	router.POST("/article/create", middleware.TokenAuthMiddleware(), controllers.CreateArticle)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "golangblog",
		})
	})

	router.Run()
}
