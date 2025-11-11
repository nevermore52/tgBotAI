package tgbot

import (
	"fmt"
	"log"
	"os"
	"tgbot/internal/database"
	"tgbot/internal/AI"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)
type Tgbot struct {
	DB database.Postgres
}

func NewTgBot(db database.Postgres) *Tgbot{
	return &Tgbot{DB:db}
}
func (tg *Tgbot) StartBot() {
	err := godotenv.Load()
	if err != nil {
		fmt.Print("Ошибка .env")
	}
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TGBOT_APIKEY"))
	if err != nil {
		log.Panic(err)
	}
	gemini, err := ai.NewGenaiChat(os.Getenv("DEEPSEEK_APIKEY"))

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	
	updates := bot.GetUpdatesChan(u)


	cmdCfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "/start",
			Description: "Комадна для запуска бота.",
		},
		tgbotapi.BotCommand{
			Command:     "/info",
			Description: "Информация о боте и колличестве запросов.",
		},
	)
	bot.Send(cmdCfg)
	



	for update := range updates {
		if !update.Message.IsCommand() {
			user := tg.DB.CheckUser(update.Message.Chat.ID)
			if user {
				req := tg.DB.CheckRequests(update.Message.Chat.ID)
				if req != 0 {
				geminiResult := gemini.GenaiChatting(update.Message.Text, update.Message.Chat.ID)
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
				user := tg.DB.CheckUser(update.Message.Chat.ID)
				if user {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Приветствую вас, вы можете задать любой вопрос в чат.")
					bot.Send(msg)
				} else {
				tg.DB.AddAccount(database.User{
					Chatid: update.Message.Chat.ID,
					Username: update.Message.Chat.UserName,
					Requests: 5,
					Admin: 0,})
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Приветствую вас, вы можете задать любой вопрос в чат .")
					bot.Send(msg)
				}

			}
		case "info":
			{
				req := tg.DB.CheckRequests(update.Message.Chat.ID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("У вас осталось: %d запросов.", req))
				bot.Send(msg)
			}
		}
	}
}
