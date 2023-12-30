package main

import (
	"context"
	"fmt"
	"log"

	"src/liveproject/githubstub/testutils"

	"github.com/google/go-github/v57/github"
)

func main() {
	ctx := context.Background()
	httpClient := testutils.HttpClientWithGithubStub("")
	client := github.NewClient(httpClient)
	u, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GitHub login: %s\n", *u.Login)
}
