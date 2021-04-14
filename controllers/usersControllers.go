package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golangblog/libs"
	"golangblog/models"
	"net/http"
)

type link struct {
	Name string
	URL  string
}

// Create user
func CreateUser(c *gin.Context) {
	var payload models.RegisterUser

	// validation
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	// check duplicate email
	email := models.CheckEmailExists(payload.Email)
	if email != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already taken."})
		return
	}

	// insert user
	userErr := models.SaveUser(payload)

	if userErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed creating account"})
		return
	}

	// send email after create account
	templateData := link{
		Name: "Potter",
		URL:  "https://i.giphy.com/media/pI2paNxecnUNW/giphy.webp",
	}

	libs.SendEmailVerification(payload.Email, templateData)

	c.JSON(http.StatusCreated, gin.H{"message": "Success, check your email to verification"})

}

// Login User
func LoginUser(c *gin.Context) {
	var payload models.LoginUser

	// validation
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	// check email
	email := models.CheckEmailExists(payload.Email)
	if email == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User not found."})
		return
	}

	users, errUser := models.VerifyLogin(payload.Email, payload.Password)
	if errUser != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid login credentials"})
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
