package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

/*
  From manning book: External APIs: GitHub sign in application

  This will use golang.org/x/oauth2, so need to go get it :-)
  Then I did go mod tidy and go mod vendor.

  Now I also need
  go get github.com/google/go-github/v47
  and the same other cmd's.

  This is the combo GitHub Web App that I downloaded from the live project.

*/

var (
	oauthConf       *oauth2.Config
	oauthHttpClient *http.Client
)

func initOAuthConfig(getEnvironValue func(string) string) {

	if len(getEnvironValue("CLIENT_ID")) == 0 || len(getEnvironValue("CLIENT_SECRET")) == 0 {
		log.Fatal("Must specify your app's CLIENT_ID and CLIENT_SECRET")
	}

	oauthConf = &oauth2.Config{
		ClientID:     getEnvironValue("CLIENT_ID"),
		ClientSecret: getEnvironValue("CLIENT_SECRET"),
		Scopes:       []string{"repo", "user"}, // see the project description for understanding why we need full scopes here
		Endpoint:     github.Endpoint,
	}
}

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/github/callback", githubCallbackHandler)
}

func main() {

	initOAuthConfig(os.Getenv)
	oauthHttpClient = &http.Client{}

	mux := http.NewServeMux()
	registerHandlers(mux)

	log.Println("Starting server..")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
