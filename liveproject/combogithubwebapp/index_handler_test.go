package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"src/liveproject/combogithubwebapp/testutils"
	"testing"

	"golang.org/x/net/publicsuffix"
)

func TestIndexHandler(t *testing.T) {

	mux := http.NewServeMux()
	registerHandlers(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	initOAuthConfig(testutils.GetenvStub)
	oauthHttpClient = testutils.HttpClientWithGithubStub(ts.URL)

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		t.Fatal(err)
	}

	testHTTPClient := testutils.HttpClientWithGithubStub(ts.URL)
	testHTTPClient.Jar = jar
	testHTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		t.Logf("via: %#+v\n", via)
		t.Logf("redirect to: %s\n", req.URL)
		t.Log()
		return nil
	}

	resp, err := testHTTPClient.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedData := "Successfully authorized to access GitHub on your behalf: test-user-1"

	if string(respBytes) != expectedData {
		t.Fatalf("Expected: %s, Got: %s", expectedData, string(respBytes))
	}
}
