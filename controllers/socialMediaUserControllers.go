package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golangblog/libs"
	"golangblog/models"
	"net/http"
)

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
	c.JSON(http.StatusOK, "successfuly login")

	jsonErr := json.Unmarshal(user, &uGoogle)
	if jsonErr != nil {
		fmt.Println("error:", err)
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
	c.JSON(http.StatusOK, "successfuly login")

	jsonErr := json.Unmarshal(user, &uFacebook)
	if jsonErr != nil {
		fmt.Println("error:", err)
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

}
