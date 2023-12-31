package main

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	gthb "golang.org/x/oauth2/github" // this is the correct import to resolve the symbol of "github"  But it now conflicts w/ go-github, so I have to create and use an alias.
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v57/github"
)

/*
  From manning book: External APIs: GitHub sign in application

  This will use golang.org/x/oauth2, so need to go get it :-)
  Then I did go mod tidy and go mod vendor.

  Now I also need
  go get github.com/google/go-github/v47
  and the same other cmd's.

  This is the combined GitHub Web App that I merged from the instructions, but I didn't write the add'l routines.
  The live project code is in combo GitHub Web App dir.
*/

func main() {

	initOAuthConfig()

	mux := http.NewServeMux()
	registerHandlers(mux)

	go func() { // I'm using a goroutine to run the server, because I have to spin the server up before the testutils stuff runs.
		log.Println("Starting server..")
		log.Fatal(http.ListenAndServe(":8080", mux))
	}()

	ctx := context.Background()
	httpClient := HttpClientWithGithubStub("")
	client := github.NewClient(httpClient)
	u, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GitHub login: %s\n", *u.Login)
}

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
		Endpoint:     gthb.Endpoint, // the problem here was I had the wrong import.  I only saw that from the solution code.
	}
}

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/github/callback", githubCallbackHandler)
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
