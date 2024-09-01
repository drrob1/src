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
		panic("apikey is required")
	}
	ai.Ctx = context.Background()
	ai.Cli = gpt3.NewClient(apikey)
	// ai.Cli = gpt3.NewClient(apikey, gpt3.WithBaseURL("https://api.openai.com/v1/chat/completions"))  Invalid request error
}

func (ai *OpenAI) PrintResp(prompt string) {
	req := gpt3.CompletionRequest{
		Prompt:      []string{prompt},
		MaxTokens:   gpt3.IntPtr(1000),
		Temperature: gpt3.Float32Ptr(0), // allowed values are 0, 1 and 2.  But article says that setting this to 2 makes it hallucinate.
		Stop:        []string{"."},
		Echo:        true,
	}
	ondata := func(resp *gpt3.CompletionResponse) {
		fmt.Print(resp.Choices[0].Text)
	}
	//turboInstruct := gpt3.GPT3Dot5Turbo + "-instruct"  doesn't work, I'm getting an exceeding free tier message
	//turboInstruct := gpt3.GPT3Dot5Turbo  doesn't work, I'm getting an exceeding free tier message
	//turboInstruct := "gpt-4o-mini" // doesn't work, giving an exceeding free tier message
	//turboInstruct := "gpt-4o" // doesn't work, giving an exceeding free tier message
	// err := ai.Cli.CompletionStreamWithEngine(ai.Ctx, "", req, ondata)  Not allow to post on v1/engings/completions

	turboInstruct := gpt3.GPT3Dot5Turbo
	err := ai.Cli.CompletionStreamWithEngine(ai.Ctx, turboInstruct, req, ondata)
	if err != nil {
		panic(err)
	}
	fmt.Println("")
}

func main() {
	fmt.Printf(" not done yet.\n")
}
