package comic

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ImageParseError struct {
	Message string
	Code    int
}

func (i ImageParseError) Error() string {
	return i.Message
}

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

func BatcaveBizMarkup(referer string, c chan *goquery.Document, clientChan chan *http.Client) *goquery.Document {
	sendErr := func() *goquery.Document {
		if c != nil {
			c <- &goquery.Document{}
		}
		if clientChan != nil {
			clientChan <- nil
		}
		return &goquery.Document{}
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: time.Second * 30,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
	}

	// GET the challange page to obtain cookies and any necessary tokens
	req, err := http.NewRequest("GET", referer, nil)
	if err != nil {
		return sendErr()
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return sendErr()
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return sendErr()
	}

	tokenRegex := regexp.MustCompile(`token:\s*"([^"]+)"`)
	matches := tokenRegex.FindSubmatch(body)

	if matches == nil {
		//  no challenge, parse directly
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return sendErr()
		}
		if c != nil {
			c <- doc
		}
		if clientChan != nil {
			clientChan <- client
		}
		return doc
	}

	encodedToken := string(matches[1])
	token, err := url.QueryUnescape(encodedToken)
	if err != nil {
		token = encodedToken
	}

	// Step 3: POST to /_v with fake browser metrics
	params := url.Values{}
	params.Set("token", token)
	params.Set("mode", "modern")
	params.Set("workTime", "462")
	params.Set("iterations", "183")
	params.Set("webdriver", "0")
	params.Set("touch", "0")
	params.Set("screen_w", "1920")
	params.Set("screen_h", "1080")
	params.Set("screen_cd", "24")

	postReq, err := http.NewRequest("POST", "https://batcave.biz/_v", strings.NewReader(params.Encode()))
	if err != nil {
		return sendErr()
	}
	for k, v := range headers {
		postReq.Header.Set(k, v)
	}
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("Referer", referer)

	postRes, err := client.Do(postReq)
	if err != nil {
		return sendErr()
	}
	defer postRes.Body.Close()
	io.ReadAll(postRes.Body)

	// GET the real page with the set cookie
	realReq, err := http.NewRequest("GET", referer, nil)
	if err != nil {
		return sendErr()
	}
	for k, v := range headers {
		realReq.Header.Set(k, v)
	}

	realRes, err := client.Do(realReq)
	if err != nil {
		return sendErr()
	}
	defer realRes.Body.Close()

	doc, err := goquery.NewDocumentFromReader(realRes.Body)
	if err != nil {
		return sendErr()
	}
	if c != nil {
		c <- doc
	}
	if clientChan != nil {
		clientChan <- client
	}
	return doc
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

func ParseReadAllComicsLinks(markup *goquery.Document, c chan []string) ([]string, error) {
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

// ParseBatcaveBizTitle extracts the chapter title from the __DATA__.chapters array
// by matching the chapter id to the last path segment of the provided URL.
func ParseBatcaveBizTitle(markup *goquery.Document, chapterURL string) string {
	slug := strings.TrimRight(chapterURL, "/")
	if i := strings.LastIndex(slug, "/"); i >= 0 {
		slug = slug[i+1:]
	}

	var title string
	markup.Find("script").Each(func(_ int, s *goquery.Selection) {
		if title != "" {
			return
		}
		text := s.Text()
		if !strings.Contains(text, "__DATA__") {
			return
		}
		chapterRegex := regexp.MustCompile(`"id"\s*:\s*` + regexp.QuoteMeta(slug) + `[^}]*?"title"\s*:\s*"([^"]+)"`)
		m := chapterRegex.FindStringSubmatch(text)
		if len(m) >= 2 {
			title = strings.ReplaceAll(m[1], `\/`, "/")
			title = strings.ReplaceAll(title, "Issue #", "")
			title = strings.ReplaceAll(title, "#", "")
		}
	})
	return title
}

// ParseBatcaveBizImageLinks extracts image URLs from the __DATA__.images JavaScript
// variable embedded in a batcave.biz page.
func ParseBatcaveBizImageLinks(markup *goquery.Document, c chan []string) ([]string, error) {
	var links []string

	markup.Find("script").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		if !strings.Contains(text, "__DATA__") {
			return
		}

		arrayRegex := regexp.MustCompile(`"images"\s*:\s*\[([^\]]+)\]`)
		arrayMatch := arrayRegex.FindStringSubmatch(text)
		if len(arrayMatch) < 2 {
			return
		}

		urlRegex := regexp.MustCompile(`"([^"]+)"`)
		for _, m := range urlRegex.FindAllStringSubmatch(arrayMatch[1], -1) {
			if len(m) >= 2 {
				links = append(links, strings.ReplaceAll(m[1], `\/`, "/"))
			}
		}
	})

	c <- links

	if len(links) > 0 {
		return links, nil
	}

	return links, ImageParseError{Message: "No images found", Code: 1}
}
