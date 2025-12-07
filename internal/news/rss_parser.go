package news

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type RSSParser struct {
	*BaseParser
}

func NewRSSParser(rssURL, source string) *RSSParser {
	return &RSSParser{
		BaseParser: NewBaseParser(rssURL, source),
	}
}

func (p *RSSParser) Parse() ([]NewsItem, error) {
	xmlContent, err := p.FetchHTML()
	if err != nil {
		log.Printf("Ошибка получения RSS фида %s: %v", p.SourceURL, err)
		return nil, err
	}

	var newsItems []NewsItem

	itemRegex := regexp.MustCompile(`(?s)<item[^>]*>(.*?)</item>`)
	items := itemRegex.FindAllStringSubmatch(xmlContent, -1)

	if len(items) == 0 {
		log.Printf("Не найдено новостей в RSS фиде %s", p.SourceURL)
		itemRegex = regexp.MustCompile(`(?s)<entry[^>]*>(.*?)</entry>`)
		items = itemRegex.FindAllStringSubmatch(xmlContent, -1)
	}

	for _, itemMatch := range items {
		if len(itemMatch) < 2 {
			continue
		}

		itemContent := itemMatch[1]
		newsItem := NewsItem{
			Source:      p.Source,
			PublishedAt: time.Now(),
		}

		titleRegex := regexp.MustCompile(`(?s)<title[^>]*><!\[CDATA\[(.*?)\]\]></title>|<title[^>]*>(.*?)</title>`)
		titleMatch := titleRegex.FindStringSubmatch(itemContent)
		if len(titleMatch) > 1 {
			if titleMatch[1] != "" {
				newsItem.Title = cleanUTF8String(cleanHTML(strings.TrimSpace(titleMatch[1])))
			} else if len(titleMatch) > 2 && titleMatch[2] != "" {
				newsItem.Title = cleanUTF8String(cleanHTML(strings.TrimSpace(titleMatch[2])))
			}
		}

		descRegex := regexp.MustCompile(`(?s)<description[^>]*><!\[CDATA\[(.*?)\]\]></description>|<description[^>]*>(.*?)</description>`)
		descMatch := descRegex.FindStringSubmatch(itemContent)
		if len(descMatch) > 1 {
			if descMatch[1] != "" {
				newsItem.Description = cleanUTF8String(cleanHTML(strings.TrimSpace(descMatch[1])))
			} else if len(descMatch) > 2 && descMatch[2] != "" {
				newsItem.Description = cleanUTF8String(cleanHTML(strings.TrimSpace(descMatch[2])))
			}
		}

		linkRegex := regexp.MustCompile(`<link[^>]*href=["']([^"']+)["'][^>]*/?>|<link[^>]*>(.*?)</link>`)
		linkMatch := linkRegex.FindStringSubmatch(itemContent)
		if len(linkMatch) > 1 {
			if linkMatch[1] != "" {
				newsItem.Link = strings.TrimSpace(linkMatch[1])
			} else if len(linkMatch) > 2 && linkMatch[2] != "" {
				newsItem.Link = strings.TrimSpace(linkMatch[2])
			}
		}

		pubDateRegex := regexp.MustCompile(`<pubDate[^>]*>(.*?)</pubDate>`)
		pubDateMatch := pubDateRegex.FindStringSubmatch(itemContent)
		if len(pubDateMatch) > 1 {
			if t, err := time.Parse(time.RFC1123Z, strings.TrimSpace(pubDateMatch[1])); err == nil {
				newsItem.PublishedAt = t
			} else if t, err := time.Parse(time.RFC1123, strings.TrimSpace(pubDateMatch[1])); err == nil {
				newsItem.PublishedAt = t
			}
		}

		if newsItem.Title != "" {
			newsItems = append(newsItems, newsItem)
		}
	}

	if len(newsItems) == 0 {
		log.Printf("Не удалось извлечь новости из RSS фида %s. Размер контента: %d", p.SourceURL, len(xmlContent))
		return nil, fmt.Errorf("не удалось извлечь новости из RSS фида")
	}

	log.Printf("Успешно извлечено %d новостей из %s", len(newsItems), p.Source)
	return newsItems, nil
}

func cleanHTML(text string) string {
	htmlTagRegex := regexp.MustCompile(`<[^>]+>`)
	cleaned := htmlTagRegex.ReplaceAllString(text, "")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&apos;", "'")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")
	return strings.TrimSpace(cleaned)
}

func cleanUTF8String(s string) string {
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

func (p *RSSParser) GetSource() string {
	return p.Source
}
