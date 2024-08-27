package comic

import (
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Markup retrieves the HTML content from a given URL and returns a goquery Document.
//
// url is the URL to retrieve the HTML content from.
// c is a channel for sending the retrieved goquery Document.
// Returns the retrieved goquery Document.
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

// ParseImageLinks parses a goquery document to extract image links.
//
// markup is the goquery document to parse for image links.
// c is a channel for sending the extracted image links.
// Returns a slice of image links and an error if no images are found.
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
