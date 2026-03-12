package comic

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
)

type ComicDownloadError struct {
	Message string
	Code    int
}

func (c ComicDownloadError) Error() string {
	return c.Message
}

// downloadFile downloads a file from a given URL and saves it to a specified location.
//
// The function takes a URL string, a page number, and a Comic struct as parameters.
// It returns an error if the download fails, and nil otherwise.
func downloadFile(url string, page int, c *Comic) error {
	pageNumber := fmt.Sprintf("%03d", page)
	formattedImagePath := fmt.Sprintf("%s %s.jpg", c.Title, pageNumber)
	imageFilepath, _ := filepath.Abs(filepath.Join(c.LibraryPath, c.Title, formattedImagePath))

	if err := os.MkdirAll(
		filepath.Dir(imageFilepath),
		os.ModePerm,
	); err != nil {
		return ComicDownloadError{
			Message: "error creating directory",
			Code:    1,
		}
	}

	var res *http.Response
	var err error
	if c.Client != nil {
		req, reqErr := http.NewRequest("GET", url, nil)
		if reqErr != nil {
			return ComicDownloadError{Message: "invalid request", Code: 1}
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		if strings.Contains(url, "batcave.biz") {
			req.Header.Set("Referer", "https://batcave.biz/")
		}
		res, err = c.Client.Do(req)
	} else {
		res, err = handleRequest(url)
	}
	if err != nil {
		return ComicDownloadError{
			Message: "invalid request",
			Code:    1,
		}
	}
	if res.StatusCode != http.StatusOK {
		return ComicDownloadError{
			Message: "bad response",
			Code:    1,
		}
	}
	defer res.Body.Close()

	imageFile, err := os.Create(imageFilepath)
	if err != nil {
		return ComicDownloadError{
			Message: "error creating image file",
			Code:    1,
		}
	}
	defer imageFile.Close()

	written, err := io.Copy(imageFile, res.Body)
	if err != nil {
		return ComicDownloadError{
			Message: "Unable to save file contents",
			Code:    1,
		}
	}

	if written == 0 {
		return ComicDownloadError{
			Message: "Unable to save file contents",
			Code:    1,
		}
	}

	return nil
}

// handleRequest sends a GET request to the provided URL, mimicking a generic browser,
// and returns the HTTP response.
//
// url - the URL to send the request to.
// *http.Response - the HTTP response from the server.
// error - an error that occurred during the request.
func handleRequest(url string) (*http.Response, error) {
	// adjust timeout and keep-alive to avoid connection timeout
	transport := &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 32,
	}

	// add cloudflare bypass
	cfTransport := cloudflarebp.AddCloudFlareByPass(transport)

	// prevents cloudflarebp from occasionally returning the wrong type
	if converted, ok := cfTransport.(*http.Transport); ok {
		transport = converted
	}

	client := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}

	// mimic generic browser
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ComicDownloadError{
			Message: "invalid request",
			Code:    1,
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, ComicDownloadError{
			Message: "invalid request",
			Code:    1,
		}
	}

	if res.StatusCode != http.StatusOK {
		return nil, ComicDownloadError{
			Message: "bad response",
			Code:    1,
		}
	}

	return res, nil
}

// Download is a method of the Comic struct that downloads multiple files concurrently.
//
// It takes an integer parameter `concurrency` which represents the number of concurrent downloads.
//
// It returns a slice of errors, each representing an error that occurred during the download process.
func (c *Comic) Download(concurrency int) []error {
	jobs := make(chan Download)
	results := make(chan error)

	for worker := 1; worker <= concurrency; worker++ {
		go workerPool(jobs, results)
	}

	for i, url := range c.Filelist {
		jobs <- Download{
			URL:   url,
			Page:  i + 1,
			Comic: c,
		}
	}

	var errors []error
	for i := 0; i < len(c.Filelist); i++ {
		err := <-results
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

type Download struct {
	URL   string
	Page  int
	Comic *Comic
}

// workerPool is a function that processes a channel of Download jobs concurrently.
//
// It takes two parameters: a receive-only channel of Download jobs and a send-only channel of errors.
// It returns no value, but sends errors to the results channel as they occur.
func workerPool(jobs <-chan Download, results chan<- error) {
	for job := range jobs {
		results <- downloadFile(job.URL, job.Page, job.Comic)
	}
}
