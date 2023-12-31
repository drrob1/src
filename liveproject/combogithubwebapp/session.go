package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

type userData struct {
	Login       string
	accessToken string
}

var sessionsStore = make(map[string]userData)

type sessionData struct {
	ID string
}

const (
	oauthStateCookie    = "OAuthState"
	sessionCookie       = "Session"
	sessionCookieMaxAge = 24 * 3600 // 24 hours
)

// Read 32 bytes from the system's random number generator
// return the base64 encoding
func getRandomString() (string, error) {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func createSession(ctx context.Context, token string) (*sessionData, error) {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)

	u, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	sessionId, err := getRandomString()
	if err != nil {
		return nil, err
	}
	sessionsStore[sessionId] = userData{
		Login:       *u.Login,
		accessToken: token,
	}
	return &sessionData{
		ID: sessionId,
	}, nil
}

func validSessionID(sessionID string) bool {
	_, ok := sessionsStore[sessionID]
	return ok
}

func getSession(r *http.Request) (*sessionData, error) {

	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil, err
	}

	if !validSessionID(c.Value) {
		return nil, fmt.Errorf("Invalid session ID")
	}
	return &sessionData{ID: c.Value}, nil
}

func validCallback(r *http.Request) bool {

	gotState := r.URL.Query().Get("state")
	c, err := r.Cookie(oauthStateCookie)
	if err != nil {
		return false
	}
	if c.Value != gotState {
		return false
	}

	return true
}

func setCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  value,
		Path:   "/",
		MaxAge: maxAge,
	})
}
