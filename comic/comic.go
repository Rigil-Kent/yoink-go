package comic

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// var debugUrl = "https://readallcomics.com/ultraman-x-avengers-001-2024/"

type Comic struct {
	URL         string
	Title       string
	Markup      *goquery.Document
	Filelist    []string
	Next        *Comic
	Prev        *Comic
	LibraryPath string
}

func extractTitleFromMarkup(c Comic) string {
	yearFormat := `^(.*?)\s+\(\d{4}(?:\s+.+)?\)`
	selection := c.Markup.Find("title")

	if selection.Length() == 0 {
		return "Untitled"
	}

	content := selection.First().Text()
	regex := regexp.MustCompile(yearFormat)
	matches := regex.FindStringSubmatch(content)

	if len(matches) != 2 {
		return "Untitled"
	}

	return strings.ReplaceAll(matches[1], ":", "")
}

func NewComic(
	url string, libraryPath string,
	imageChannel chan []string,
	markupChannel chan *goquery.Document,
) *Comic {
	c := &Comic{
		URL:         url,
		LibraryPath: libraryPath,
	}

	go Markup(c.URL, markupChannel)

	markup := <-markupChannel
	c.Markup = markup
	c.Title = extractTitleFromMarkup(*c)

	go ParseImageLinks(markup, imageChannel)
	links := <-imageChannel

	c.Filelist = links

	return c
}

func (c *Comic) Cover() (imageFilepath string, err error) {
	for _, image := range c.Filelist {
		if strings.HasSuffix(image, "000.jpg") || strings.HasSuffix(image, "001.jpg") {
			image, err := filepath.Abs(image)
			if err != nil {
				return image, ImageParseError{Message: err.Error(), Code: 1}
			}
			return image, nil
		}
	}
	return "", ImageParseError{Message: "No cover found", Code: 1}
}
