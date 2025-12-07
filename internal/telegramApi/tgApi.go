package tgbot

import (
	"fmt"
	"log"
	"os"
	"strings"
	ai "tgbot/internal/AI"
	"tgbot/internal/database"
	"tgbot/internal/news"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type Tgbot struct {
	DB          database.Postgres
	NewsManager *news.NewsManager
}

func NewTgBot(db database.Postgres) *Tgbot {
	newsManager := news.NewNewsManager()

	newsManager.AddParser(news.NewRSSParser(
		"https://feeds.bbci.co.uk/russian/rss.xml",
		"BBC News",
	))

	newsManager.AddParser(news.NewRSSParser(
		"https://www.interfax.ru/rss.asp",
		"Интерфакс",
	))

	newsManager.AddParser(news.NewRSSParser(
		"https://ria.ru/export/rss2/index.xml",
		"РИА Новости",
	))

	return &Tgbot{
		DB:          db,
		NewsManager: newsManager,
	}
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
	if err != nil {
		log.Panic(err)
	}

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
				if tg.isNewsRequest(update.Message.Text) {
					loadingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ Загружаю новости...")
					loadingMsgSent, _ := bot.Send(loadingMsg)

					newsItems, err := tg.NewsManager.GetLatestNews(10)
					if err != nil {
						log.Printf("Ошибка получения новостей: %v", err)
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Ошибка при получении новостей. Попробуйте позже.")
						bot.Send(errorMsg)
						deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, loadingMsgSent.MessageID)
						bot.Send(deleteMsg)
						continue
					}

					if len(newsItems) == 0 {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "К сожалению, новости не найдены. Попробуйте позже.")
						bot.Send(errorMsg)
						deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, loadingMsgSent.MessageID)
						bot.Send(deleteMsg)
						continue
					}

					newsMessages := news.FormatNewsForTelegram(newsItems, 10)

					deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, loadingMsgSent.MessageID)
					bot.Send(deleteMsg)

					for i, newsText := range newsMessages {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, newsText)
						msg.ParseMode = "HTML"
						msg.DisableWebPagePreview = false

						_, err = bot.Send(msg)
						if err != nil {
							log.Printf("Ошибка отправки новостей (сообщение %d): %v", i+1, err)
							msg.ParseMode = ""
							_, err2 := bot.Send(msg)
							if err2 != nil {
								log.Printf("Критическая ошибка отправки новостей: %v", err2)
								errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Ошибка при отправке новостей. Попробуйте позже.")
								bot.Send(errorMsg)
							}
						}
					}
					continue
				}

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
						Chatid:   update.Message.Chat.ID,
						Username: update.Message.Chat.UserName,
						Requests: 5,
						Admin:    0})
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

func (tg *Tgbot) isNewsRequest(text string) bool {
	if text == "" {
		return false
	}

	lowerText := strings.ToLower(text)

	newsKeywords := []string{
		"новости",
		"новость",
		"хочу новости",
		"покажи новости",
		"дай новости",
		"пришли новости",
		"что нового",
		"что происходит",
		"актуальные новости",
		"последние новости",
		"свежие новости",
		"мировые новости",
		"новости мира",
		"что в мире",
		"что случилось",
	}

	for _, keyword := range newsKeywords {
		if strings.Contains(lowerText, keyword) {
			return true
		}
	}

	return false
}
