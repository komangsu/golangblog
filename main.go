package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gobeam/custom-validator"
	"golangblog/controllers"
	"golangblog/database"
	"net/http"
)

func main() {
	router := gin.Default()

	// connect to database
	database.InitDB()

	router.Use(validator.Errors())
	{
		router.POST("/users", controllers.CreateUser)
		router.POST("/login", controllers.LoginUser)

	}
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "anjeng",
		})
	})

	router.Run()
}
