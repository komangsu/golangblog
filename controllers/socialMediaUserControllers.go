package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golangblog/config"
	"golangblog/libs"
	"golangblog/models"
	"net/http"
)

const domain = "localhost"

func HandleMain(c *gin.Context) {
	var htmlIndex = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
</head>
<body>
<p>
	<a href="/login/google">Google</a>
</p>
<p>
	<a href="/login/facebook">Facebook</a>
</p>
</body>
</html>
`
	fmt.Fprintf(c.Writer, htmlIndex)
}
func GoogleLogin(c *gin.Context) {
	url := libs.GoogleOauthConfig.AuthCodeURL(libs.OauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleAuthorized(c *gin.Context) {
	var uGoogle models.UserGoogle

	user, err := libs.UserGoogleInfo(c.Query("state"), c.Query("code"))
	if err != nil {
		fmt.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	jsonErr := json.Unmarshal(user, &uGoogle)
	if jsonErr != nil {
		fmt.Println("error:", err)
	}

	// check duplicate email
	email := models.CheckEmailExists(uGoogle.Email)
	if email != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already taken."})
		return
	}

	// save it to database
	u, userErr := models.SaveGoogleUser(uGoogle)
	if userErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed save user"})
		return
	}

	// save to confirmation
	confErr := models.SaveConfirmation(u.ID)
	if confErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "cannot insert confirmation id"})
		return
	}

	verifErr := models.VerifyAccountModel(u.Email)
	if verifErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to verifying your account"})
		return
	}

	// create token
	token, errToken := config.CreateToken(u.ID)
	if errToken != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "There was an error authenticating"})
		return
	}

	// save to redis
	setErr := config.CreateAuth(u.ID, token)
	if setErr != nil {
		c.JSON(http.StatusUnprocessableEntity, setErr.Error())
	}

	// set cookie
	c.SetCookie("access_token", token.AccessToken, 0, "/login/google/authorized", domain, false, true)
	c.SetCookie("refresh_token", token.RefreshToken, 0, "/login/google/authorized", domain, false, true)
}

func FacebookLogin(c *gin.Context) {
	url := libs.FacebookOauth.AuthCodeURL(libs.OauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func FacebookAuthorized(c *gin.Context) {
	type dataPicture struct {
		Url string `json:"url"`
	}
	type data struct {
		Picture dataPicture `json:"data"`
	}
	type userFacebook struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar data   `json:"picture"`
	}

	var uFacebook userFacebook
	user, err := libs.UserFacebookInfo(c.Query("state"), c.Query("code"))
	if err != nil {
		fmt.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return

	}

	jsonErr := json.Unmarshal(user, &uFacebook)
	if jsonErr != nil {
		fmt.Println("error:", err)
	}

	// check duplicate email
	email := models.CheckEmailExists(uFacebook.Email)
	if email != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already taken."})
		return
	}

	// save to database
	u, userErr := models.SaveFacebookUser(uFacebook.Name, uFacebook.Email)
	if userErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed save user"})
		return
	}

	// save to confirmation
	confErr := models.SaveConfirmation(u.ID)
	if confErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "cannot insert confirmation id"})
		return
	}

	verifErr := models.VerifyAccountModel(u.Email)
	if verifErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to verifying your account"})
		return
	}

	// create token
	token, errToken := config.CreateToken(u.ID)
	if errToken != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "There was an error authenticating"})
		return
	}

	// save to redis
	setErr := config.CreateAuth(u.ID, token)
	if setErr != nil {
		c.JSON(http.StatusUnprocessableEntity, setErr.Error())
	}

	// set cookies
	c.SetCookie("access_token", token.AccessToken, 0, "/login/facebook/authorized", domain, false, true)
	c.SetCookie("refresh_token", token.RefreshToken, 0, "/login/facebook/authorized", domain, false, true)

}
