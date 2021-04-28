package libs

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	GoogleOauthConfig *oauth2.Config

	googleUrl    = os.Getenv("APP_URL") + "/login/google/authorized"
	googleId     = os.Getenv("GOOGLE_ID")
	googleSecret = os.Getenv("GOOGLE_SECRET")

	OauthStateString = "random"

	FacebookOauth  *oauth2.Config
	facebookId     = os.Getenv("FACEBOOK_ID")
	facebookSecret = os.Getenv("FACEBOOK_SECRET")
	facebookUrl    = os.Getenv("APP_URL") + "/login/facebook/authorized"
)

func init() {
	GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  googleUrl,
		ClientID:     googleId,
		ClientSecret: googleSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}

	FacebookOauth = &oauth2.Config{
		RedirectURL:  facebookUrl,
		ClientID:     facebookId,
		ClientSecret: facebookSecret,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}
}

func UserGoogleInfo(state string, code string) ([]byte, error) {
	if state != OauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := GoogleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	return contents, nil
}
