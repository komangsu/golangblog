package controllers

import (
	"github.com/gin-gonic/gin"
	"golangblog/database"
	"golangblog/models"
	"net/http"
)

func CreateArticle(c *gin.Context) {
	var article models.Article

	tokenAuth, err := database.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		return
	}

	userId, fetchErr := database.FetchAuth(tokenAuth)
	if fetchErr != nil {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		return
	}

	articleErr := models.CreateArticle(article.Title, article.Content, userId)
	if articleErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed created article"})
		return
	}
	c.JSON(http.StatusCreated, "Successfully created article")
}
