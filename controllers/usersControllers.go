package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golangblog/models"
	"net/http"
)

// Create user
func CreateUser(c *gin.Context) {
	var user models.User
	var payload models.RegisterUser

	// trimspace
	user.Prepare()

	// validation
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	// check duplicate email
	_, errMail := models.GetEmail(payload.Email)
	if errMail != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": errMail})
		return
	}

	// insert user
	_, userErr := models.SaveUser(payload.Username, payload.Email, payload.Password)

	if userErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": userErr,
		})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "There was an error authenticating."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
