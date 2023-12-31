package testutils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type githubStub struct {
	serverAddr string
}

func (t *githubStub) RoundTrip(r *http.Request) (*http.Response, error) {

	if strings.Contains(r.URL.String(), "127.0.0.1") {
		return http.DefaultTransport.RoundTrip(r)
	}

	if strings.Contains(r.URL.String(), "https://github.com/login/oauth/authorize") {

		var buf bytes.Buffer
		gotState := r.URL.Query().Get("state")
		v := url.Values{}
		v.Add("state", gotState)
		v.Add("code", "abcd1234")

		buf.WriteString(t.serverAddr + "/github/callback")
		buf.WriteByte('?')
		buf.WriteString(v.Encode())

		resp := http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Header: map[string][]string{
				"Location": {buf.String()},
			},
		}
		return &resp, nil
	}

	if strings.Contains(r.URL.String(), "https://github.com/login/oauth/access_token") {
		responseBody := "access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&scope=repo%2Cgist&token_type=bearer"
		respReader := io.NopCloser(strings.NewReader(responseBody))
		resp := http.Response{
			StatusCode:    http.StatusOK,
			Body:          respReader,
			ContentLength: int64(len(responseBody)),
			Header: map[string][]string{
				"Content-Type": {"application/x-www-form-urlencoded"},
			},
		}
		return &resp, nil
	}

	if r.URL.String() == "https://api.github.com/user" {
		responseBody := `
{
  "login": "test-user-1",
  "id": 1,		
  "type": "User",
  "site_admin": false,
  "name": "test user 1",
  "company": "GitHub",
  "email": "test-user-1@github.com"
}`
		respReader := io.NopCloser(strings.NewReader(responseBody))
		resp := http.Response{
			StatusCode:    http.StatusOK,
			Body:          respReader,
			ContentLength: int64(len(responseBody)),
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		}
		return &resp, nil
	}

	return nil, fmt.Errorf("github interceptor: unknown URL: %v", r.URL.String())
}

func HttpClientWithGithubStub(serverAddr string) *http.Client {
	return &http.Client{
		Transport: &githubStub{serverAddr: serverAddr},
	}
}
