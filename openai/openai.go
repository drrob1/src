package openai

import (
	"context"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	"os"
)

/*
  From Linux Magazine 270 May 2023.  Creating a Go program to interact w/ ChatGPT.  GPT means Generative Pretrained Transformer.

  31 Aug 2024 -- Started typing this is from the magazine article.
   1 Sep 2024 -- Separated out the apikey because it seems that github won't allow the key to be committed.
                 I emailed mschilli@perlmeister.com for help.  Admittedly, he wrote this article ~1.5 yrs ago.  The openAI API clearly changed since then.
                 Mike Schilli
   3 Sep 2024 -- He answered me and provided a working routine.  He wrote that the API changed.  He also said I have to
     1) create a new project in openAI's settings.  Top left it says personal / default project.  Clicking on the arrows by default project lets me create a new project.
                    I called it "using API"
     2) create a new secret api key for this project.  Left column -> Your Profile -> center User API Keys -> named the key "using API" and generated it and copied it to a .bat file.
     3) in the project's settings (gear icon top right) go to limits and set the project's budget to whatever you prepaid
     4) use the updated openai.go (below), as the API has changed since the article was published.
*/

type OpenAI struct {
	Ctx context.Context
	Cli gpt3.Client
}

func NewAI() *OpenAI {
	return &OpenAI{}
}

func (ai *OpenAI) Init() {
	apikey, ok := os.LookupEnv("chatgptkey")
	if !ok {
		panic("/n/n apikey is required \n\n")
	}
	ai.Ctx = context.Background()
	ai.Cli = gpt3.NewClient(apikey)
}

func (ai *OpenAI) PrintResp(prompt string) {
	req := gpt3.ChatCompletionRequest{
		Model: gpt3.GPT3Dot5Turbo,
		Messages: []gpt3.ChatCompletionRequestMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1000,
		Temperature: gpt3.Float32Ptr(0),
	}
	resp, err := ai.Cli.ChatCompletion(ai.Ctx, req)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", resp.Choices[0].Message.Content)
}
