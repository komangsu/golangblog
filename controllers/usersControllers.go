package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golangblog/libs"
	"golangblog/models"
	"net/http"
)

// Create user
func CreateUser(c *gin.Context) {
	var user models.User

	// trimspace
	user.Prepare()

	// validation
	if err := c.ShouldBindBodyWith(&user, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	// insert user
	_, userErr := user.SaveUser()

	if userErr != nil {
		formaterror := libs.Formaterror(userErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": formaterror,
		})
	}
	c.JSON(http.StatusCreated, "successfully created user.")
}

// get user
func GetUser(c *gin.Context) {
	username := c.Param("username")

	user, err := models.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"users": user,
	})
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
	} else if users.Email != payload.Email {
		c.JSON(http.StatusUnauthorized, "email not valid")
	}

	// create access token
	token, errToken := models.CreateToken(users.ID)
	if errToken != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "There was an error authenticating."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
