package controllers

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golangblog/config"
	"golangblog/libs"
	"golangblog/models"
	"golangblog/schemas"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

type bodylink struct {
	Name string
	URL  string
}

type emailreset struct {
	URL string
}

var refresh_secret = []byte(os.Getenv("REFRESH_SECRET"))

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
	token, errToken := config.CreateToken(users.ID)
	if errToken != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "There was an error authenticating."})
		return
	}

	setErr := config.CreateAuth(users.ID, token)
	if setErr != nil {
		c.JSON(http.StatusUnprocessableEntity, setErr.Error())
	}

	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}

	if !confirmation.Activated {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Account is not actived, check your email to verified",
		})
	} else {
		c.JSON(http.StatusOK, tokens)
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

func RevokeToken(c *gin.Context) {
	detail, err := config.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		return
	}

	deleted, delErr := config.DeleteAuth(detail.AccessUuid)
	if delErr != nil || deleted == 0 {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		return
	}
	c.JSON(http.StatusOK, "Access token revoked")
}

func RefreshToken(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	refreshToken := mapToken["refresh_token"]

	// verify token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return refresh_secret, nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// check token valid
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		jti, ok := claims["jti"].(string)
		if !ok {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["identity"]), 10, 64)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, "Error occured")
			return
		}
		deleted, delErr := config.DeleteAuth(jti)
		if delErr != nil || deleted == 0 {
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		// create new access & refresh token
		ts, createErr := config.CreateToken(userId)
		if createErr != nil {
			c.JSON(http.StatusForbidden, createErr.Error())
			return
		}

		// save token to redis
		saveErr := config.CreateAuth(userId, ts)
		if saveErr != nil {
			c.JSON(http.StatusForbidden, saveErr.Error())
			return
		}

		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}

		c.JSON(http.StatusCreated, tokens)
	} else {
		c.JSON(http.StatusUnauthorized, "refresh expired")
	}
}

func GetUsers(c *gin.Context) {
	users := models.GetUser()
	c.JSON(http.StatusOK, users)
}

func SendPasswordReset(c *gin.Context) {
	// get email
	var payload models.SendReset

	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		_ = c.AbortWithError(422, err).SetType(gin.ErrorTypeBind)
		return
	}

	// find email in database
	email := models.CheckEmailExists(payload.Email)
	if email == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user with that email address not found."})
		return
	}

	userId, _ := models.FindUserByEmail(payload.Email)
	// check user activated or not
	confirmation, _ := models.FindConfirmation(userId)
	if !confirmation.Activated {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Please activated you're user first"})
		return
	}

	// check email on password reset if it doesen't exist then add the email
	pass_reset := models.CheckEmailPasswordReset(payload.Email)
	if pass_reset == 0 {
		models.SavePasswordReset(payload.Email)
		// send email
		link := os.Getenv("APP_URL") + "/test"
		templateData := emailreset{
			URL: link,
		}

		runtime.GOMAXPROCS(1)
		go libs.SendEmailPasswordReset(payload.Email, templateData)
		c.JSON(http.StatusOK, gin.H{"message": "We have e-mailed your password reset link!"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "already saved"})
	}
}
