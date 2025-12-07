package news

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type NewsItem struct {
	Title       string
	Description string
	Link        string
	PublishedAt time.Time
	Source      string
}

type Parser interface {
	Parse() ([]NewsItem, error)
	GetSource() string
}

type BaseParser struct {
	SourceURL string
	Source    string
	Client    *http.Client
}

func NewBaseParser(sourceURL, source string) *BaseParser {
	return &BaseParser{
		SourceURL: sourceURL,
		Source:    source,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *BaseParser) FetchHTML() (string, error) {
	resp, err := p.Client.Get(p.SourceURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении страницы: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("неверный статус код: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %w", err)
	}

	return string(body), nil
}

func ExtractTextFromHTML(htmlContent string) string {

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var text strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(strings.TrimSpace(n.Data) + " ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(doc)
	return strings.TrimSpace(text.String())
}

func FindLinks(htmlContent string) []string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var links []string
	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c)
		}
	}
	findLinks(doc)
	return links
}
