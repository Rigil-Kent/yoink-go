package comic

import (
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Markup(url string, c chan *goquery.Document) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		return &goquery.Document{}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return &goquery.Document{}
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return &goquery.Document{}
	}

	markup, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return &goquery.Document{}
	}

	c <- markup
	return markup
}

func ParseImageLinks(markup *goquery.Document, c chan []string) ([]string, error) {
	var links []string
	markup.Find("img").Each(func(_ int, image *goquery.Selection) {
		link, _ := image.Attr("src")
		if !strings.Contains(link, "logo") && (strings.Contains(link, "bp.blogspot.com") || strings.Contains(link, "blogger.googleusercontent") || strings.Contains(link, "covers")) {
			links = append(links, link)
		}
	})

	c <- links

	if len(links) > 0 {
		return links, nil
	}

	return links, ImageParseError{Message: "No images found", Code: 1}
}
