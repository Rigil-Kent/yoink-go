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

// extractTitleFromMarkup extracts the title from the comic's markup.
//
// c is the Comic instance containing the markup to extract the title from.
// Returns the extracted title as a string.
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

// NewComic creates a new Comic instance from the provided URL and library path.
//
// url is the URL of the comic to be parsed.
// libraryPath is the path to the local library where the comic will be stored.
// imageChannel is a channel for receiving image links.
// markupChannel is a channel for receiving the comic's markup.
//
// Returns a pointer to the newly created Comic instance.
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

// Cover returns the absolute filepath of the cover image of the comic.
//
// It iterates through the list of images associated with the comic and returns the first image that ends with "000.jpg" or "001.jpg".
// If no such image is found, it returns an error.
// Returns the absolute filepath of the cover image and an error.
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
