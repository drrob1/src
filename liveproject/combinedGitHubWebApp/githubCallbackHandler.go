package main

import (
	"log"
	"net/http"
)

func githubCallbackHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if !validCallback(req) {
		log.Println("State tokens don't match. Ignoring callback.")
		http.Error(w, "Invalid callback request", http.StatusBadRequest)
		return
	}
	code := req.URL.Query().Get("code")
	t, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		log.Println("oAuth exchange error: ", err)
		http.Error(w, "Error logging in.", http.StatusInternalServerError)
		return
	}

	s, err := createSession(ctx, t.AccessToken)
	if err != nil {
		log.Println("Error creating session: ", err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	setCookie(w, sessionCookie, s.ID, sessionCookieMaxAge)
	setCookie(w, oauthStateCookie, "", -1)
	http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
	return
}
