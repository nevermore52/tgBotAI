package news

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

type NewsManager struct {
	parsers []Parser
}

func NewNewsManager() *NewsManager {
	return &NewsManager{
		parsers: []Parser{},
	}
}

func (nm *NewsManager) AddParser(parser Parser) {
	nm.parsers = append(nm.parsers, parser)
}

func (nm *NewsManager) GetLatestNews(limit int) ([]NewsItem, error) {
	var allNews []NewsItem
	successCount := 0

	for _, parser := range nm.parsers {
		items, err := parser.Parse()
		if err != nil {
			log.Printf("Ошибка парсинга источника %s: %v", parser.GetSource(), err)
			continue
		}

		if len(items) == 0 {
			log.Printf("Источник %s вернул 0 новостей", parser.GetSource())
			continue
		}

		successCount++
		log.Printf("Успешно получено %d новостей из %s", len(items), parser.GetSource())

		sourceLimit := limit
		if limit > 0 && len(nm.parsers) > 1 {
			sourceLimit = limit / len(nm.parsers)
			if sourceLimit < 3 {
				sourceLimit = 3
			}
		}
		if sourceLimit > 0 && len(items) > sourceLimit {
			items = items[:sourceLimit]
		}

		allNews = append(allNews, items...)
	}

	if len(allNews) == 0 {
		if successCount == 0 {
			return nil, fmt.Errorf("не удалось получить новости ни из одного источника")
		}
	}

	log.Printf("Всего получено %d новостей из %d успешных источников", len(allNews), successCount)
	return allNews, nil
}

func FormatNewsForTelegram(news []NewsItem, limit int) []string {
	if len(news) == 0 {
		return []string{"К сожалению, новости не найдены. Попробуйте позже."}
	}

	if limit > 0 && len(news) > limit {
		news = news[:limit]
	}

	var messages []string
	var currentMessage strings.Builder
	currentMessage.WriteString("📰 <b>Актуальные новости</b>\n\n")

	const maxLength = 4000

	for i, item := range news {
		newsItemText := fmt.Sprintf("<b>%d. %s</b>\n", i+1, escapeHTML(item.Title))

		if item.Description != "" {
			desc := item.Description
			if len(desc) > 150 {
				desc = desc[:150] + "..."
			}
			newsItemText += escapeHTML(desc) + "\n"
		}

		if item.Link != "" {
			newsItemText += fmt.Sprintf("<a href=\"%s\">Читать далее</a>\n", item.Link)
		}

		if item.Source != "" {
			newsItemText += fmt.Sprintf("<i>Источник: %s</i>\n", escapeHTML(item.Source))
		}

		newsItemText += "\n"

		if currentMessage.Len()+len(newsItemText) > maxLength && currentMessage.Len() > 0 {
			messages = append(messages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString("📰 <b>Актуальные новости (продолжение)</b>\n\n")
		}

		currentMessage.WriteString(newsItemText)
	}

	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}

	if len(messages) == 0 {
		return []string{"К сожалению, новости не найдены. Попробуйте позже."}
	}

	return messages
}

func escapeHTML(text string) string {
	text = cleanUTF8(text)
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}

func cleanUTF8(s string) string {
	if !utf8.ValidString(s) {
		var result strings.Builder
		result.Grow(len(s))

		for len(s) > 0 {
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError && size == 1 {
				s = s[1:]
				continue
			}
			result.WriteRune(r)
			s = s[size:]
		}
		return result.String()
	}
	return s
}

func (nm *NewsManager) GetNewsFromSource(sourceName string, limit int) ([]NewsItem, error) {
	for _, parser := range nm.parsers {
		if parser.GetSource() == sourceName {
			items, err := parser.Parse()
			if err != nil {
				return nil, err
			}

			if limit > 0 && len(items) > limit {
				items = items[:limit]
			}

			return items, nil
		}
	}

	return nil, fmt.Errorf("источник '%s' не найден", sourceName)
}
