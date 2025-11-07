package genai

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/genai"
)

type GenaiChat struct {
	client 	*genai.Client
	ctx     context.Context
}

func NewGenaiChat() (*GenaiChat, error){
	ctx := context.Background()
	genai, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: "AIzaSyBeCd6zrt2XgFI9VhDxDjaz9xFOGOhwDE0"})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &GenaiChat{
		client: genai,
		ctx: ctx,
}, nil
}

func (o GenaiChat) GenaiChatting(text string) string {
	result, err := o.client.Models.GenerateContent(o.ctx, "gemini-2.5-flash-lite", genai.Text(text), nil)
	if err != nil {
		log.Fatal(err)
	}

	return result.Text()
}
