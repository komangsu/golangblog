package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golangblog/models"
	"net/http"
)

// Create user
func CreateUser(c *gin.Context) {
	var payload models.RegisterUser

	// validation
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	// check duplicate email
	u, _ := models.GetEmail(payload.Email)
	if u.Email != "" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Email is already taken!"})
		return
	}

	// insert user
	userErr := models.SaveUser(payload)

	if userErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, "successfully created user.")
}

// Login User
func LoginUser(c *gin.Context) {
	var payload models.LoginUser

	// validation
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	users, errUser := models.SignIn(payload.Email, payload.Password)
	if errUser != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUser.Error()})
		return
	}

	// create access token
	token, errToken := models.CreateToken(users.ID)
	if errToken != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "There was an error authenticating."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
