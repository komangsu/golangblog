package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golangblog/libs"
	"golangblog/models"
	"golangblog/schemas"
	"net/http"
	"os"
	"runtime"
)

type bodylink struct {
	Name string
	URL  string
}

// Create user
func CreateUser(c *gin.Context) {
	var payload schemas.RegisterUser

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
	u, userErr := models.SaveUser(payload)

	if userErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed creating account"})
		return
	}

	confErr := models.SaveConfirmation(u.ID)
	if confErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "cannot insert confirmation id"})
		return
	}

	// send email after create account
	token, _ := models.CreateAuthToken(u.Email)
	link := os.Getenv("APP_URL") + "/confirm-email?token=" + token
	templateData := bodylink{
		Name: u.Username,
		URL:  link,
	}

	runtime.GOMAXPROCS(1)
	go libs.SendEmailVerification(payload.Email, templateData)

	c.JSON(http.StatusCreated, gin.H{"message": "Success, check your email to verification"})

}

// Login User
func LoginUser(c *gin.Context) {
	var payload schemas.LoginUser

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
	confirmation, _ := models.FindConfirmation(users.ID)

	if errUser != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid login credentials"})
		return
	}

	// create access token
	token, errToken := models.CreateAuthToken(users.Email)
	if errToken != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "There was an error authenticating."})
		return
	}

	if !confirmation.Activated {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Account is not actived, check your email to verified",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func VerifyAccount(c *gin.Context) {
	verifyToken, _ := c.GetQuery("token")

	userId, _ := models.DecodeAuthToken(verifyToken)

	err := models.VerifyAccountModel(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to verifying your account,try again"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account verified, log in"})
}
