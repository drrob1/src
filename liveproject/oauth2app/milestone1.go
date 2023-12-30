package main

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github" // this is the correct import to resolve the symbol of "github"
	"log"
	"net/http"
	"os"
	//"github.com/google/go-github/v47/github"
	//"github.com/google/go-github/v57/github" // current as of Dec-30-2023, but is not correct for this project.
)

/*
  From manning book: External API's: GitHub sign in application

  This will use golang.org/x/oauth2, so need to go get it :-)
  Then I did go mod tidy and go mod vendor.

  Now I also need
  go get github.com/google/go-github/v47
  and the same other cmd's.
*/

var oauthConf *oauth2.Config

type userData struct {
	Login       string // github login
	accessToken string // GitHub provides this.
}

func initOAuthConfig() {
	if len(os.Getenv("CLIENT_ID")) == 0 || len(os.Getenv("CLIENT_SECRET")) == 0 {
		log.Fatal("Must specify your app's CLIENT_ID and CLIENT_SECRET")
	}

	oauthConf = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"repo", "user"},
		Endpoint:     github.Endpoint, // the problem here was I had the wrong import.  I only saw that from the solution code.
	}
}

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/github/callback", githubCallbackHandler)
}

func main() {

	initOAuthConfig()

	mux := http.NewServeMux()
	registerHandlers(mux)
	log.Println("Starting server..")
	log.Fatal(http.ListenAndServe(":8080", mux))

}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	s, err := getSession(req)
	if err != nil {
		stateToken, err := getRandomString()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		githubLoginUrl := oauthConf.AuthCodeURL(stateToken)
		setCookie(w, oauthStateCookie, stateToken, 600)
		http.Redirect(w, req, githubLoginUrl, http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprintf(w, "Successfully authorized to access GitHub on your behalf: %#v", sessionsStore[s.ID].Login)
}
