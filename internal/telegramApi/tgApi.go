package tgbot

import (
	"log"
	"tgbot/internal/database"
	"tgbot/internal/genai"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
type Tgbot struct {
	DB database.Postgres
}

func NewTgBot(db database.Postgres) *Tgbot{
	return &Tgbot{DB:db}
}
func (tg *Tgbot) StartBot() {

	bot, err := tgbotapi.NewBotAPI("8575220672:AAEoa0kb1VKBt8Oy8gARLBWjVul6lzwNV_c")
	if err != nil {
		log.Panic(err)
	}
	gemini, err := genai.NewGenaiChat()

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	
	updates := bot.GetUpdatesChan(u)

	defer recover()

	for update := range updates {
		if !update.Message.IsCommand() {
			user := tg.DB.CheckUser(update.Message.Chat.ID)
			if user {
				req := tg.DB.CheckRequests(update.Message.Chat.ID)
				if req != 0 {
				geminiResult := gemini.GenaiChatting(update.Message.Text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, geminiResult)
				msg.ReplyToMessageID = update.Message.MessageID
				err = tg.DB.MinusRequest(update.Message.Chat.ID)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас закончились запросы.")
						bot.Send(msg)
				}
				bot.Send(msg)		
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас закончились запросы.")
					bot.Send(msg)
				}	
			}
		}
	
		switch update.Message.Command() {
		case "start":
			{
				err := tg.DB.AddAccount(database.User{
					Chatid: update.Message.Chat.ID,
					Username: update.Message.Chat.UserName,
					Requests: 5,
					Admin: 0,})
				if err != "" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы использовали команду /start, но вы уже зарегистрированы.")
					bot.Send(msg)
				}
			}
		case "settings":
			{
				
			}
		}
	}
}
