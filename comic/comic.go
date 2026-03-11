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
	regex := regexp.MustCompile(yearFormat)

	extractFrom := func(text string) string {
		matches := regex.FindStringSubmatch(text)
		if len(matches) != 2 {
			return ""
		}
		return strings.ReplaceAll(matches[1], ":", "")
	}

	title := extractFrom(c.Markup.Find("title").First().Text())

	if strings.HasPrefix(title, "#") {
		if h1 := extractFrom(c.Markup.Find("h1").First().Text()); h1 != "" && !strings.HasPrefix(h1, "#") {
			return h1
		}
		if slug := titleFromSlug(c.URL); slug != "" {
			return slug
		}
	}

	if title != "" {
		return title
	}

	return "Untitled"
}

// titleFromSlug derives a comic title from the last path segment of a URL.
// It strips a trailing year (-YYYY), replaces hyphens with spaces, and title-cases the result.
func titleFromSlug(url string) string {
	slug := strings.TrimRight(url, "/")
	if i := strings.LastIndex(slug, "/"); i >= 0 {
		slug = slug[i+1:]
	}
	slug = regexp.MustCompile(`-\d{4}$`).ReplaceAllString(slug, "")
	if slug == "" {
		return ""
	}
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
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

	if strings.Contains(url, "batcave.biz") {
		go BatcaveBizMarkup(url, markupChannel)
	} else {
		go Markup(url, markupChannel)
	}

	markup := <-markupChannel
	c.Markup = markup
	c.Title = extractTitleFromMarkup(*c)

	if strings.Contains(url, "batcave.biz") {
		go ParseBatcaveBizImageLinks(markup, imageChannel)
	} else {
		go ParseImageLinks(markup, imageChannel)
	}
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
