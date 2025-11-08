package genai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-deepseek/deepseek"
	"github.com/go-deepseek/deepseek/request"
)

type GenaiChat struct {
	client deepseek.Client
	messageList map[int64][]*request.Message
}

func NewGenaiChat(apikey string) (*GenaiChat, error){
	deepseek, err := deepseek.NewClient(apikey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &GenaiChat{
		client: deepseek,
		messageList: make(map[int64][]*request.Message),
}, nil
}

func (o *GenaiChat) GenaiChatting(text string, id int64) string {
	currentTime := time.Now()
	currentDate := currentTime.Format("2006-01-02")
	currentDateTime := currentTime.Format("2006-01-02 15:04:05")
	o.messageList[id] = append(o.messageList[id], &request.Message{Role: "system", Content: fmt.Sprintf("Ты полезный ассистент. Текущая дата: %s. Текущее время: %s Сегодня Отвечай на вопросы с учетом актуальной даты и времени.", currentDate, currentDateTime)})
	o.messageList[id] = append(o.messageList[id], &request.Message{Role: "user", Content: text})
	chatReq := &request.ChatCompletionsRequest{
		Model: deepseek.DEEPSEEK_CHAT_MODEL,
		Stream: false,
		Messages: o.messageList[id],
		}
	
	
	chatResp, err := o.client.CallChatCompletionsChat(context.Background(), chatReq)
	if err != nil {
		log.Printf("Ошибка %s", err)
	}
	o.messageList[id] = append(o.messageList[id], &request.Message{Role: "assistant", Content: chatResp.Choices[0].Message.Content})
	return chatResp.Choices[0].Message.Content
}
