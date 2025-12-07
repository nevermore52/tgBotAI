package news

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type ExampleParser struct {
	*BaseParser
}

func NewExampleParser(sourceURL, source string) *ExampleParser {
	return &ExampleParser{
		BaseParser: NewBaseParser(sourceURL, source),
	}
}

func (p *ExampleParser) Parse() ([]NewsItem, error) {
	htmlContent, err := p.FetchHTML()
	if err != nil {
		return nil, err
	}

	newsItems := []NewsItem{}

	titleRegex := regexp.MustCompile(`<h[1-3][^>]*>([^<]+)</h[1-3]>`)
	titleMatches := titleRegex.FindAllStringSubmatch(htmlContent, -1)

	links := FindLinks(htmlContent)

	for i, match := range titleMatches {
		if len(match) > 1 {
			item := NewsItem{
				Title:       strings.TrimSpace(match[1]),
				Description: "",
				Link:        "",
				PublishedAt: time.Now(),
				Source:      p.Source,
			}

			if i < len(links) && strings.HasPrefix(links[i], "http") {
				item.Link = links[i]
			}

			newsItems = append(newsItems, item)
		}
	}

	return newsItems, nil
}

func (p *ExampleParser) GetSource() string {
	return p.Source
}

func ParseRSS(rssURL string) ([]NewsItem, error) {
	return nil, fmt.Errorf("RSS парсер еще не реализован")
}
